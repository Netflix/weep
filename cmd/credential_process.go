/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/netflix/weep/pkg/logging"

	"github.com/netflix/weep/pkg/aws"

	"github.com/netflix/weep/pkg/creds"
	"github.com/netflix/weep/pkg/util"

	"gopkg.in/ini.v1"

	"github.com/spf13/cobra"
)

func init() {
	CredentialProcessCmd.PersistentFlags().BoolVarP(&generate, "generate", "g", false, "generate ~/.aws/config with credential process config")
	CredentialProcessCmd.PersistentFlags().StringVarP(&destinationConfig, "output", "o", getDefaultAwsConfigFile(), "output file for AWS config")
	CredentialProcessCmd.PersistentFlags().BoolVarP(&prettyPrint, "pretty", "p", false, "when combined with --generate/-g, use 'account_name-role_name' format for generated profiles instead of arn")
	rootCmd.AddCommand(CredentialProcessCmd)
}

var CredentialProcessCmd = &cobra.Command{
	Use:   "credential_process [role_name]",
	Short: credentialProcessShortHelp,
	Long:  credentialProcessLongHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCredentialProcess,
}

func writeConfigFile(roles []creds.ConsolemeRolesResponse, destination string) error {
	var configINI *ini.File
	var err error

	// Disable pretty format, but still put spaces around `=`
	ini.PrettyFormat = false
	ini.PrettyEqual = true

	if util.FileExists(destination) {
		// There's an existing config file, so we'll load it in and update the existing contents
		configINI, err = ini.Load(destination)
		if err != nil {
			return err
		}
	} else {
		// Config file doesn't exist yet. Create it with the same perms as awscli
		err = util.CreateFile(destination, 0700, 0600)
		if err != nil {
			return err
		}
		configINI = ini.Empty()
	}

	replacer := strings.NewReplacer("_", "-", " ", "-")

	for _, r := range roles {
		var profileName string
		if prettyPrint {
			accountName := strings.ToLower(replacer.Replace(r.AccountName))
			roleName := strings.ToLower(replacer.Replace(r.RoleName))
			profileName = fmt.Sprintf("profile %s-%s", accountName, roleName)
		} else {
			profileName = fmt.Sprintf("profile %s", r.Arn)
		}
		command := fmt.Sprintf("weep credential_process %s", r.Arn)
		configINI.Section(profileName).Key("credential_process").SetValue(command)
	}
	err = configINI.SaveTo(destinationConfig)
	if err != nil {
		return err
	}

	return nil
}

func generateCredentialProcessConfig(destination string) error {
	if destination == "" {
		return fmt.Errorf("no destination provided")
	}
	client, err := creds.GetClient()
	if err != nil {
		return err
	}
	roles, err := client.RolesExtended()
	if err != nil {
		return err
	}
	err = writeConfigFile(roles, destination)
	if err != nil {
		return err
	}
	return nil
}

func runCredentialProcess(cmd *cobra.Command, args []string) error {
	if generate {
		logging.Log.Infoln("Generate credential_process")
		err := generateCredentialProcessConfig(destination)
		if err != nil {
			logging.LogError(err, "Error generating credential_process")
			return err
		}
		return nil
	}
	if len(args) == 0 {
		err := fmt.Errorf("role_name not provided")
		logging.LogError(err, "Error getting role")
		return err
	}
	role := args[0]
	logging.Log.WithFields(logrus.Fields{"role": role}).Infoln("Getting credentials")
	credentials, err := creds.GetCredentials(role, noIpRestrict, assumeRole, "")
	if err != nil {
		logging.LogError(err, "Error getting credentials")
		return err
	}
	return printCredentialProcess(credentials)
}

func printCredentialProcess(credentials *aws.Credentials) error {
	expirationTimeFormat := credentials.Expiration.Format(time.RFC3339)

	credentialProcessOutput := &creds.CredentialProcess{
		Version:         1,
		AccessKeyId:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      expirationTimeFormat,
	}

	b, err := json.Marshal(credentialProcessOutput)
	if err != nil {
		logging.LogError(err, "Error parsing credential response")
		return err
	}
	fmt.Printf(string(b))
	return nil
}
