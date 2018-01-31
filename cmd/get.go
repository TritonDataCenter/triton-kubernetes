package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/get"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:       "get [manager or cluster]",
	Short:     "Display resource information",
	Long:      `Get allows you to get cluster manager details.`,
	ValidArgs: []string{"manager", "cluster"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(`"triton-kubernetes get" requires one argument`)
		}

		for _, validArg := range cmd.ValidArgs {
			if validArg == args[0] {
				return nil
			}
		}

		return fmt.Errorf(`invalid argument "%s" for "triton-kubernetes get"`, args[0])
	},
	Run: getCmdFunc,
}

func getCmdFunc(cmd *cobra.Command, args []string) {
	// Get silent mode value
	silentMode, err := cmd.Flags().GetBool("silent")
	if err != nil {
		silentMode = false
	}
	fmt.Printf("Silent Mode: %v\n", silentMode)

	remoteBackend, err := util.PromptForBackend(silentMode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	getType := args[0]
	switch getType {
	case "manager":
		fmt.Println("get manager called")
		err := get.GetManager(remoteBackend, silentMode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "cluster":
		fmt.Println("get cluster called")
		err := get.GetCluster(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
