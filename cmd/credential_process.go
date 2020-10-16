package cmd

import (
	"fmt"
	"github.com/netflix/weep/consoleme"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	CredentialProcessCmd.PersistentFlags().BoolVarP(&exportNoIPRestrict, "no-ip", "n", false, "remove IP restrictions")
	rootCmd.AddCommand(CredentialProcessCmd)
}

var CredentialProcessCmd = &cobra.Command{
	Use:   "credential_process [role_name]",
	Short: "Retrieve credentials and writes them in credential_process format",
	Args:  cobra.ExactArgs(1),
	RunE:  runCredentialProcess,
}

func runCredentialProcess(cmd *cobra.Command, args []string) error {
	exportRole = args[0]
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	creds, err := client.GetRoleCredentials(exportRole, exportNoIPRestrict)
	if err != nil {
		return err
	}
	printCredentialProcess(creds)
	return nil
}


func printCredentialProcess(creds consoleme.AwsCredentials) {
	expirationTimeFormat := time.Unix(creds.Expiration, 0).Format(time.RFC3339)

	fmt.Printf("{\n  \"Version\": 1,\n  \"AccessKeyId\": \"%s\",\n  \"SecretAccessKey\": \"%s\",\n  \"SessionToken\": \"%s\", \n  \"Expiration\": \"%s\"\n}",
		creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken, expirationTimeFormat)
}
