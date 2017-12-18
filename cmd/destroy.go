package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:       "destroy [manager or cluster or node]",
	Short:     "Destroy cluster managers, kubernetes clusters or individual kubernetes cluster nodes.",
	Long:      `Create allows you to create a new cluster manager or a new kubernetes cluster or an individual kubernetes cluster node.`,
	ValidArgs: []string{"manager", "cluster", "node"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(`"triton-kubernetes destory" requires one argument`)
		}

		for _, validArg := range cmd.ValidArgs {
			if validArg == args[0] {
				return nil
			}
		}

		return fmt.Errorf(`invalid argument "%s" for "triton-kubernetes destory"`, args[0])
	},
	Run: destroyCmdFunc,
}

func destroyCmdFunc(cmd *cobra.Command, args []string) {
	destroyType := args[0]
	switch destroyType {
	case "manager":
		fmt.Println("destroy manager called")
	case "cluster":
		fmt.Println("destroy cluster called")
	case "node":
		fmt.Println("destroy node called")
	}
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
