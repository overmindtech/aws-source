package proc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	awsnetworkfirewall "github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awsnetworkmanager "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/autoscaling"
	"github.com/overmindtech/aws-source/sources/cloudfront"
	"github.com/overmindtech/aws-source/sources/cloudwatch"
	"github.com/overmindtech/aws-source/sources/directconnect"
	"github.com/overmindtech/aws-source/sources/dynamodb"
	"github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/aws-source/sources/ecs"
	"github.com/overmindtech/aws-source/sources/efs"
	"github.com/overmindtech/aws-source/sources/eks"
	"github.com/overmindtech/aws-source/sources/elb"
	"github.com/overmindtech/aws-source/sources/elbv2"
	"github.com/overmindtech/aws-source/sources/iam"
	"github.com/overmindtech/aws-source/sources/lambda"
	"github.com/overmindtech/aws-source/sources/networkfirewall"
	"github.com/overmindtech/aws-source/sources/networkmanager"
	"github.com/overmindtech/aws-source/sources/rds"
	"github.com/overmindtech/aws-source/sources/route53"
	"github.com/overmindtech/aws-source/sources/s3"
	"github.com/overmindtech/aws-source/sources/sns"
	"github.com/overmindtech/aws-source/sources/sqs"
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

		assumecnf, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return aws.Config{}, fmt.Errorf("could not load default config from environment: %v", err)
		}

		options = append(options, config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscredsv2.NewAssumeRoleProvider(
				sts.NewFromConfig(assumecnf),
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

// InitializeAwsSourceEngine initializes an Engine with AWS sources, returns the
// engine, and an error if any. The xontext provided will be used for the rate
// limit buckets and should not be cancelled until the source is shut down
func InitializeAwsSourceEngine(ctx context.Context, natsOptions auth.NATSOptions, awsAuthConfig AwsAuthConfig, maxParallel int) (*discovery.Engine, error) {
	e, err := discovery.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("error initializing Engine: %w", err)
	}

	e.Name = "aws-source"
	e.NATSOptions = &natsOptions
	e.MaxParallelExecutions = maxParallel

	if len(awsAuthConfig.Regions) == 0 {
		log.Fatal("No regions specified")
	}

	var globalDone bool

	for _, region := range awsAuthConfig.Regions {
		region = strings.Trim(region, " ")

		configCtx, configCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer configCancel()

		cfg, err := awsAuthConfig.GetAWSConfig(region)
		if err != nil {
			configCancel()
			return nil, fmt.Errorf("error getting AWS config for region %v: %w", region, err)
		}

		if log.GetLevel() == log.TraceLevel {
			// Add OTel instrumentation
			cfg.HTTPClient = &http.Client{
				Transport: otelhttp.NewTransport(http.DefaultTransport),
			}
		}

		// Work out what account we're using. This will be used in item scopes
		stsClient := sts.NewFromConfig(cfg)

		callerID, err := stsClient.GetCallerIdentity(configCtx, &sts.GetCallerIdentityInput{})
		if err != nil {
			lf := log.Fields{
				"region":   region,
				"strategy": awsAuthConfig.Strategy,
			}
			if awsAuthConfig.TargetRoleARN != "" {
				lf["targetRoleARN"] = awsAuthConfig.TargetRoleARN
				lf["externalID"] = awsAuthConfig.ExternalID
			}
			log.WithError(err).WithFields(lf).Fatal("Error retrieving account information")
		}

		// Create an EC2 rate limit which limits the source to 50% of the
		// overall rate limit
		ec2RateLimit := sources.LimitBucket{
			MaxCapacity: 50,
			RefillRate:  10,
		}

		// Apparently Autoscaling has a separate bucket to EC2 but I'm going
		// to assume the values are the same, the documentation for rate
		// limiting for everything other than EC2 is very poor
		autoScalingRateLimit := sources.LimitBucket{
			MaxCapacity: 50,
			RefillRate:  10,
		}

		// IAM's rate limit is 20 per second, so we'll use 50% of that at
		// maximum. See:
		// https://docs.aws.amazon.com/singlesignon/latest/userguide/limits.html
		iamRateLimit := sources.LimitBucket{
			MaxCapacity: 10,
			RefillRate:  10,
		}

		directConnectRateLimit := sources.LimitBucket{
			// Use EC2 limits as it's not documented
			MaxCapacity: 50,
			RefillRate:  10,
		}

		networkManagerRateLimit := sources.LimitBucket{
			// Use EC2 limits as it's not documented
			MaxCapacity: 50,
			RefillRate:  10,
		}

		ec2RateLimit.Start(ctx)
		autoScalingRateLimit.Start(ctx)
		iamRateLimit.Start(ctx)
		directConnectRateLimit.Start(ctx)
		networkManagerRateLimit.Start(ctx)

		// Create shared clients for each API
		autoscalingClient := awsautoscaling.NewFromConfig(cfg)
		cloudfrontClient := awscloudfront.NewFromConfig(cfg)
		cloudwatchClient := awscloudwatch.NewFromConfig(cfg)
		directconnectClient := awsdirectconnect.NewFromConfig(cfg)
		dynamodbClient := awsdynamodb.NewFromConfig(cfg)
		ec2Client := awsec2.NewFromConfig(cfg)
		ecsClient := awsecs.NewFromConfig(cfg)
		efsClient := awsefs.NewFromConfig(cfg)
		eksClient := awseks.NewFromConfig(cfg)
		elbClient := awselasticloadbalancing.NewFromConfig(cfg)
		elbv2Client := awselasticloadbalancingv2.NewFromConfig(cfg)
		lambdaClient := awslambda.NewFromConfig(cfg)
		networkfirewallClient := awsnetworkfirewall.NewFromConfig(cfg)
		rdsClient := awsrds.NewFromConfig(cfg)
		snsClient := awssns.NewFromConfig(cfg)
		sqsClient := awssqs.NewFromConfig(cfg)
		route53Client := awsroute53.NewFromConfig(cfg)
		networkmanagerClient := awsnetworkmanager.NewFromConfig(cfg)

		sources := []discovery.Source{
			// EC2
			ec2.NewAddressSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewCapacityReservationFleetSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewCapacityReservationSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewEgressOnlyInternetGatewaySource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewIamInstanceProfileAssociationSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewImageSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewInstanceEventWindowSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewInstanceSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewInstanceStatusSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewInternetGatewaySource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewKeyPairSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewLaunchTemplateSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewLaunchTemplateVersionSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewNatGatewaySource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewNetworkAclSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewNetworkInterfacePermissionSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewNetworkInterfaceSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewPlacementGroupSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewReservedInstanceSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewRouteTableSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewSecurityGroupSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewSnapshotSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewSubnetSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewVolumeSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewVolumeStatusSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewVpcPeeringConnectionSource(ec2Client, *callerID.Account, region, &ec2RateLimit),
			ec2.NewVpcSource(ec2Client, *callerID.Account, region, &ec2RateLimit),

			// EFS (I'm assuming it shares its rate limit with EC2))
			efs.NewAccessPointSource(efsClient, *callerID.Account, region, &ec2RateLimit),
			efs.NewBackupPolicySource(efsClient, *callerID.Account, region, &ec2RateLimit),
			efs.NewFileSystemSource(efsClient, *callerID.Account, region, &ec2RateLimit),
			efs.NewMountTargetSource(efsClient, *callerID.Account, region, &ec2RateLimit),
			efs.NewReplicationConfigurationSource(efsClient, *callerID.Account, region, &ec2RateLimit),

			// EKS
			eks.NewAddonSource(eksClient, *callerID.Account, region),
			eks.NewClusterSource(eksClient, *callerID.Account, region),
			eks.NewFargateProfileSource(eksClient, *callerID.Account, region),
			eks.NewNodegroupSource(eksClient, *callerID.Account, region),

			// Route 53
			route53.NewHealthCheckSource(route53Client, *callerID.Account, region),
			route53.NewHostedZoneSource(route53Client, *callerID.Account, region),
			route53.NewResourceRecordSetSource(route53Client, *callerID.Account, region),

			// Cloudwatch
			cloudwatch.NewAlarmSource(cloudwatchClient, *callerID.Account, region),

			// IAM
			iam.NewGroupSource(cfg, *callerID.Account, region, &iamRateLimit),
			iam.NewInstanceProfileSource(cfg, *callerID.Account, region, &iamRateLimit),
			iam.NewPolicySource(cfg, *callerID.Account, region, &iamRateLimit),
			iam.NewRoleSource(cfg, *callerID.Account, region, &iamRateLimit),
			iam.NewUserSource(cfg, *callerID.Account, region, &iamRateLimit),

			// Lambda
			lambda.NewFunctionSource(lambdaClient, *callerID.Account, region),
			lambda.NewLayerSource(lambdaClient, *callerID.Account, region),
			lambda.NewLayerVersionSource(lambdaClient, *callerID.Account, region),

			// ECS
			ecs.NewCapacityProviderSource(ecsClient, *callerID.Account, region),
			ecs.NewClusterSource(ecsClient, *callerID.Account, region),
			ecs.NewContainerInstanceSource(ecsClient, *callerID.Account, region),
			ecs.NewServiceSource(ecsClient, *callerID.Account, region),
			ecs.NewTaskDefinitionSource(ecsClient, *callerID.Account, region),
			ecs.NewTaskSource(ecsClient, *callerID.Account, region),

			// DynamoDB
			dynamodb.NewBackupSource(dynamodbClient, *callerID.Account, region),
			dynamodb.NewTableSource(dynamodbClient, *callerID.Account, region),

			// RDS
			rds.NewDBClusterParameterGroupSource(rdsClient, *callerID.Account, region),
			rds.NewDBClusterSource(rdsClient, *callerID.Account, region),
			rds.NewDBInstanceSource(rdsClient, *callerID.Account, region),
			rds.NewDBParameterGroupSource(rdsClient, *callerID.Account, region),
			rds.NewDBSubnetGroupSource(rdsClient, *callerID.Account, region),
			rds.NewOptionGroupSource(rdsClient, *callerID.Account, region),

			// Autoscaling
			autoscaling.NewAutoScalingGroupSource(autoscalingClient, *callerID.Account, region, &autoScalingRateLimit),

			// ELB
			elb.NewInstanceHealthSource(elbClient, *callerID.Account, region),
			elb.NewLoadBalancerSource(elbClient, *callerID.Account, region),

			// ELBv2
			elbv2.NewListenerSource(elbv2Client, *callerID.Account, region),
			elbv2.NewLoadBalancerSource(elbv2Client, *callerID.Account, region),
			elbv2.NewRuleSource(elbv2Client, *callerID.Account, region),
			elbv2.NewTargetGroupSource(elbv2Client, *callerID.Account, region),
			elbv2.NewTargetHealthSource(elbv2Client, *callerID.Account, region),

			// Network Firewall
			networkfirewall.NewFirewallSource(networkfirewallClient, *callerID.Account, region),
			networkfirewall.NewFirewallPolicySource(networkfirewallClient, *callerID.Account, region),
			networkfirewall.NewRuleGroupSource(networkfirewallClient, *callerID.Account, region),
			networkfirewall.NewTLSInspectionConfigurationSource(networkfirewallClient, *callerID.Account, region),

			// Direct Connect
			directconnect.NewDirectConnectGatewaySource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewDirectConnectGatewayAssociationSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewDirectConnectGatewayAssociationProposalSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewConnectionSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewDirectConnectGatewayAttachmentSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewVirtualInterfaceSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewVirtualGatewaySource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewCustomerMetadataSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewLagSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewLocationSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewHostedConnectionSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewInterconnectSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),
			directconnect.NewRouterConfigurationSource(directconnectClient, *callerID.Account, region, &directConnectRateLimit),

			// Network Manager
			networkmanager.NewGlobalNetworkSource(networkmanagerClient, *callerID.Account, region),
			networkmanager.NewSiteSource(networkmanagerClient, *callerID.Account, region, &networkManagerRateLimit),
			networkmanager.NewVPCAttachmentSource(networkmanagerClient, *callerID.Account, region, &networkManagerRateLimit),

			// SQS
			sqs.NewQueueSource(sqsClient, *callerID.Account, region),

			// SNS
			sns.NewSubscriptionSource(snsClient, *callerID.Account, region),
			sns.NewTopicSource(snsClient, *callerID.Account, region),
			sns.NewPlatformApplicationSource(snsClient, *callerID.Account, region),
			sns.NewEndpointSource(snsClient, *callerID.Account, region),
			sns.NewDataProtectionPolicySource(snsClient, *callerID.Account, region),
		}

		e.AddSources(sources...)

		// Add "global" sources (those that aren't tied to a region, like
		// cloudfront). but only do this once for the first region. For
		// these APIs it doesn't matter which region we call them from, we
		// get global results
		if !globalDone {
			e.AddSources(
				// Cloudfront
				cloudfront.NewCachePolicySource(cloudfrontClient, *callerID.Account),
				cloudfront.NewContinuousDeploymentPolicySource(cloudfrontClient, *callerID.Account),
				cloudfront.NewDistributionSource(cloudfrontClient, *callerID.Account),
				cloudfront.NewFunctionSource(cloudfrontClient, *callerID.Account),
				cloudfront.NewKeyGroupSource(cloudfrontClient, *callerID.Account),
				cloudfront.NewOriginAccessControlSource(cloudfrontClient, *callerID.Account),
				cloudfront.NewOriginRequestPolicySource(cloudfrontClient, *callerID.Account),
				cloudfront.NewResponseHeadersPolicySource(cloudfrontClient, *callerID.Account),
				cloudfront.NewRealtimeLogConfigsSource(cloudfrontClient, *callerID.Account),
				cloudfront.NewStreamingDistributionSource(cloudfrontClient, *callerID.Account),

				// S3
				s3.NewS3Source(cfg, *callerID.Account),
			)
			globalDone = true
		}
	}

	return e, nil
}
