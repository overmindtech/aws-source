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
// engine, and an error if any. The context provided will be used for the rate
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

	var globalDone atomic.Bool

	p := pool.New().WithContext(ctx).WithCancelOnError()

	for _, region := range awsAuthConfig.Regions {
		region = strings.Trim(region, " ")
		p.Go(func(ctx context.Context) error {
			configCtx, configCancel := context.WithTimeout(ctx, 10*time.Second)
			defer configCancel()

			cfg, err := awsAuthConfig.GetAWSConfig(region)
			if err != nil {
				configCancel()
				return fmt.Errorf("error getting AWS config for region %v: %w", region, err)
			}

			// Add OTel instrumentation
			cfg.HTTPClient = &http.Client{
				Transport: otelhttp.NewTransport(http.DefaultTransport),
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
				log.WithError(err).WithFields(lf).Error("Error retrieving account information")
				return fmt.Errorf("error getting caller identity for region %v: %w", region, err)
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
				ec2.NewAddressSource(ec2Client, *callerID.Account, region),
				ec2.NewCapacityReservationFleetSource(ec2Client, *callerID.Account, region),
				ec2.NewCapacityReservationSource(ec2Client, *callerID.Account, region),
				ec2.NewEgressOnlyInternetGatewaySource(ec2Client, *callerID.Account, region),
				ec2.NewIamInstanceProfileAssociationSource(ec2Client, *callerID.Account, region),
				ec2.NewImageSource(ec2Client, *callerID.Account, region),
				ec2.NewInstanceEventWindowSource(ec2Client, *callerID.Account, region),
				ec2.NewInstanceSource(ec2Client, *callerID.Account, region),
				ec2.NewInstanceStatusSource(ec2Client, *callerID.Account, region),
				ec2.NewInternetGatewaySource(ec2Client, *callerID.Account, region),
				ec2.NewKeyPairSource(ec2Client, *callerID.Account, region),
				ec2.NewLaunchTemplateSource(ec2Client, *callerID.Account, region),
				ec2.NewLaunchTemplateVersionSource(ec2Client, *callerID.Account, region),
				ec2.NewNatGatewaySource(ec2Client, *callerID.Account, region),
				ec2.NewNetworkAclSource(ec2Client, *callerID.Account, region),
				ec2.NewNetworkInterfacePermissionSource(ec2Client, *callerID.Account, region),
				ec2.NewNetworkInterfaceSource(ec2Client, *callerID.Account, region),
				ec2.NewPlacementGroupSource(ec2Client, *callerID.Account, region),
				ec2.NewReservedInstanceSource(ec2Client, *callerID.Account, region),
				ec2.NewRouteTableSource(ec2Client, *callerID.Account, region),
				ec2.NewSecurityGroupSource(ec2Client, *callerID.Account, region),
				ec2.NewSnapshotSource(ec2Client, *callerID.Account, region),
				ec2.NewSubnetSource(ec2Client, *callerID.Account, region),
				ec2.NewVolumeSource(ec2Client, *callerID.Account, region),
				ec2.NewVolumeStatusSource(ec2Client, *callerID.Account, region),
				ec2.NewVpcPeeringConnectionSource(ec2Client, *callerID.Account, region),
				ec2.NewVpcSource(ec2Client, *callerID.Account, region),

				// EFS (I'm assuming it shares its rate limit with EC2))
				efs.NewAccessPointSource(efsClient, *callerID.Account, region),
				efs.NewBackupPolicySource(efsClient, *callerID.Account, region),
				efs.NewFileSystemSource(efsClient, *callerID.Account, region),
				efs.NewMountTargetSource(efsClient, *callerID.Account, region),
				efs.NewReplicationConfigurationSource(efsClient, *callerID.Account, region),

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
				iam.NewGroupSource(iamClient, *callerID.Account, region),
				iam.NewInstanceProfileSource(iamClient, *callerID.Account, region),
				iam.NewPolicySource(iamClient, *callerID.Account, region),
				iam.NewRoleSource(iamClient, *callerID.Account, region),
				iam.NewUserSource(iamClient, *callerID.Account, region),

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
				autoscaling.NewAutoScalingGroupSource(autoscalingClient, *callerID.Account, region),

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
				directconnect.NewDirectConnectGatewaySource(directconnectClient, *callerID.Account, region),
				directconnect.NewDirectConnectGatewayAssociationSource(directconnectClient, *callerID.Account, region),
				directconnect.NewDirectConnectGatewayAssociationProposalSource(directconnectClient, *callerID.Account, region),
				directconnect.NewConnectionSource(directconnectClient, *callerID.Account, region),
				directconnect.NewDirectConnectGatewayAttachmentSource(directconnectClient, *callerID.Account, region),
				directconnect.NewVirtualInterfaceSource(directconnectClient, *callerID.Account, region),
				directconnect.NewVirtualGatewaySource(directconnectClient, *callerID.Account, region),
				directconnect.NewCustomerMetadataSource(directconnectClient, *callerID.Account, region),
				directconnect.NewLagSource(directconnectClient, *callerID.Account, region),
				directconnect.NewLocationSource(directconnectClient, *callerID.Account, region),
				directconnect.NewHostedConnectionSource(directconnectClient, *callerID.Account, region),
				directconnect.NewInterconnectSource(directconnectClient, *callerID.Account, region),
				directconnect.NewRouterConfigurationSource(directconnectClient, *callerID.Account, region),

				// Network Manager
				networkmanager.NewConnectAttachmentSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewConnectPeerAssociationSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewConnectPeerSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewCoreNetworkPolicySource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewCoreNetworkSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewNetworkResourceRelationshipsSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewSiteToSiteVpnAttachmentSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewTransitGatewayConnectPeerAssociationSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewTransitGatewayPeeringSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewTransitGatewayRegistrationSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewTransitGatewayRouteTableAttachmentSource(networkmanagerClient, *callerID.Account, region),
				networkmanager.NewVPCAttachmentSource(networkmanagerClient, *callerID.Account, region),

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
