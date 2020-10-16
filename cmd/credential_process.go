package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/netflix/weep/consoleme"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	CredentialProcessCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	rootCmd.AddCommand(CredentialProcessCmd)
}

var CredentialProcessCmd = &cobra.Command{
	Use:   "credential_process [role_name]",
	Short: "Retrieve credentials and writes them in credential_process format",
	Args:  cobra.ExactArgs(1),
	RunE:  runCredentialProcess,
}

func runCredentialProcess(cmd *cobra.Command, args []string) error {
	role = args[0]
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	creds, err := client.GetRoleCredentials(role, noIpRestrict)
	if err != nil {
		return err
	}
	printCredentialProcess(creds)
	return nil
}

func printCredentialProcess(creds consoleme.AwsCredentials) {
	expirationTimeFormat := time.Unix(creds.Expiration, 0).Format(time.RFC3339)

	credentialProcessOutput := &consoleme.CredentialProcess{
		Version:         1,
		AccessKeyId:     creds.AccessKeyId,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      expirationTimeFormat,
	}

	b, err := json.Marshal(credentialProcessOutput)
	if err != nil {
		log.Error(err)
	}
	fmt.Printf(string(b))
}
