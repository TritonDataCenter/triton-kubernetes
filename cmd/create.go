package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/create"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:       "create [manager or cluster or node]",
	Short:     "Create cluster managers, kubernetes clusters or individual kubernetes cluster nodes.",
	Long:      `Create allows you to create a new cluster manager or a new kubernetes cluster or an individual kubernetes cluster node.`,
	ValidArgs: []string{"manager", "cluster", "node"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(`"triton-kubernetes create" requires one argument`)
		}

		for _, validArg := range cmd.ValidArgs {
			if validArg == args[0] {
				return nil
			}
		}

		return fmt.Errorf(`invalid argument "%s" for "triton-kubernetes create"`, args[0])
	},
	Run: createCmdFunc,
}

func createCmdFunc(cmd *cobra.Command, args []string) {
	// Get silent mode value
	silentMode, err := cmd.Flags().GetBool("silent")
	if err != nil {
		silentMode = false
	}
	if silentMode {
		fmt.Println("Running in silent mode")
	}

	remoteBackend, err := util.PromptForBackend(silentMode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	createType := args[0]
	switch createType {
	case "manager":
		fmt.Println("create manager called")
		err := create.NewTritonManager(remoteBackend, silentMode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "cluster":
		fmt.Println("create cluster called")
		err := create.NewCluster(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "node":
		fmt.Println("create node called")
		err := create.NewNode(remoteBackend, silentMode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(createCmd)

	// createCmd.AddCommand(...)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
