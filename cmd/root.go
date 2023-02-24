package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/getsentry/sentry-go"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/aws-source/sources/autoscaling"
	"github.com/overmindtech/aws-source/sources/dynamodb"
	"github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/aws-source/sources/ecs"
	"github.com/overmindtech/aws-source/sources/eks"
	"github.com/overmindtech/aws-source/sources/elasticloadbalancing"
	"github.com/overmindtech/aws-source/sources/iam"
	"github.com/overmindtech/aws-source/sources/lambda"
	"github.com/overmindtech/aws-source/sources/rds"
	"github.com/overmindtech/aws-source/sources/route53"
	"github.com/overmindtech/aws-source/sources/s3"
	"github.com/overmindtech/connect"
	"github.com/overmindtech/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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

		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := viper.GetString("nats-name-prefix")
		natsJWT := viper.GetString("nats-jwt")
		natsNKeySeed := viper.GetString("nats-nkey-seed")
		maxParallel := viper.GetInt("max-parallel")
		hostname, err := os.Hostname()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not determine hostname for use in NATS connection name")

			os.Exit(1)
		}

		var regions []string
		viper.UnmarshalKey("aws-regions", &regions)

		strategy := viper.GetString("aws-access-strategy")
		accessKeyID := viper.GetString("aws-access-key-id")
		secretAccessKey := viper.GetString("aws-secret-access-key")
		externalID := viper.GetString("aws-external-id")
		targetRoleARN := viper.GetString("aws-target-role-arn")
		autoConfig := viper.GetBool("auto-config")
		healthCheckPort := viper.GetInt("health-check-port")

		var natsNKeySeedLog string
		var tokenClient connect.TokenClient

		if natsNKeySeed != "" {
			natsNKeySeedLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":        natsServers,
			"nats-name-prefix":    natsNamePrefix,
			"nats-jwt":            natsJWT,
			"nats-nkey-seed":      natsNKeySeedLog,
			"max-parallel":        maxParallel,
			"aws-regions":         regions,
			"aws-access-strategy": strategy,
			"aws-external-id":     externalID,
			"aws-target-role-arn": targetRoleARN,
			"auto-config":         autoConfig,
			"health-check-port":   healthCheckPort,
		}).Info("Got config")

		// Validate the auth params and create a token client if we are using
		// auth
		if natsJWT != "" || natsNKeySeed != "" {
			var err error

			tokenClient, err = createTokenClient(natsJWT, natsNKeySeed)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Fatal("Error validating authentication info")
			}
		}

		e, err := discovery.NewEngine()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("Error initializing Engine")
		}
		e.Name = "aws-source"
		e.NATSOptions = &connect.NATSOptions{
			NumRetries:        -1,
			RetryDelay:        5 * time.Second,
			Servers:           natsServers,
			ConnectionName:    fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
			ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
			MaxReconnects:     -1,
			ReconnectWait:     1 * time.Second,
			ReconnectJitter:   1 * time.Second,
			TokenClient:       tokenClient,
		}
		e.MaxParallelExecutions = maxParallel

		for _, region := range regions {
			region = strings.Trim(region, " ")

			configCtx, configCancel := context.WithTimeout(context.Background(), 10*time.Second)

			cfg, err := getAWSConfig(strategy, region, accessKeyID, secretAccessKey, externalID, targetRoleARN, autoConfig)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Fatal("Error loading config")
			}

			// Work out what account we're using. This will be used in item scopes
			stsClient := sts.NewFromConfig(cfg)

			var callerID *sts.GetCallerIdentityOutput

			callerID, err = stsClient.GetCallerIdentity(configCtx, &sts.GetCallerIdentityInput{})

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Fatal("Error retrieving account information")
			}

			// Cancel config load context and release resources
			configCancel()

			// Create an EC2 rate limit which limits the source to 50% of the
			// overall rate limit
			ec2RateLimit := ec2.LimitBucket{
				MaxCapacity: 50,
				RefillRate:  10,
			}

			// Apparently Autoscaling has a separate bucket to EC2 but I'm going
			// to assume the values are the same, the documentation for rate
			// limiting for everything other than EC2 is very poor
			autoScalingRateLimit := ec2.LimitBucket{
				MaxCapacity: 50,
				RefillRate:  10,
			}

			rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
			defer rateLimitCancel()

			ec2RateLimit.Start(rateLimitCtx)
			autoScalingRateLimit.Start(rateLimitCtx)

			sources := []discovery.Source{
				// ELB
				&elasticloadbalancing.ELBSource{
					Config:    cfg,
					AccountID: *callerID.Account,
				},
				&elasticloadbalancing.ELBv2Source{
					Config:    cfg,
					AccountID: *callerID.Account,
				},

				// EC2
				ec2.NewAvailabilityZoneSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewInstanceSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewSecurityGroupSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewVpcSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewVolumeSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewImageSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewAddressSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewInternetGatewaySource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewKeyPairSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewNatGatewaySource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewNetworkInterfaceSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewRegionSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewSubnetSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewEgressOnlyInternetGatewaySource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewInstanceStatusSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewSecurityGroupSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewInstanceEventWindowSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewLaunchTemplateSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewLaunchTemplateVersionSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewNetworkAclSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewNetworkInterfacePermissionSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewPlacementGroupSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewRouteTableSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewReservedInstanceSource(cfg, *callerID.Account, &ec2RateLimit),
				ec2.NewSnapshotSource(cfg, *callerID.Account, &ec2RateLimit),

				// S3
				s3.NewS3Source(cfg, *callerID.Account),

				// EKS
				eks.NewClusterSource(cfg, *callerID.Account, region),
				eks.NewAddonSource(cfg, *callerID.Account, region),
				eks.NewFargateProfileSource(cfg, *callerID.Account, region),
				eks.NewNodegroupSource(cfg, *callerID.Account, region),

				// Route 53
				route53.NewHostedZoneSource(cfg, *callerID.Account, region),
				route53.NewResourceRecordSetSource(cfg, *callerID.Account, region),

				// IAM
				iam.NewGroupSource(cfg, *callerID.Account, region),
				iam.NewUserSource(cfg, *callerID.Account, region),
				iam.NewRoleSource(cfg, *callerID.Account, region),
				iam.NewPolicySource(cfg, *callerID.Account, region),

				// Lambda
				lambda.NewFunctionSource(cfg, *callerID.Account, region),
				lambda.NewLayerSource(cfg, *callerID.Account, region),
				lambda.NewLayerVersionSource(cfg, *callerID.Account, region),

				// ECS
				ecs.NewClusterSource(cfg, *callerID.Account, region),
				ecs.NewCapacityProviderSource(cfg, *callerID.Account),
				ecs.NewContainerInstanceSource(cfg, *callerID.Account, region),
				ecs.NewServiceSource(cfg, *callerID.Account, region),
				ecs.NewTaskDefinitionSource(cfg, *callerID.Account, region),
				ecs.NewTaskSource(cfg, *callerID.Account, region),

				// DynamoDB
				dynamodb.NewTableSource(cfg, *callerID.Account, region),
				dynamodb.NewBackupSource(cfg, *callerID.Account, region),

				// RDS
				rds.NewDBInstanceSource(cfg, *callerID.Account),
				rds.NewDBClusterSource(cfg, *callerID.Account),
				rds.NewDBParameterGroupSource(cfg, *callerID.Account, region),
				rds.NewDBClusterParameterGroupSource(cfg, *callerID.Account, region),
				rds.NewDBSubnetGroupSource(cfg, *callerID.Account),
				rds.NewOptionGroupSource(cfg, *callerID.Account),

				// Autoscaling
				autoscaling.NewAutoScalingGroupSource(cfg, *callerID.Account, &autoScalingRateLimit),
			}

			e.AddSources(sources...)
		}

		// Start HTTP server for status
		healthCheckPath := "/healthz"

		http.HandleFunc(healthCheckPath, func(rw http.ResponseWriter, r *http.Request) {
			// Check that NATS is connected
			if !e.IsNATSConnected() {
				http.Error(rw, "NATS not connected", http.StatusInternalServerError)
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

			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", healthCheckPort), nil))
		}()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start HTTP server for /healthz health checks")
		}

		err = e.Start()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start engine")
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	// General config options
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/srcman/config/source.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")

	// Config required by all sources in order to connect to NATS. You shouldn't
	// need to change these
	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to")
	rootCmd.PersistentFlags().String("nats-name-prefix", "", "A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client")
	rootCmd.PersistentFlags().String("nats-jwt", "", "The JWT token that should be used to authenticate to NATS, provided in raw format e.g. eyJ0eXAiOiJKV1Q...")
	rootCmd.PersistentFlags().String("nats-nkey-seed", "", "The NKey seed which corresponds to the NATS JWT e.g. SUAFK6QUC...")
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 10), "Max number of requests to run in parallel")

	// Custom flags for this source
	rootCmd.PersistentFlags().String("aws-access-strategy", "access-key", "The strategy to use to access this customer's AWS account. Valid values: 'access-key', 'external-id'. Default: 'access-key'.")
	rootCmd.PersistentFlags().String("aws-access-key-id", "", "The ID of the access key to use")
	rootCmd.PersistentFlags().String("aws-secret-access-key", "", "The secret access key to use for auth")
	rootCmd.PersistentFlags().String("aws-external-id", "", "The external ID to use when assuming the customer's role")
	rootCmd.PersistentFlags().String("aws-target-role-arn", "", "The role to assume in the customer's account")
	rootCmd.PersistentFlags().String("aws-regions", "", "Comma-separated list of AWS regions that this source should operate in")
	rootCmd.PersistentFlags().BoolP("auto-config", "a", false, "Use the local AWS config, the same as the AWS CLI could use. This can be set up with \"aws configure\"")
	rootCmd.PersistentFlags().IntP("health-check-port", "", 8080, "The port that the health check should run on")

	// tracing
	rootCmd.PersistentFlags().String("honeycomb-api-key", "", "If specified, configures opentelemetry libraries to submit traces to honeycomb")
	rootCmd.PersistentFlags().String("sentry-dsn", "", "If specified, configures sentry libraries to capture errors")
	rootCmd.PersistentFlags().String("run-mode", "release", "Set the run mode for this service, 'release', 'debug' or 'test'. Defaults to 'release'.")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

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
				viper.BindPFlag(f.Name, f)
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
		if err := initTracing(tracingOpts...); err != nil {
			log.Fatal(err)
		}
	}
	// shut down tracing at the end of the process
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		shutdownTracing()
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

func getAWSConfig(strategy, region, accessKeyID, secretAccessKey, externalID, roleARN string, autoConfig bool) (aws.Config, error) {
	if autoConfig {
		return config.LoadDefaultConfig(context.Background())
	}
	// Validate inputs
	if region == "" {
		return aws.Config{}, errors.New("aws-region cannot be blank")
	}

	if strategy == "access-key" {
		if accessKeyID == "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-access-key-id cannot be blank")
		}
		if secretAccessKey == "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-secret-access-key cannot be blank")
		}
		if externalID != "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-external-id must be blank")
		}
		if roleARN != "" {
			return aws.Config{}, errors.New("with access-key strategy, aws-target-role-arn must be blank")
		}

		config := getStaticAWSConfig(region, accessKeyID, secretAccessKey)
		return config, nil
	} else if strategy == "external-id" {
		if accessKeyID != "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-access-key-id must be blank")
		}
		if secretAccessKey != "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-secret-access-key must be blank")
		}
		if externalID == "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-external-id cannot be blank")
		}
		if roleARN == "" {
			return aws.Config{}, errors.New("with external-id strategy, aws-target-role-arn cannot be blank")
		}

		return getAssumedRoleAWSConfig(region, externalID, roleARN)
	} else {
		return aws.Config{}, errors.New("invalid aws-access-strategy")
	}
}

func getAssumedRoleAWSConfig(region, externalID, targetRoleARN string) (aws.Config, error) {
	ctx := context.Background()

	assumecnf, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("could not load default config from environment: %v", err)
	}

	stsclient := sts.NewFromConfig(assumecnf)
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscredsv2.NewAssumeRoleProvider(
				stsclient,
				targetRoleARN,
				func(aro *stscredsv2.AssumeRoleOptions) {
					aro.ExternalID = &externalID
				},
			)),
		),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("could not assume the target role: %v", err)
	}
	return cfg, nil
}

func getStaticAWSConfig(region string, accessKeyID string, secretAccessKey string) aws.Config {
	return aws.Config{
		Region:      region,
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	}
}

// createTokenClient Creates a basic token client that will authenticate to NATS
// using the given values
func createTokenClient(natsJWT string, natsNKeySeed string) (connect.TokenClient, error) {
	var kp nkeys.KeyPair
	var err error

	if natsJWT == "" {
		return nil, errors.New("nats-jwt was blank. This is required when using authentication")
	}

	if natsNKeySeed == "" {
		return nil, errors.New("nats-nkey-seed was blank. This is required when using authentication")
	}

	if _, err = jwt.DecodeUserClaims(natsJWT); err != nil {
		return nil, fmt.Errorf("could not parse nats-jwt: %v", err)
	}

	if kp, err = nkeys.FromSeed([]byte(natsNKeySeed)); err != nil {
		return nil, fmt.Errorf("could not parse nats-nkey-seed: %v", err)
	}

	return connect.NewBasicTokenClient(natsJWT, kp), nil
}

// TerminationLogHook A hook that logs fatal errors to the termination log
type TerminationLogHook struct{}

func (t TerminationLogHook) Levels() []log.Level {
	return []log.Level{log.FatalLevel}
}

func (t TerminationLogHook) Fire(e *log.Entry) error {
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
