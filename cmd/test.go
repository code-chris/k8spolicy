package cmd

import (
	"k8spolicy/internal"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run's all configured rules against the manifests to test",
	Run: func(cmd *cobra.Command, args []string) {
		internal.DownloadPolicies()
		internal.DownloadCharts()
		internal.RunConftest()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
