package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stugotech/golog"
	"github.com/stugotech/roadrunner/server"
)

var cfgFile string
var logger = golog.NewPackageLogger()

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "roadrunner",
	Short: "Solves ACME challenges",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	pflags := RootCmd.PersistentFlags()
	pflags.StringVar(&cfgFile, "config", "", "config file")
	pflags.StringP(server.ListenKey, "l", "0.0.0.0:8080", "interface to listen for challenges on")
	pflags.String(server.PathPrefixKey, ".well-known/acme-challenge", "first component of URI path to challenges")
	pflags.String(server.StoreKey, "etcd", "KV store to use [etcd|consul|boltdb|zookeeper]")
	pflags.StringSlice(server.StoreNodesKey, []string{"127.0.0.1:2379"}, "comma-seperated list of KV (URI authority only)")
	pflags.String(server.StorePrefixKey, "coyote", "prefix to use when looking up values in KV store (will look in \"challenges\" sub path)")

	// load all flags into viper
	viper.BindPFlags(pflags)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".roadrunner") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME") // adding home directory as search path
	viper.AddConfigPath("/etc/roadrunner/")
	viper.SetEnvPrefix("roadrunner")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info("using config file", golog.String("file", viper.ConfigFileUsed()))
	}
}
