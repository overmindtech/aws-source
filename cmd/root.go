package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/aws-source/adapters/apigateway"
	"github.com/overmindtech/aws-source/adapters/autoscaling"
	"github.com/overmindtech/aws-source/adapters/cloudfront"
	"github.com/overmindtech/aws-source/adapters/cloudwatch"
	"github.com/overmindtech/aws-source/adapters/directconnect"
	"github.com/overmindtech/aws-source/adapters/dynamodb"
	"github.com/overmindtech/aws-source/adapters/ec2"
	"github.com/overmindtech/aws-source/adapters/ecs"
	"github.com/overmindtech/aws-source/adapters/efs"
	"github.com/overmindtech/aws-source/adapters/eks"
	"github.com/overmindtech/aws-source/adapters/elb"
	"github.com/overmindtech/aws-source/adapters/elbv2"
	"github.com/overmindtech/aws-source/adapters/iam"
	"github.com/overmindtech/aws-source/adapters/kms"
	"github.com/overmindtech/aws-source/adapters/lambda"
	"github.com/overmindtech/aws-source/adapters/networkfirewall"
	"github.com/overmindtech/aws-source/adapters/networkmanager"
	"github.com/overmindtech/aws-source/adapters/rds"
	"github.com/overmindtech/aws-source/adapters/route53"
	"github.com/overmindtech/aws-source/adapters/s3"
	"github.com/overmindtech/aws-source/adapters/sns"
	"github.com/overmindtech/aws-source/adapters/sqs"
	"github.com/overmindtech/aws-source/proc"
	"github.com/overmindtech/aws-source/tracing"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdp-go/auth"
	"github.com/overmindtech/sdp-go/sdpconnect"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"golang.org/x/oauth2"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aws-source",
	Short: "Remote primary source for AWS",
	Long: `This sources looks for AWS resources in your account.
`,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			err := recover()

			if err != nil {
				sentry.CurrentHub().Recover(err)
				defer sentry.Flush(time.Second * 5)
				panic(err)
			}
		}()

		var err error

		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsJWT := viper.GetString("nats-jwt")
		natsNKeySeed := viper.GetString("nats-nkey-seed")
		maxParallel := viper.GetInt("max-parallel")
		apiKey := viper.GetString("api-key")
		app := viper.GetString("app")
		healthCheckPort := viper.GetInt("health-check-port")
		natsConnectionName := viper.GetString("nats-connection-name")
		sourceName := viper.GetString("source-name")
		sourceUUIDString := viper.GetString("source-uuid")

		awsAuthConfig := proc.AwsAuthConfig{
			Strategy:        viper.GetString("aws-access-strategy"),
			AccessKeyID:     viper.GetString("aws-access-key-id"),
			SecretAccessKey: viper.GetString("aws-secret-access-key"),
			ExternalID:      viper.GetString("aws-external-id"),
			TargetRoleARN:   viper.GetString("aws-target-role-arn"),
			Profile:         viper.GetString("aws-profile"),
			AutoConfig:      viper.GetBool("auto-config"),
		}

		err = viper.UnmarshalKey("aws-regions", &awsAuthConfig.Regions)
		if err != nil {
			log.WithError(err).Fatal("Could not parse aws-regions")
		}

		var natsNKeySeedLog string
		if natsNKeySeed != "" {
			natsNKeySeedLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":         natsServers,
			"nats-jwt":             natsJWT,
			"nats-nkey-seed":       natsNKeySeedLog,
			"nats-connection-name": natsConnectionName,
			"max-parallel":         maxParallel,
			"aws-regions":          awsAuthConfig.Regions,
			"aws-access-strategy":  awsAuthConfig.Strategy,
			"aws-external-id":      awsAuthConfig.ExternalID,
			"aws-target-role-arn":  awsAuthConfig.TargetRoleARN,
			"aws-profile":          awsAuthConfig.Profile,
			"auto-config":          awsAuthConfig.AutoConfig,
			"health-check-port":    healthCheckPort,
			"app":                  app,
			"source-name":          sourceName,
			"source-uuid":          sourceUUIDString,
		}).Info("Got config")

		var sourceUUID uuid.UUID
		if sourceUUIDString == "" {
			sourceUUID = uuid.New()
		} else {
			sourceUUID, err = uuid.Parse(sourceUUIDString)

			if err != nil {
				log.WithError(err).Fatal("Could not parse source UUID")
			}
		}

		// Determine the required Overmind URLs
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		oi, err := sdp.NewOvermindInstance(ctx, app)
		if err != nil {
			log.WithError(err).Fatal("Could not determine Overmind instance URLs")
		}

		// Validate the auth params and create a token client if we are using
		// auth
		var natsTokenClient auth.TokenClient
		var authenticatedClient http.Client
		var heartbeatOptions *discovery.HeartbeatOptions
		if apiKey != "" {
			natsTokenClient, err = auth.NewAPIKeyClient(oi.ApiUrl.String(), apiKey)

			if err != nil {
				sentry.CaptureException(err)

				log.WithError(err).Fatal("Could not create API key client")
			}

			tokenSource := auth.NewAPIKeyTokenSource(apiKey, oi.ApiUrl.String())
			transport := oauth2.Transport{
				Source: tokenSource,
				Base:   http.DefaultTransport,
			}
			authenticatedClient = http.Client{
				Transport: otelhttp.NewTransport(&transport),
			}

			heartbeatOptions = &discovery.HeartbeatOptions{
				ManagementClient: sdpconnect.NewManagementServiceClient(
					&authenticatedClient,
					oi.ApiUrl.String(),
				),
				Frequency: time.Second * 30,
			}
		} else if natsJWT != "" || natsNKeySeed != "" {
			natsTokenClient, err = createTokenClient(natsJWT, natsNKeySeed)
			log.Info("Using NATS authentication, no heartbeat will be sent")

			if err != nil {
				log.WithError(err).Fatal("Error validating NATS authentication info")
			}
		}

		natsOptions := auth.NATSOptions{
			NumRetries:        -1,
			RetryDelay:        5 * time.Second,
			Servers:           natsServers,
			ConnectionName:    natsConnectionName,
			ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
			MaxReconnects:     -1,
			ReconnectWait:     1 * time.Second,
			ReconnectJitter:   1 * time.Second,
			TokenClient:       natsTokenClient,
		}

		rateLimitContext, rateLimitCancel := context.WithCancel(context.Background())
		defer rateLimitCancel()

		configs, err := proc.CreateAWSConfigs(awsAuthConfig)
		if err != nil {
			log.WithError(err).Fatal("Could not create AWS configs")
		}

		e, err := proc.InitializeAwsSourceEngine(
			rateLimitContext,
			sourceName,
			tracing.ServiceVersion,
			sourceUUID,
			natsOptions,
			heartbeatOptions,
			maxParallel,
			999_999, // Very high max retries as it'll time out after 15min anyway
			configs...,
		)
		if err != nil {
			log.WithError(err).Fatal("Could not initialize AWS source")
		}

		// Start HTTP server for status
		healthCheckPath := "/healthz"

		http.HandleFunc(healthCheckPath, func(rw http.ResponseWriter, r *http.Request) {
			ctx, span := healthCheckTracer().Start(r.Context(), "healthcheck")
			defer span.End()

			err := e.HealthCheck(ctx)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprint(rw, "ok")
		})

		log.WithFields(log.Fields{
			"port": healthCheckPort,
			"path": healthCheckPath,
		}).Debug("Starting healthcheck server")

		go func() {
			defer sentry.Recover()

			server := &http.Server{
				Addr:         fmt.Sprintf(":%v", healthCheckPort),
				Handler:      nil,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
			}
			err := server.ListenAndServe()

			log.WithError(err).WithFields(log.Fields{
				"port": healthCheckPort,
				"path": healthCheckPath,
			}).Error("Could not start HTTP server for /healthz health checks")
		}()

		err = e.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Could not start engine")
		}

		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs

		log.Info("Stopping engine")

		err = e.Stop()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not stop engine")

			os.Exit(1)
		}
		log.Info("Stopped")

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(docJSONCmd)
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Fatal("Could not determine hostname for use in NATS connection name and source name")
	}

	// General config options
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/srcman/config/source.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")

	// Config required by all sources in order to connect to NATS. You shouldn't
	// need to change these
	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to")
	rootCmd.PersistentFlags().String("nats-jwt", "", "The JWT token that should be used to authenticate to NATS, provided in raw format e.g. eyJ0eXAiOiJKV1Q...")
	rootCmd.PersistentFlags().String("nats-nkey-seed", "", "The NKey seed which corresponds to the NATS JWT e.g. SUAFK6QUC...")
	rootCmd.PersistentFlags().String("nats-connection-name", hostname, "The name that the source should use to connect to NATS")

	rootCmd.PersistentFlags().String("api-key", "", "The API key to use to authenticate to the Overmind API")
	// Support API Keys in the environment
	err = viper.BindEnv("api-key", "OVM_API_KEY", "API_KEY")
	if err != nil {
		log.WithError(err).Fatal("could not bind api key to env")
	}

	rootCmd.PersistentFlags().String("app", "https://app.overmind.tech", "The URL of the Overmind app to use")
	rootCmd.PersistentFlags().Int("max-parallel", 2_000, "Max number of requests to run in parallel")

	// Custom flags for this source
	rootCmd.PersistentFlags().String("aws-access-strategy", "defaults", "The strategy to use to access this customer's AWS account. Valid values: 'access-key', 'external-id', 'sso-profile', 'defaults'. Default: 'defaults'.")
	rootCmd.PersistentFlags().String("aws-access-key-id", "", "The ID of the access key to use")
	rootCmd.PersistentFlags().String("aws-secret-access-key", "", "The secret access key to use for auth")
	rootCmd.PersistentFlags().String("aws-external-id", "", "The external ID to use when assuming the customer's role")
	rootCmd.PersistentFlags().String("aws-target-role-arn", "", "The role to assume in the customer's account")
	rootCmd.PersistentFlags().String("aws-profile", "", "The AWS SSO Profile to use. Defaults to $AWS_PROFILE, then whatever the AWS SDK's SSO config defaults to")
	rootCmd.PersistentFlags().String("aws-regions", "", "Comma-separated list of AWS regions that this source should operate in")
	rootCmd.PersistentFlags().BoolP("auto-config", "a", false, "Use the local AWS config, the same as the AWS CLI could use. This can be set up with \"aws configure\"")
	rootCmd.PersistentFlags().IntP("health-check-port", "", 8080, "The port that the health check should run on")
	rootCmd.PersistentFlags().String("source-name", fmt.Sprintf("aws-source-%v", hostname), "The name of the source")
	rootCmd.PersistentFlags().String("source-uuid", "", "The UUID of the source, is this is blank it will be auto-generated. This is used in heartbeats and shouldn't be supplied usually")

	// tracing
	rootCmd.PersistentFlags().String("honeycomb-api-key", "", "If specified, configures opentelemetry libraries to submit traces to honeycomb")
	rootCmd.PersistentFlags().String("sentry-dsn", "", "If specified, configures sentry libraries to capture errors")
	rootCmd.PersistentFlags().String("run-mode", "release", "Set the run mode for this service, 'release', 'debug' or 'test'. Defaults to 'release'.")

	// Bind these to viper
	cobra.CheckErr(viper.BindPFlags(rootCmd.PersistentFlags()))

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not parse log level")
		}

		log.AddHook(TerminationLogHook{})

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				err = viper.BindPFlag(f.Name, f)
				if err != nil {
					log.WithError(err).Fatal("could not bind flag to viper")
				}
			}
		})

		honeycomb_api_key := viper.GetString("honeycomb-api-key")
		tracingOpts := make([]otlptracehttp.Option, 0)
		if honeycomb_api_key != "" {
			tracingOpts = []otlptracehttp.Option{
				otlptracehttp.WithEndpoint("api.honeycomb.io"),
				otlptracehttp.WithHeaders(map[string]string{"x-honeycomb-team": honeycomb_api_key}),
			}
		}
		if err := tracing.InitTracing(tracingOpts...); err != nil {
			log.Fatal(err)
		}
	}
	// shut down tracing at the end of the process
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		tracing.ShutdownTracing()
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	replacer := strings.NewReplacer("-", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}

// createTokenClient Creates a basic token client that will authenticate to NATS
// using the given values
func createTokenClient(natsJWT string, natsNKeySeed string) (auth.TokenClient, error) {
	var kp nkeys.KeyPair
	var err error

	if natsJWT == "" {
		return nil, errors.New("nats-jwt was blank. This is required when using authentication")
	}

	if natsNKeySeed == "" {
		return nil, errors.New("nats-nkey-seed was blank. This is required when using authentication")
	}

	if _, err = jwt.DecodeUserClaims(natsJWT); err != nil {
		return nil, fmt.Errorf("could not parse nats-jwt: %w", err)
	}

	if kp, err = nkeys.FromSeed([]byte(natsNKeySeed)); err != nil {
		return nil, fmt.Errorf("could not parse nats-nkey-seed: %w", err)
	}

	return auth.NewBasicTokenClient(natsJWT, kp), nil
}

// TerminationLogHook A hook that logs fatal errors to the termination log
type TerminationLogHook struct{}

func (t TerminationLogHook) Levels() []log.Level {
	return []log.Level{log.FatalLevel}
}

func (t TerminationLogHook) Fire(e *log.Entry) error {
	// shutdown tracing first to ensure all spans are flushed
	tracing.ShutdownTracing()
	tLog, err := os.OpenFile("/dev/termination-log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	var message string

	message = e.Message

	for k, v := range e.Data {
		message = fmt.Sprintf("%v %v=%v", message, k, v)
	}

	_, err = tLog.WriteString(message)

	return err
}

// documentation subcommand for generating json
var docJSONCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate JSON documentation",
	Long:  `Generate JSON documentation for the source`,
	Run: func(cmd *cobra.Command, args []string) {
		allMetadata := []sdp.AdapterMetadata{
			apigateway.APIGatewayMetadata(),
			apigateway.RestAPIMetadata(),
			autoscaling.AutoScalingGroupMetadata(),
			cloudfront.CachePolicyMetadata(),
			cloudfront.ContinuousDeploymentPolicyMetadata(),
			cloudfront.DistributionMetadata(),
			cloudfront.FunctionMetadata(),
			cloudfront.KeyGroupMetadata(),
			cloudfront.OriginAccessControlMetadata(),
			cloudfront.OriginRequestPolicySourceMetadata(),
			cloudfront.RealtimeLogConfigsMetadata(),
			cloudfront.ResponseHeadersPolicyMetadata(),
			cloudfront.StreamingDistributionMetadata(),
			cloudwatch.AlarmMetadata(),
			directconnect.ConnectionMetadata(),
			directconnect.CustomerMetadata(),
			directconnect.DirectConnectGatewayAssociationMetadata(),
			directconnect.DirectConnectGatewayAssociationProposalMetadata(),
			directconnect.DirectConnectGatewayAttachmentMetadata(),
			directconnect.DirectConnectGatewayMetadata(),
			directconnect.HostedConnectionMetadata(),
			directconnect.InterconnectMetadata(),
			directconnect.LagMetadata(),
			directconnect.LocationMetadata(),
			directconnect.RouterConfigurationSourceMetadata(),
			directconnect.VirtualGatewayMetadata(),
			directconnect.VirtualInterfaceMetadata(),
			dynamodb.BackupMetadata(),
			dynamodb.TableMetadata(),
			ec2.AddressMetadata(),
			ec2.CapacityReservationFleetMetadata(),
			ec2.CapacityReservationMetadata(),
			ec2.EgressInternetGatewayMetadata(),
			ec2.IamInstanceProfileAssociationMetadata(),
			ec2.ImageMetadata(),
			ec2.InstanceEventWindowMetadata(),
			ec2.InstanceStatusMetadata(),
			ec2.InstanceMetadata(),
			ec2.InternetGatewayMetadata(),
			ec2.KeyPairMetadata(),
			ec2.LaunchTemplateVersionMetadata(),
			ec2.LaunchTemplateMetadata(),
			ec2.NatGatewayMetadata(),
			ec2.NetworkAclMetadata(),
			ec2.NetworkInterfacePermissionMetadata(),
			ec2.NetworkInterfaceMetadata(),
			ec2.PlacementGroupMetadata(),
			ec2.ReservedInstanceMetadata(),
			ec2.RouteTableMetadata(),
			ec2.SecurityGroupRuleMetadata(),
			ec2.SecurityGroupMetadata(),
			ec2.SnapshotMetadata(),
			ec2.SubnetMetadata(),
			ec2.VolumeStatusMetadata(),
			ec2.VolumeMetadata(),
			ec2.VpcEndpointMetadata(),
			ec2.VpcPeeringConnectionMetadata(),
			ec2.VpcMetadata(),
			ecs.CapacityProviderMetadata(),
			ecs.ClusterMetadata(),
			ecs.ContainerInstanceMetadata(),
			ecs.ServiceMetadata(),
			ecs.TaskDefinitionMetadata(),
			ecs.TaskMetadata(),
			efs.AccessPointMetadata(),
			efs.BackupPolicyMetadata(),
			efs.FileSystemMetadata(),
			efs.MountTargetMetadata(),
			efs.ReplicationConfigurationMetadata(),
			eks.AddonMetadata(),
			eks.ClusterMetadata(),
			eks.FargateProfileMetadata(),
			eks.NodeGroupMetadata(),
			elb.LoadBalancerMetadata(),
			elb.InstanceHealthMetadata(),
			elbv2.LoadBalancerMetadata(),
			elbv2.ListenerMetadata(),
			elbv2.RuleMetadata(),
			elbv2.TargetGroupMetadata(),
			elbv2.TargetHealthMetadata(),
			iam.GroupMetadata(),
			iam.InstanceProfileMetadata(),
			iam.PolicyMetadata(),
			iam.RoleMetadata(),
			iam.UserMetadata(),
			kms.AliasMetadata(),
			kms.CustomKeyStoreMetadata(),
			kms.GrantMetadata(),
			kms.KeyMetadata(),
			kms.KeyPolicyMetadata(),
			lambda.FunctionMetadata(),
			lambda.LayerVersionMetadata(),
			lambda.LayerMetadata(),
			networkfirewall.FirewallPolicyMetadata(),
			networkfirewall.FirewallMetadata(),
			networkfirewall.RuleGroupMetadata(),
			networkfirewall.TLSInspectionConfigurationMetadata(),
			networkmanager.ConnectAttachmentMetadata(),
			networkmanager.ConnectPeerAssociationMetadata(),
			networkmanager.ConnectPeerMetadata(),
			networkmanager.ConnectionMetadata(),
			networkmanager.CoreNetworkPolicyMetadata(),
			networkmanager.CoreNetworkMetadata(),
			networkmanager.DeviceMetadata(),
			networkmanager.GlobalNetworkMetadata(),
			networkmanager.LinkAssociationMetadata(),
			networkmanager.LinkMetadata(),
			networkmanager.NetworkResourceRelationshipMetadata(),
			networkmanager.SiteToSiteVpnAttachmentMetadata(),
			networkmanager.SiteMetadata(),
			networkmanager.TransitGatewayConnectPeerAssociationMetadata(),
			networkmanager.TransitGatewayPeeringMetadata(),
			networkmanager.TransitGatewayRegistrationMetadata(),
			networkmanager.TransitGatewayRouteTableAttachmentMetadata(),
			networkmanager.VPCAttachmentMetadata(),
			rds.DBClusterParameterGroupMetadata(),
			rds.DBClusterMetadata(),
			rds.DBInstanceMetadata(),
			rds.DBParameterGroupMetadata(),
			rds.DBSubnetGroupMetadata(),
			rds.OptionGroupMetadata(),
			route53.HealthCheckMetadata(),
			route53.HostedZoneMetadata(),
			route53.ResourceRecordSetMetadata(),
			s3.S3Metadata(),
			sns.DataProtectionPolicyMetadata(),
			sns.EndpointMetadata(),
			sns.PlatformApplicationMetadata(),
			sns.SubscriptionMetadata(),
			sns.TopicMetadata(),
			sqs.QueueMetadata(),
		}
		err := discovery.AdapterMetadataToJSONFile(allMetadata, "docs-data")
		if err != nil {
			log.WithError(err).Fatal("Could not generate JSON documentation")
		}
	},
}
