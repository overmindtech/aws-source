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
	"github.com/getsentry/sentry-go"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
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
	"github.com/overmindtech/aws-source/tracing"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go/auth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

		var err error

		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := viper.GetString("nats-name-prefix")
		natsJWT := viper.GetString("nats-jwt")
		natsNKeySeed := viper.GetString("nats-nkey-seed")
		maxParallel := viper.GetInt("max-parallel")
		apiKey := viper.GetString("api-key")
		apiPath := viper.GetString("api-path")
		healthCheckPort := viper.GetInt("health-check-port")

		hostname, err := os.Hostname()
		if err != nil {
			log.WithError(err).Fatal("Could not determine hostname for use in NATS connection name")
		}

		awsAuthConfig := AwsAuthConfig{
			Strategy:        viper.GetString("aws-access-strategy"),
			AccessKeyID:     viper.GetString("aws-access-key-id"),
			SecretAccessKey: viper.GetString("aws-secret-access-key"),
			ExternalID:      viper.GetString("aws-external-id"),
			TargetRoleARN:   viper.GetString("aws-target-role-arn"),
			Profile:         viper.GetString("aws-profile"),
			AutoConfig:      viper.GetBool("auto-config"),
		}

		var regions []string
		viper.UnmarshalKey("aws-regions", &awsAuthConfig.Regions)

		var natsNKeySeedLog string
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
			"aws-access-strategy": awsAuthConfig.Strategy,
			"aws-external-id":     awsAuthConfig.ExternalID,
			"aws-target-role-arn": awsAuthConfig.TargetRoleARN,
			"aws-profile":         awsAuthConfig.Profile,
			"auto-config":         awsAuthConfig.AutoConfig,
			"health-check-port":   healthCheckPort,
		}).Info("Got config")

		// Validate the auth params and create a token client if we are using
		// auth
		var tokenClient auth.TokenClient
		if apiKey != "" {
			tokenClient, err = auth.NewAPIKeyClient(apiPath, apiKey)

			if err != nil {
				sentry.CaptureException(err)

				log.WithError(err).Fatal("Could not create API key client")
			}
		} else if natsJWT != "" || natsNKeySeed != "" {
			tokenClient, err = createTokenClient(natsJWT, natsNKeySeed)

			if err != nil {
				log.WithError(err).Fatal("Error validating NATS authentication info")
			}
		}

		natsOptions := auth.NATSOptions{
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

		e, err := InitializeAwsSourceEngine(natsOptions, awsAuthConfig, maxParallel)
		if err != nil {
			log.WithError(err).Error("Could not initialize aws source")
			return
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

			err := http.ListenAndServe(fmt.Sprintf(":%v", healthCheckPort), nil)

			log.WithError(err).WithFields(log.Fields{
				"port": healthCheckPort,
				"path": healthCheckPath,
			}).Error("Could not start HTTP server for /healthz health checks")
		}()

		err = e.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start engine")

			os.Exit(1)
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

	rootCmd.PersistentFlags().String("api-key", "", "The API key to use to authenticate to the Overmind API")
	// Support API Keys in the environment
	err := viper.BindEnv("api-key", "OVM_API_KEY", "API_KEY")
	if err != nil {
		log.WithError(err).Fatal("could not bind api key to env")
	}

	rootCmd.PersistentFlags().String("api-path", "https://api.app.overmind.tech", "The URL of the Overmind API")
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
		return nil, fmt.Errorf("could not parse nats-jwt: %v", err)
	}

	if kp, err = nkeys.FromSeed([]byte(natsNKeySeed)); err != nil {
		return nil, fmt.Errorf("could not parse nats-nkey-seed: %v", err)
	}

	return auth.NewBasicTokenClient(natsJWT, kp), nil
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

func InitializeAwsSourceEngine(natsOptions auth.NATSOptions, awsAuthConfig AwsAuthConfig, maxParallel int) (*discovery.Engine, error) {
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

		rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
		defer rateLimitCancel()

		ec2RateLimit.Start(rateLimitCtx)
		autoScalingRateLimit.Start(rateLimitCtx)
		iamRateLimit.Start(rateLimitCtx)
		directConnectRateLimit.Start(rateLimitCtx)
		networkManagerRateLimit.Start(rateLimitCtx)

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
