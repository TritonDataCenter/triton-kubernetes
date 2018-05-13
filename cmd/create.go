package cmd

import (
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/create"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
	Long:  `Create allows you to create resources that triton-kubernetes can manage.`,
}

var createManagerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Create Manager",
	Run: func(cmd *cobra.Command, args []string) {
		remoteBackend, err := util.PromptForBackend()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = create.NewManager(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Create Kubernetes Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		remoteBackend, err := util.PromptForBackend()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = create.NewCluster(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var createNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Create Node",
	Run: func(cmd *cobra.Command, args []string) {
		remoteBackend, err := util.PromptForBackend()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = create.NewNode(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var createBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create Cluster Backup",
	Run: func(cmd *cobra.Command, args []string) {
		remoteBackend, err := util.PromptForBackend()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = create.NewBackup(remoteBackend)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.AddCommand(createManagerCmd, createClusterCmd, createNodeCmd, createBackupCmd)
}
