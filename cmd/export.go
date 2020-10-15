package cmd

import (
	"fmt"
	"github.com/netflix/weep/consoleme"
	"github.com/spf13/cobra"
)

var (
	exportRole         string
	exportNoIPRestrict bool
)

func init() {
	exportCmd.PersistentFlags().StringVar(&exportRole, "role", "", "name of role")
	exportCmd.PersistentFlags().BoolVar(&exportNoIPRestrict, "no-ip", false, "remove IP restrictions")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Retrieve credentials to be exported as environment variables",
	RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	creds, err := client.GetRoleCredentials(exportRole, exportNoIPRestrict)
	if err != nil {
		return err
	}
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s && export AWS_SECRET_ACCESS_KEY=%s && export AWS_SESSION_TOKEN=%s\n",
		creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
	return nil
}
