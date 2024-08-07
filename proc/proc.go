package proc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
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
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	awsnetworkfirewall "github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awsnetworkmanager "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sourcegraph/conc/pool"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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
func InitializeAwsSourceEngine(ctx context.Context, natsOptions auth.NATSOptions, maxParallel int, configs ...aws.Config) (*discovery.Engine, error) {
	e, err := discovery.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("error initializing Engine: %w", err)
	}

	e.Name = "aws-source"
	e.NATSOptions = &natsOptions
	e.MaxParallelExecutions = maxParallel

	if len(configs) == 0 {
		return nil, errors.New("No configs specified")
	}

	var globalDone atomic.Bool

	p := pool.New().WithContext(ctx).WithCancelOnError()

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

			sources := []discovery.Source{
				// EC2
				ec2.NewAddressSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewCapacityReservationFleetSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewCapacityReservationSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewEgressOnlyInternetGatewaySource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewIamInstanceProfileAssociationSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewImageSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewInstanceEventWindowSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewInstanceSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewInstanceStatusSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewInternetGatewaySource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewKeyPairSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewLaunchTemplateSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewLaunchTemplateVersionSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewNatGatewaySource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewNetworkAclSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewNetworkInterfacePermissionSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewNetworkInterfaceSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewPlacementGroupSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewReservedInstanceSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewRouteTableSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewSecurityGroupRuleSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewSecurityGroupSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewSnapshotSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewSubnetSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewVolumeSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewVolumeStatusSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewVpcEndpointSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewVpcPeeringConnectionSource(ec2Client, *callerID.Account, cfg.Region),
				ec2.NewVpcSource(ec2Client, *callerID.Account, cfg.Region),

				// EFS (I'm assuming it shares its rate limit with EC2))
				efs.NewAccessPointSource(efsClient, *callerID.Account, cfg.Region),
				efs.NewBackupPolicySource(efsClient, *callerID.Account, cfg.Region),
				efs.NewFileSystemSource(efsClient, *callerID.Account, cfg.Region),
				efs.NewMountTargetSource(efsClient, *callerID.Account, cfg.Region),
				efs.NewReplicationConfigurationSource(efsClient, *callerID.Account, cfg.Region),

				// EKS
				eks.NewAddonSource(eksClient, *callerID.Account, cfg.Region),
				eks.NewClusterSource(eksClient, *callerID.Account, cfg.Region),
				eks.NewFargateProfileSource(eksClient, *callerID.Account, cfg.Region),
				eks.NewNodegroupSource(eksClient, *callerID.Account, cfg.Region),

				// Route 53
				route53.NewHealthCheckSource(route53Client, *callerID.Account, cfg.Region),
				route53.NewHostedZoneSource(route53Client, *callerID.Account, cfg.Region),
				route53.NewResourceRecordSetSource(route53Client, *callerID.Account, cfg.Region),

				// Cloudwatch
				cloudwatch.NewAlarmSource(cloudwatchClient, *callerID.Account, cfg.Region),

				// IAM
				iam.NewGroupSource(iamClient, *callerID.Account, cfg.Region),
				iam.NewInstanceProfileSource(iamClient, *callerID.Account, cfg.Region),
				iam.NewPolicySource(iamClient, *callerID.Account, cfg.Region),
				iam.NewRoleSource(iamClient, *callerID.Account, cfg.Region),
				iam.NewUserSource(iamClient, *callerID.Account, cfg.Region),

				// Lambda
				lambda.NewFunctionSource(lambdaClient, *callerID.Account, cfg.Region),
				lambda.NewLayerSource(lambdaClient, *callerID.Account, cfg.Region),
				lambda.NewLayerVersionSource(lambdaClient, *callerID.Account, cfg.Region),

				// ECS
				ecs.NewCapacityProviderSource(ecsClient, *callerID.Account, cfg.Region),
				ecs.NewClusterSource(ecsClient, *callerID.Account, cfg.Region),
				ecs.NewContainerInstanceSource(ecsClient, *callerID.Account, cfg.Region),
				ecs.NewServiceSource(ecsClient, *callerID.Account, cfg.Region),
				ecs.NewTaskDefinitionSource(ecsClient, *callerID.Account, cfg.Region),
				ecs.NewTaskSource(ecsClient, *callerID.Account, cfg.Region),

				// DynamoDB
				dynamodb.NewBackupSource(dynamodbClient, *callerID.Account, cfg.Region),
				dynamodb.NewTableSource(dynamodbClient, *callerID.Account, cfg.Region),

				// RDS
				rds.NewDBClusterParameterGroupSource(rdsClient, *callerID.Account, cfg.Region),
				rds.NewDBClusterSource(rdsClient, *callerID.Account, cfg.Region),
				rds.NewDBInstanceSource(rdsClient, *callerID.Account, cfg.Region),
				rds.NewDBParameterGroupSource(rdsClient, *callerID.Account, cfg.Region),
				rds.NewDBSubnetGroupSource(rdsClient, *callerID.Account, cfg.Region),
				rds.NewOptionGroupSource(rdsClient, *callerID.Account, cfg.Region),

				// Autoscaling
				autoscaling.NewAutoScalingGroupSource(autoscalingClient, *callerID.Account, cfg.Region),

				// ELB
				elb.NewInstanceHealthSource(elbClient, *callerID.Account, cfg.Region),
				elb.NewLoadBalancerSource(elbClient, *callerID.Account, cfg.Region),

				// ELBv2
				elbv2.NewListenerSource(elbv2Client, *callerID.Account, cfg.Region),
				elbv2.NewLoadBalancerSource(elbv2Client, *callerID.Account, cfg.Region),
				elbv2.NewRuleSource(elbv2Client, *callerID.Account, cfg.Region),
				elbv2.NewTargetGroupSource(elbv2Client, *callerID.Account, cfg.Region),
				elbv2.NewTargetHealthSource(elbv2Client, *callerID.Account, cfg.Region),

				// Network Firewall
				networkfirewall.NewFirewallSource(networkfirewallClient, *callerID.Account, cfg.Region),
				networkfirewall.NewFirewallPolicySource(networkfirewallClient, *callerID.Account, cfg.Region),
				networkfirewall.NewRuleGroupSource(networkfirewallClient, *callerID.Account, cfg.Region),
				networkfirewall.NewTLSInspectionConfigurationSource(networkfirewallClient, *callerID.Account, cfg.Region),

				// Direct Connect
				directconnect.NewDirectConnectGatewaySource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewDirectConnectGatewayAssociationSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewDirectConnectGatewayAssociationProposalSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewConnectionSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewDirectConnectGatewayAttachmentSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewVirtualInterfaceSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewVirtualGatewaySource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewCustomerMetadataSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewLagSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewLocationSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewHostedConnectionSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewInterconnectSource(directconnectClient, *callerID.Account, cfg.Region),
				directconnect.NewRouterConfigurationSource(directconnectClient, *callerID.Account, cfg.Region),

				// Network Manager
				networkmanager.NewConnectAttachmentSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewConnectPeerAssociationSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewConnectPeerSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewCoreNetworkPolicySource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewCoreNetworkSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewNetworkResourceRelationshipsSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewSiteToSiteVpnAttachmentSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewTransitGatewayConnectPeerAssociationSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewTransitGatewayPeeringSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewTransitGatewayRegistrationSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewTransitGatewayRouteTableAttachmentSource(networkmanagerClient, *callerID.Account, cfg.Region),
				networkmanager.NewVPCAttachmentSource(networkmanagerClient, *callerID.Account, cfg.Region),

				// SQS
				sqs.NewQueueSource(sqsClient, *callerID.Account, cfg.Region),

				// SNS
				sns.NewSubscriptionSource(snsClient, *callerID.Account, cfg.Region),
				sns.NewTopicSource(snsClient, *callerID.Account, cfg.Region),
				sns.NewPlatformApplicationSource(snsClient, *callerID.Account, cfg.Region),
				sns.NewEndpointSource(snsClient, *callerID.Account, cfg.Region),
				sns.NewDataProtectionPolicySource(snsClient, *callerID.Account, cfg.Region),
			}

			e.AddSources(sources...)

			// Add "global" sources (those that aren't tied to a region, like
			// cloudfront). but only do this once for the first region. For
			// these APIs it doesn't matter which region we call them from, we
			// get global results
			if globalDone.CompareAndSwap(false, true) {
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

					// Networkmanager
					networkmanager.NewGlobalNetworkSource(networkmanagerClient, *callerID.Account),
					networkmanager.NewSiteSource(networkmanagerClient, *callerID.Account),
					networkmanager.NewLinkSource(networkmanagerClient, *callerID.Account),
					networkmanager.NewDeviceSource(networkmanagerClient, *callerID.Account),
					networkmanager.NewLinkAssociationSource(networkmanagerClient, *callerID.Account),
					networkmanager.NewConnectionSource(networkmanagerClient, *callerID.Account),
				)
			}
			return nil
		})
	}

	err = p.Wait()
	if err != nil {
		return nil, err
	}

	return e, nil
}
