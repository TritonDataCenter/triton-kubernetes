package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "triton-kubernetes",
	Short: "A tool to deploy Kubernetes to multiple cloud providers",
	Long: `This is a multi-cloud Kubernetes solution. Triton Kubernetes has a global
cluster manager which can manage multiple clusters across regions/data-centers and/or clouds. 
Cluster manager can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.
For an example set up, look at the How-To section.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.triton-kubernetes.yaml)")
	rootCmd.PersistentFlags().Bool("non-interactive", false, "Prevent interactive prompts")
	rootCmd.PersistentFlags().Bool("terraform-configuration", false, "Create terraform configuration only")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.BindPFlag("non-interactive", rootCmd.Flags().Lookup("non-interactive"))
	if viper.GetBool("non-interactive") {
		fmt.Println("Running in non interactive mode")
	}
	viper.BindPFlag("terraform-configuration", rootCmd.Flags().Lookup("terraform-configuration"))
	if viper.GetBool("terraform-configuration") {
		fmt.Println("Will not create infrastructure, only terraform configuration")
	}
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".triton-kubernetes") // name of config file (without extension)
		viper.AddConfigPath("$HOME")              // adding home directory as first search path
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
