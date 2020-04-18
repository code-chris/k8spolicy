package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8spolicy/config"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8spolicy",
	Short: "Check K8s yaml-files and Helm-Charts with rego policies.",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .k8spolicy.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find current directory.
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in current directory with name ".k8spolicy.yaml".
		viper.AddConfigPath(dir)
		viper.SetConfigName(".k8spolicy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		config.Conf = &config.Config{}
		err = viper.Unmarshal(config.Conf)
		if err != nil {
			fmt.Printf("unable to decode into config struct, %v", err)
		} else {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		fmt.Println("No configfile found. Please specify one.")
		os.Exit(1)
	}
}
