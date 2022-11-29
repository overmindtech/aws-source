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
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/aws-source/sources/elasticloadbalancing"
	"github.com/overmindtech/aws-source/sources/securitygroup"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/multiconn"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aws-source",
	Short: "Remote primary source for kubernetes",
	Long: `A template for building sources.

Edit this once you have created your source
`,
	Run: func(cmd *cobra.Command, args []string) {
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

		region := viper.GetString("aws-region")
		accessKeyID := viper.GetString("aws-access-key-id")
		secretAccessKey := viper.GetString("aws-secret-access-key")
		autoConfig := viper.GetBool("auto-config")

		var natsNKeySeedLog string
		var tokenClient multiconn.TokenClient

		if natsNKeySeed != "" {
			natsNKeySeedLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":          natsServers,
			"nats-name-prefix":      natsNamePrefix,
			"nats-jwt":              natsJWT,
			"nats-nkey-seed":        natsNKeySeedLog,
			"max-parallel":          maxParallel,
			"aws-region":            region,
			"aws-access-key-id":     accessKeyID,
			"aws-secret-access-key": secretAccessKey,
			"auto-config":           autoConfig,
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

		e := discovery.Engine{
			Name: "kubernetes-source",
			NATSOptions: &multiconn.NATSConnectionOptions{
				CommonOptions: multiconn.CommonOptions{
					NumRetries: -1,
					RetryDelay: 5 * time.Second,
				},
				Servers:           natsServers,
				ConnectionName:    fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
				ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
				MaxReconnects:     -1,
				ReconnectWait:     1 * time.Second,
				ReconnectJitter:   1 * time.Second,
				TokenClient:       tokenClient,
			},
			MaxParallelExecutions: maxParallel,
		}

		// TODO: Create a way to load config for auth that will work within srcman ⚠️
		// Load config and create client which will be re-used for all connections
		configCtx, configCancel := context.WithTimeout(context.Background(), 10*time.Second)

		cfg, err := getAWSConfig(region, accessKeyID, secretAccessKey, autoConfig)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error loading config")

			os.Exit(1)
		}

		// Work out what account we're using. This will be used in item contexts
		stsClient := sts.NewFromConfig(cfg)

		var callerID *sts.GetCallerIdentityOutput

		callerID, err = stsClient.GetCallerIdentity(configCtx, &sts.GetCallerIdentityInput{})

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error retrieving account information")

			os.Exit(1)
		}

		// Cancel config load context and release resources
		configCancel()

		sources := []discovery.Source{
			&elasticloadbalancing.ELBSource{
				Config:    cfg,
				AccountID: *callerID.Account,
			},
			&elasticloadbalancing.ELBv2Source{
				Config:    cfg,
				AccountID: *callerID.Account,
			},
			&ec2.InstanceSource{
				Config:    cfg,
				AccountID: *callerID.Account,
			},
			&securitygroup.SecurityGroupSource{
				Config:    cfg,
				AccountID: *callerID.Account,
			},
		}

		e.AddSources(sources...)

		// Start HTTP server for status
		healthCheckPort := 8080
		healthCheckPath := "/healthz"

		http.HandleFunc(healthCheckPath, func(rw http.ResponseWriter, r *http.Request) {
			if e.IsNATSConnected() {
				fmt.Fprint(rw, "ok")
			} else {
				http.Error(rw, "NATS not connected", http.StatusInternalServerError)
			}
		})

		log.WithFields(log.Fields{
			"port": healthCheckPort,
			"path": healthCheckPath,
		}).Debug("Starting healthcheck server")

		go func() {
			log.Fatal(http.ListenAndServe(":8080", nil))
		}()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start HTTP server for /healthz health checks")

			os.Exit(1)
		}

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
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 10), "Max number of requests to run in parallel")

	// Custom flags for this source
	rootCmd.PersistentFlags().String("aws-region", "", "The AWS region that this source should operate in")
	rootCmd.PersistentFlags().String("aws-access-key-id", "", "The ID of the access key to use")
	rootCmd.PersistentFlags().String("aws-secret-access-key", "", "The secret access key to use for auth")
	rootCmd.PersistentFlags().BoolP("auto-config", "a", false, "Use the local AWS config, the same as the AWS CLI could use. This can be set up with \"aws configure\"")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				viper.BindPFlag(f.Name, f)
			}
		})
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

func getAWSConfig(region string, accessKeyID string, secretAccessKey string, autoConfig bool) (aws.Config, error) {
	if autoConfig {
		return config.LoadDefaultConfig(context.Background())
	}
	// Vaidate inputs
	if region == "" {
		return aws.Config{}, errors.New("aws-region cannot be blank")
	}
	if accessKeyID == "" {
		return aws.Config{}, errors.New("aws-access-key-id cannot be blank")
	}
	if secretAccessKey == "" {
		return aws.Config{}, errors.New("aws-secret-access-key cannot be blank")
	}

	config := aws.Config{
		Region:      region,
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	}

	return config, nil
}

// createTokenClient Creates a basic token client that will authenticate to NATS
// using the given values
func createTokenClient(natsJWT string, natsNKeySeed string) (multiconn.TokenClient, error) {
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

	return multiconn.NewBasicTokenClient(natsJWT, kp), nil
}
