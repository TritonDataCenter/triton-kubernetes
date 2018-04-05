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
		fmt.Printf("triton-kubernetes v0.8.2 (%s)\n", GitHash)
	},
}
