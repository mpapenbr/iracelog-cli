/*
Copyright 2024 Markus Papenbrock
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mpapenbr/iracelog-cli/cmd/event"
	"github.com/mpapenbr/iracelog-cli/cmd/live"
	"github.com/mpapenbr/iracelog-cli/cmd/provider"
	"github.com/mpapenbr/iracelog-cli/cmd/stress"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/version"
)

const envPrefix = "iracelog-cli"

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "iracelog-cli",
	Short:   "Command line interface for iRacelog",
	Long:    ``,
	Version: version.FullVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// util.SetupLogger(config.DefaultCliArgs())
		if _, err := log.InitLoggerManager(config.DefaultCliArgs()); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing logger: %v", err)
			os.Exit(1)
		}
		fmt.Printf("iracelog-cli logLevel:%s\n", config.DefaultCliArgs().LogLevel)
	},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.iracelog-cli.yml)")
	rootCmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Addr,
		"addr", "a", "localhost:8080", "ISM gRPC address")
	rootCmd.PersistentFlags().BoolVar(&config.DefaultCliArgs().Insecure,
		"insecure", false,
		"allow insecure (non-tls) gRPC connections (used for development only)")
	rootCmd.PersistentFlags().StringVar(&config.DefaultCliArgs().LogConfig,
		"log-config",
		"logger.yml",
		"configuration file for logger")
	rootCmd.PersistentFlags().StringVar(&config.DefaultCliArgs().LogLevel,
		"log-level",
		"info",
		"controls the log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().StringVar(&config.DefaultCliArgs().LogFormat,
		"log-format",
		"text",
		"controls the log output format (json, text)")

	rootCmd.AddCommand(event.NewEventCmd())
	rootCmd.AddCommand(provider.NewProviderCmd())
	rootCmd.AddCommand(live.NewLiveCmd())
	rootCmd.AddCommand(stress.NewStressCmd())

	// add commands here
	// e.g. rootCmd.AddCommand(sampleCmd.NewSampleCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".iracelog-cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".iracelog-cli")
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	bindFlags(rootCmd, viper.GetViper())
	for _, cmd := range rootCmd.Commands() {
		bindFlags(cmd, viper.GetViper())
	}
}

// Bind each cobra flag to its associated viper configuration
// (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their
		// equivalent keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if err := v.BindEnv(f.Name,
				fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)); err != nil {
				fmt.Fprintf(os.Stderr, "Could not bind env var %s: %v", f.Name, err)
			}
		}
		// Apply the viper config value to the flag when the flag is not set and viper
		// has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				fmt.Fprintf(os.Stderr, "Could set flag value for %s: %v", f.Name, err)
			}
		}
	})
}
