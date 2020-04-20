package cmd

import (
	"k8spolicy/internal"
	"os"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run's all configured rules against the manifests to test",
	Run: func(cmd *cobra.Command, args []string) {
		skipPolicies, _ := cmd.Flags().GetBool("skip-policy-download")
		skipConftest, _ := cmd.Flags().GetBool("skip-conftest-download")

		internal.DownloadPolicies(skipPolicies || os.Getenv("K8SPOLICY_SKIP_POLICY_DOWNLOAD") == "true")
		internal.DownloadCharts()
		internal.RunConftest(skipConftest || os.Getenv("K8SPOLICY_SKIP_CONFTEST_DOWNLOAD") == "true")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolP("skip-conftest-download", "", false, "Do not download the conftest binary")
	testCmd.Flags().BoolP("skip-policy-download", "", false, "Do not download the policy files")
}
