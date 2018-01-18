package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of triton-kubernetes",
	Long:  `All software has versions. This is triton-kubernetes's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("triton-kubernetes v0.0.1")
	},
}
