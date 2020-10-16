package cmd

import (
	"fmt"
	"github.com/netflix/weep/consoleme"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func init() {
	exportCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export [role_name]",
	Short: "Retrieve credentials to be exported as environment variables",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
	role = args[0]
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	creds, err := client.GetRoleCredentials(role, noIpRestrict)
	if err != nil {
		return err
	}
	printExport(creds)
	return nil
}

// isFish will try its best to identify if we're running in fish shell
func isFish() bool {
	shellVar := os.Getenv("SHELL")

	if strings.Contains(shellVar, "fish") {
		return true
	} else {
		return false
	}
}

func printExport(creds consoleme.AwsCredentials) {
	if isFish() {
		// fish has a different way of setting variables than bash/zsh and others
		fmt.Printf("set -x AWS_ACCESS_KEY_ID %s && set -x AWS_SECRET_ACCESS_KEY %s && set -x AWS_SESSION_TOKEN %s\n",
			creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
	} else {
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s && export AWS_SECRET_ACCESS_KEY=%s && export AWS_SESSION_TOKEN=%s\n",
			creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
	}
}
