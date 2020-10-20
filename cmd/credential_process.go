package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/netflix/weep/util"
	ini "gopkg.in/ini.v1"

	"github.com/netflix/weep/consoleme"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	CredentialProcessCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	GenerateCredentialProcessCmd.PersistentFlags().StringVarP(&destinationConfig, "output", "o", getDefaultAwsConfigFile(), "output file for AWS config")
	rootCmd.AddCommand(CredentialProcessCmd)
	rootCmd.AddCommand(GenerateCredentialProcessCmd)
}

var GenerateCredentialProcessCmd = &cobra.Command{
	Use:   "generate_credential_process_config",
	Short: "Writes all of your eligible roles as profiles in your AWS Config to source credentials from Weep",
	RunE:  runGenerateCredentialProcessConfig,
}

var CredentialProcessCmd = &cobra.Command{
	Use:   "credential_process [role_name]",
	Short: "Retrieve credentials and writes them in credential_process format",
	Args:  cobra.ExactArgs(1),
	RunE:  runCredentialProcess,
}

func writeConfigFile(roles []string) error {
	var configINI *ini.File
	var err error

	// Disable pretty format, but still put spaces around `=`
	ini.PrettyFormat = false
	ini.PrettyEqual = true

	if util.FileExists(destination) {
		configINI, err = ini.Load(destination)
		if err != nil {
			return err
		}
	} else {
		configINI = ini.Empty()
	}

	for _, r := range roles {
		profileName := fmt.Sprintf("profile %s", r)
		command := fmt.Sprintf("weep credential_process %s", r)
		configINI.Section(profileName).Key("credential_process").SetValue(command)
	}
	err = configINI.SaveTo(destinationConfig)
	if err != nil {
		return err
	}

	return nil
}

func runGenerateCredentialProcessConfig(cmd *cobra.Command, args []string) error {
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	roles, err := client.Roles()
	if err != nil {
		return err
	}
	err = writeConfigFile(roles)
	if err != nil {
		return err
	}
	return nil
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
