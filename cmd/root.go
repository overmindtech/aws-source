package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/source-template/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "source-template",
	Short: "Remote primary source for kubernetes",
	Long: `A template for building sources.

Edit this once you have created your source
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := viper.GetString("nats-name-prefix")
		natsCAFile := viper.GetString("nats-ca-file")
		natsJWTFile := viper.GetString("nats-jwt-file")
		natsNKeyFile := viper.GetString("nats-nkey-file")
		maxParallel := viper.GetInt("max-parallel")
		hostname, err := os.Hostname()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not determine hostname for use in NATS connection name")

			os.Exit(1)
		}

		// ⚠️ Your custom configration goes here
		yourCustomFlag := viper.GetString("your-custom-flag")

		log.WithFields(log.Fields{
			"nats-servers":     natsServers,
			"nats-name-prefix": natsNamePrefix,
			"nats-ca-file":     natsCAFile,
			"nats-jwt-file":    natsJWTFile,
			"nats-nkey-file":   natsNKeyFile,
			"max-parallel":     maxParallel,
			"your-custom-flag": yourCustomFlag,
		}).Info("Got config")

		e := discovery.Engine{
			Name: "kubernetes-source",
			NATSOptions: &discovery.NATSOptions{
				URLs:           natsServers,
				ConnectionName: fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
				ConnectTimeout: (10 * time.Second), // TODO: Make configurable
				NumRetries:     999,                // We are in a container so wait forever
				CAFile:         natsCAFile,
				NkeyFile:       natsNKeyFile,
				JWTFile:        natsJWTFile,
			},
			MaxParallelExecutions: maxParallel,
		}

		// ⚠️ Here is where you add your sources
		colourNameSource := sources.ColourNameSource{}

		e.AddSources(&colourNameSource)

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

		err = e.Connect()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not connect to NATS")

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
	rootCmd.PersistentFlags().String("nats-ca-file", "", "Path to the CA file that NATS should use when connecting over TLS")
	rootCmd.PersistentFlags().String("nats-jwt-file", "", "Path to the file containing the user JWT")
	rootCmd.PersistentFlags().String("nats-nkey-file", "", "Path to the file containing the NKey seed")
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 2), "Max number of requests to run in parallel")

	// ⚠️ Add your own custom config options below, the example "your-custom-flag"
	// should be replaced with your own config or deleted
	rootCmd.PersistentFlags().String("your-custom-flag", "someDefaultValue.conf", "Description of what your option is meant to do")

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
