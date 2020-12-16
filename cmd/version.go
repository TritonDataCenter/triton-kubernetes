package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var cliVersion string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of triton-kubernetes",
	Long:  `All software has versions. This is triton-kubernetes's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cliVersion == "" {
			fmt.Print("no version set for this build... ")
			cliVersion = "local"
		}
		fmt.Printf("triton-kubernetes 1.0.1-pre1 (%s)\n", cliVersion)
	},
}
