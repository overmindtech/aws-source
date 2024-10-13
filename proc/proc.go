package proc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	awsapigateway "github.com/aws/aws-sdk-go-v2/service/apigateway"
	awsautoscaling "github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awscloudfront "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awscloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awsdirectconnect "github.com/aws/aws-sdk-go-v2/service/directconnect"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	awsefs "github.com/aws/aws-sdk-go-v2/service/efs"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awselasticloadbalancing "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awselasticloadbalancingv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awskms "github.com/aws/aws-sdk-go-v2/service/kms"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	awsnetworkfirewall "github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awsnetworkmanager "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go/auth"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// This package contains a few functions needed by the CLI to load this in-proc.
// These can not go into `/sources` because that would cause an import cycle
// with everything else.

type AwsAuthConfig struct {
	Strategy        string
	AccessKeyID     string
	SecretAccessKey string
	ExternalID      string
	TargetRoleARN   string
	Profile         string
	AutoConfig      bool

	Regions []string
}

func (c AwsAuthConfig) GetAWSConfig(region string) (aws.Config, error) {
	// Validate inputs
	if region == "" {
		return aws.Config{}, errors.New("aws-region cannot be blank")
	}

	ctx := context.Background()

	options := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithAppID("Overmind"),
	}

	if c.AutoConfig {
		if c.Strategy != "defaults" {
			log.WithField("aws-access-strategy", c.Strategy).Warn("auto-config is set to true, but aws-access-strategy is not set to 'defaults'. This may cause unexpected behaviour")
		}
		return config.LoadDefaultConfig(ctx, options...)
	}

	if c.Strategy == "defaults" {
		return config.LoadDefaultConfig(ctx, options...)
	} else if c.Strategy == "access-key" {
		if c.AccessKeyID == "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-access-key-id cannot be blank")
		}
		if c.SecretAccessKey == "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-secret-access-key cannot be blank")
		}
		if c.ExternalID != "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-external-id must be blank")
		}
		if c.TargetRoleARN != "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-target-role-arn must be blank")
		}
		if c.Profile != "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-profile must be blank")
		}

		options = append(options, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""),
		))

		return config.LoadDefaultConfig(ctx, options...)
	} else if c.Strategy == "external-id" {
		if c.AccessKeyID != "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-access-key-id must be blank")
		}
		if c.SecretAccessKey != "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-secret-access-key must be blank")
		}
		if c.ExternalID == "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-external-id cannot be blank")
		}
		if c.TargetRoleARN == "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-target-role-arn cannot be blank")
		}
		if c.Profile != "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-profile must be blank")
		}

		assumeConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return aws.Config{}, fmt.Errorf("could not load default config from environment: %w", err)
		}

		options = append(options, config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscredsv2.NewAssumeRoleProvider(
				sts.NewFromConfig(assumeConfig),
				c.TargetRoleARN,
				func(aro *stscredsv2.AssumeRoleOptions) {
					aro.ExternalID = &c.ExternalID
				},
			)),
		))

		return config.LoadDefaultConfig(ctx, options...)
	} else if c.Strategy == "sso-profile" {
		if c.AccessKeyID != "" {
			return aws.Config{}, errors.New("with sso-profile strategy, aws-access-key-id must be blank")
		}
		if c.SecretAccessKey != "" {
			return aws.Config{}, errors.New("with sso-profile strategy, aws-secret-access-key must be blank")
		}
		if c.ExternalID != "" {
			return aws.Config{}, errors.New("with sso-profile strategy, aws-external-id must be blank")
		}
		if c.TargetRoleARN != "" {
			return aws.Config{}, errors.New("with sso-profile strategy, aws-target-role-arn must be blank")
		}
		if c.Profile == "" {
			return aws.Config{}, errors.New("with sso-profile strategy, aws-profile cannot be blank")
		}

		options = append(options, config.WithSharedConfigProfile(c.Profile))

		return config.LoadDefaultConfig(ctx, options...)
	} else {
		return aws.Config{}, errors.New("invalid aws-access-strategy")
	}
}

// Takes AwsAuthConfig options and converts these into a slice of AWS configs,
// one for each region. These can then be passed to
// `InitializeAwsSourceEngine()“ to actually start the source
func CreateAWSConfigs(awsAuthConfig AwsAuthConfig) ([]aws.Config, error) {
	if len(awsAuthConfig.Regions) == 0 {
		return nil, errors.New("no regions specified")
	}

	configs := make([]aws.Config, 0, len(awsAuthConfig.Regions))

	for _, region := range awsAuthConfig.Regions {
		region = strings.Trim(region, " ")

		cfg, err := awsAuthConfig.GetAWSConfig(region)
		if err != nil {
			return nil, fmt.Errorf("error getting AWS config for region %v: %w", region, err)
		}

		// Add OTel instrumentation
		cfg.HTTPClient = &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}

// InitializeAwsSourceEngine initializes an Engine with AWS sources, returns the
// engine, and an error if any. The context provided will be used for the rate
// limit buckets and should not be cancelled until the source is shut down. AWS
// configs should be provided for each region that is enabled
func InitializeAwsSourceEngine(ctx context.Context, name string, version string, engineUUID uuid.UUID, natsOptions auth.NATSOptions, heartbeatOptions *discovery.HeartbeatOptions, maxParallel int, maxRetries uint64, configs ...aws.Config) (*discovery.Engine, error) {
	e, err := discovery.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("error initializing Engine: %w", err)
	}

	var startupErrorMutex sync.Mutex
	startupError := errors.New("source is starting")
	if heartbeatOptions != nil {
		heartbeatOptions.HealthCheck = func() error {
			startupErrorMutex.Lock()
			defer startupErrorMutex.Unlock()
			return startupError
		}
		e.HeartbeatOptions = heartbeatOptions
	}

	e.Name = "aws-source"
	e.NATSOptions = &natsOptions
	e.MaxParallelExecutions = maxParallel
	e.Version = version
	e.Name = name
	e.UUID = engineUUID
	e.Type = "aws"

	e.StartSendingHeartbeats(ctx)

	if len(configs) == 0 {
		return nil, errors.New("No configs specified")
	}

	var globalDone atomic.Bool

	var b backoff.BackOff
	b = backoff.NewExponentialBackOff(
		backoff.WithMaxInterval(30*time.Second),
		backoff.WithMaxElapsedTime(15*time.Minute),
	)
	b = backoff.WithMaxRetries(b, maxRetries)
	tick := backoff.NewTicker(b)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case _, ok := <-tick.C:
			if !ok {
				// If the backoff stops, then we should stop trying to
				// initialize and just return the error
				return nil, err
			}

			p := pool.New().WithContext(ctx)

			for _, cfg := range configs {
				p.Go(func(ctx context.Context) error {
					configCtx, configCancel := context.WithTimeout(ctx, 10*time.Second)
					defer configCancel()

					// Work out what account we're using. This will be used in item scopes
					stsClient := sts.NewFromConfig(cfg)

					callerID, err := stsClient.GetCallerIdentity(configCtx, &sts.GetCallerIdentityInput{})
					if err != nil {
						lf := log.Fields{
							"region": cfg.Region,
						}
						log.WithError(err).WithFields(lf).Error("Error retrieving account information")
						return fmt.Errorf("error getting caller identity for region %v: %w", cfg.Region, err)
					}

					// Create shared clients for each API
					autoscalingClient := awsautoscaling.NewFromConfig(cfg, func(o *awsautoscaling.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					cloudfrontClient := awscloudfront.NewFromConfig(cfg, func(o *awscloudfront.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					cloudwatchClient := awscloudwatch.NewFromConfig(cfg, func(o *awscloudwatch.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					directconnectClient := awsdirectconnect.NewFromConfig(cfg, func(o *awsdirectconnect.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					dynamodbClient := awsdynamodb.NewFromConfig(cfg, func(o *awsdynamodb.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					ec2Client := awsec2.NewFromConfig(cfg, func(o *awsec2.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					ecsClient := awsecs.NewFromConfig(cfg, func(o *awsecs.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					efsClient := awsefs.NewFromConfig(cfg, func(o *awsefs.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					eksClient := awseks.NewFromConfig(cfg, func(o *awseks.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					elbClient := awselasticloadbalancing.NewFromConfig(cfg, func(o *awselasticloadbalancing.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					elbv2Client := awselasticloadbalancingv2.NewFromConfig(cfg, func(o *awselasticloadbalancingv2.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					lambdaClient := awslambda.NewFromConfig(cfg, func(o *awslambda.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					networkfirewallClient := awsnetworkfirewall.NewFromConfig(cfg, func(o *awsnetworkfirewall.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					rdsClient := awsrds.NewFromConfig(cfg, func(o *awsrds.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					snsClient := awssns.NewFromConfig(cfg, func(o *awssns.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					sqsClient := awssqs.NewFromConfig(cfg, func(o *awssqs.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					route53Client := awsroute53.NewFromConfig(cfg, func(o *awsroute53.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					networkmanagerClient := awsnetworkmanager.NewFromConfig(cfg, func(o *awsnetworkmanager.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					iamClient := awsiam.NewFromConfig(cfg, func(o *awsiam.Options) {
						o.RetryMode = aws.RetryModeAdaptive
						// Increase this from the default of 3 since IAM as such low rate limits
						o.RetryMaxAttempts = 5
					})
					kmsClient := awskms.NewFromConfig(cfg, func(o *awskms.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})
					apigatewayClient := awsapigateway.NewFromConfig(cfg, func(o *awsapigateway.Options) {
						o.RetryMode = aws.RetryModeAdaptive
					})

					configuredAdapters := []discovery.Adapter{
						// EC2
						adapters.NewAddressAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewCapacityReservationFleetAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewCapacityReservationAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewEgressOnlyInternetGatewayAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewIamInstanceProfileAssociationAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewImageAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewInstanceEventWindowAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewInstanceAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewInstanceStatusAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewInternetGatewayAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewKeyPairAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewLaunchTemplateAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewLaunchTemplateVersionAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewNatGatewayAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewNetworkAclAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewNetworkInterfacePermissionAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewNetworkInterfaceAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewPlacementGroupAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewReservedInstanceAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewRouteTableAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewSecurityGroupRuleAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewSecurityGroupAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewSnapshotAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewSubnetAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewVolumeAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewVolumeStatusAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewVpcEndpointAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewVpcPeeringConnectionAdapter(ec2Client, *callerID.Account, cfg.Region),
						adapters.NewVpcAdapter(ec2Client, *callerID.Account, cfg.Region),

						// EFS (I'm assuming it shares its rate limit with EC2))
						adapters.NewAccessPointAdapter(efsClient, *callerID.Account, cfg.Region),
						adapters.NewBackupPolicyAdapter(efsClient, *callerID.Account, cfg.Region),
						adapters.NewFileSystemAdapter(efsClient, *callerID.Account, cfg.Region),
						adapters.NewMountTargetAdapter(efsClient, *callerID.Account, cfg.Region),
						adapters.NewReplicationConfigurationAdapter(efsClient, *callerID.Account, cfg.Region),

						// EKS
						adapters.NewAddonAdapter(eksClient, *callerID.Account, cfg.Region),
						adapters.NewClusterAdapter(eksClient, *callerID.Account, cfg.Region),
						adapters.NewFargateProfileAdapter(eksClient, *callerID.Account, cfg.Region),
						adapters.NewNodegroupAdapter(eksClient, *callerID.Account, cfg.Region),

						// Route 53
						adapters.NewHealthCheckAdapter(route53Client, *callerID.Account, cfg.Region),
						adapters.NewHostedZoneAdapter(route53Client, *callerID.Account, cfg.Region),
						adapters.NewResourceRecordSetAdapter(route53Client, *callerID.Account, cfg.Region),

						// Cloudwatch
						adapters.NewAlarmAdapter(cloudwatchClient, *callerID.Account, cfg.Region),

						// IAM
						adapters.NewGroupAdapter(iamClient, *callerID.Account, cfg.Region),
						adapters.NewInstanceProfileAdapter(iamClient, *callerID.Account, cfg.Region),
						adapters.NewPolicyAdapter(iamClient, *callerID.Account, cfg.Region),
						adapters.NewRoleAdapter(iamClient, *callerID.Account, cfg.Region),
						adapters.NewUserAdapter(iamClient, *callerID.Account, cfg.Region),

						// Lambda
						adapters.NewFunctionAdapter(lambdaClient, *callerID.Account, cfg.Region),
						adapters.NewLayerAdapter(lambdaClient, *callerID.Account, cfg.Region),
						adapters.NewLayerVersionAdapter(lambdaClient, *callerID.Account, cfg.Region),

						// ECS
						adapters.NewCapacityProviderAdapter(ecsClient, *callerID.Account, cfg.Region),
						adapters.NewECSClusterAdapter(ecsClient, *callerID.Account, cfg.Region),
						adapters.NewContainerInstanceAdapter(ecsClient, *callerID.Account, cfg.Region),
						adapters.NewServiceAdapter(ecsClient, *callerID.Account, cfg.Region),
						adapters.NewTaskDefinitionAdapter(ecsClient, *callerID.Account, cfg.Region),
						adapters.NewTaskAdapter(ecsClient, *callerID.Account, cfg.Region),

						// DynamoDB
						adapters.NewBackupAdapter(dynamodbClient, *callerID.Account, cfg.Region),
						adapters.NewTableAdapter(dynamodbClient, *callerID.Account, cfg.Region),

						// RDS
						adapters.NewDBClusterParameterGroupAdapter(rdsClient, *callerID.Account, cfg.Region),
						adapters.NewDBClusterAdapter(rdsClient, *callerID.Account, cfg.Region),
						adapters.NewDBInstanceAdapter(rdsClient, *callerID.Account, cfg.Region),
						adapters.NewDBParameterGroupAdapter(rdsClient, *callerID.Account, cfg.Region),
						adapters.NewDBSubnetGroupAdapter(rdsClient, *callerID.Account, cfg.Region),
						adapters.NewOptionGroupAdapter(rdsClient, *callerID.Account, cfg.Region),

						// Autoscaling
						adapters.NewAutoScalingGroupAdapter(autoscalingClient, *callerID.Account, cfg.Region),

						// ELB
						adapters.NewInstanceHealthAdapter(elbClient, *callerID.Account, cfg.Region),
						adapters.NewELBLoadBalancerAdapter(elbClient, *callerID.Account, cfg.Region),

						// ELBv2
						adapters.NewListenerAdapter(elbv2Client, *callerID.Account, cfg.Region),
						adapters.NewLoadBalancerAdapter(elbv2Client, *callerID.Account, cfg.Region),
						adapters.NewRuleAdapter(elbv2Client, *callerID.Account, cfg.Region),
						adapters.NewTargetGroupAdapter(elbv2Client, *callerID.Account, cfg.Region),
						adapters.NewTargetHealthAdapter(elbv2Client, *callerID.Account, cfg.Region),

						// Network Firewall
						adapters.NewFirewallAdapter(networkfirewallClient, *callerID.Account, cfg.Region),
						adapters.NewFirewallPolicyAdapter(networkfirewallClient, *callerID.Account, cfg.Region),
						adapters.NewRuleGroupAdapter(networkfirewallClient, *callerID.Account, cfg.Region),
						adapters.NewTLSInspectionConfigurationAdapter(networkfirewallClient, *callerID.Account, cfg.Region),

						// Direct Connect
						adapters.NewDirectConnectGatewayAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewDirectConnectGatewayAssociationAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewDirectConnectGatewayAssociationProposalAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewDirectconnectConnectionAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewDirectConnectGatewayAttachmentAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewVirtualInterfaceAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewVirtualGatewayAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewCustomerMetadataAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewLagAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewLocationAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewHostedConnectionAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewInterconnectAdapter(directconnectClient, *callerID.Account, cfg.Region),
						adapters.NewRouterConfigurationAdapter(directconnectClient, *callerID.Account, cfg.Region),

						// Network Manager
						adapters.NewConnectAttachmentAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewConnectPeerAssociationAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewConnectPeerAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewCoreNetworkPolicyAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewCoreNetworkAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewNetworkResourceRelationshipsAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewSiteToSiteVpnAttachmentAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewTransitGatewayConnectPeerAssociationAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewTransitGatewayPeeringAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewTransitGatewayRegistrationAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewTransitGatewayRouteTableAttachmentAdapter(networkmanagerClient, *callerID.Account, cfg.Region),
						adapters.NewVPCAttachmentAdapter(networkmanagerClient, *callerID.Account, cfg.Region),

						// SQS
						adapters.NewQueueAdapter(sqsClient, *callerID.Account, cfg.Region),

						// SNS
						adapters.NewSubscriptionAdapter(snsClient, *callerID.Account, cfg.Region),
						adapters.NewTopicAdapter(snsClient, *callerID.Account, cfg.Region),
						adapters.NewPlatformApplicationAdapter(snsClient, *callerID.Account, cfg.Region),
						adapters.NewEndpointAdapter(snsClient, *callerID.Account, cfg.Region),
						adapters.NewDataProtectionPolicyAdapter(snsClient, *callerID.Account, cfg.Region),

						// KMS
						adapters.NewKeyAdapter(kmsClient, *callerID.Account, cfg.Region),
						adapters.NewCustomKeyStoreAdapter(kmsClient, *callerID.Account, cfg.Region),
						adapters.NewAliasAdapter(kmsClient, *callerID.Account, cfg.Region),
						adapters.NewGrantAdapter(kmsClient, *callerID.Account, cfg.Region),
						adapters.NewKeyPolicyAdapter(kmsClient, *callerID.Account, cfg.Region),

						// ApiGateway
						adapters.NewRestApiAdapter(apigatewayClient, *callerID.Account, cfg.Region),
						adapters.NewResourceAdapter(apigatewayClient, *callerID.Account, cfg.Region),
					}

					e.AddAdapters(configuredAdapters...)

					// Add "global" sources (those that aren't tied to a region, like
					// cloudfront). but only do this once for the first region. For
					// these APIs it doesn't matter which region we call them from, we
					// get global results
					if globalDone.CompareAndSwap(false, true) {
						e.AddAdapters(
							// Cloudfront
							adapters.NewCachePolicyAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewContinuousDeploymentPolicyAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewDistributionAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewCloudfrontFunctionAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewKeyGroupAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewOriginAccessControlAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewOriginRequestPolicyAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewResponseHeadersPolicyAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewRealtimeLogConfigsAdapter(cloudfrontClient, *callerID.Account),
							adapters.NewStreamingDistributionAdapter(cloudfrontClient, *callerID.Account),

							// S3
							adapters.NewS3Adapter(cfg, *callerID.Account),

							// Networkmanager
							adapters.NewGlobalNetworkAdapter(networkmanagerClient, *callerID.Account),
							adapters.NewSiteAdapter(networkmanagerClient, *callerID.Account),
							adapters.NewLinkAdapter(networkmanagerClient, *callerID.Account),
							adapters.NewDeviceAdapter(networkmanagerClient, *callerID.Account),
							adapters.NewLinkAssociationAdapter(networkmanagerClient, *callerID.Account),
							adapters.NewConnectionAdapter(networkmanagerClient, *callerID.Account),
						)
					}
					return nil
				})
			}

			err = p.Wait()
			startupErrorMutex.Lock()
			startupError = err
			startupErrorMutex.Unlock()
			brokenHeart := e.SendHeartbeat(ctx) // Send the error immediately
			if brokenHeart != nil {
				log.WithError(brokenHeart).Error("Error sending heartbeat")
			}

			if err != nil {
				log.WithError(err).Debug("Error initializing sources")
			} else {
				log.Debug("Sources initialized")
				// If there is no error then return the engine
				return e, nil
			}
		}
	}
}
