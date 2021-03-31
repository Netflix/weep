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
	"time"

	"github.com/netflix/weep/util"
	"gopkg.in/ini.v1"

	"github.com/netflix/weep/creds"
	"github.com/spf13/cobra"
)

func init() {
	CredentialProcessCmd.PersistentFlags().BoolVarP(&generate, "generate", "g", false, "generate ~/.aws/config with credential process config")
	CredentialProcessCmd.PersistentFlags().StringVarP(&destinationConfig, "output", "o", getDefaultAwsConfigFile(), "output file for AWS config")
	rootCmd.AddCommand(CredentialProcessCmd)
}

var CredentialProcessCmd = &cobra.Command{
	Use:   "credential_process [role_name]",
	Short: credentialProcessShortHelp,
	Long:  credentialProcessLongHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCredentialProcess,
}

func writeConfigFile(roles []string, destination string) error {
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

func generateCredentialProcessConfig(destination string) error {
	if destination == "" {
		return fmt.Errorf("no destination provided")
	}
	client, err := creds.GetClient(region)
	if err != nil {
		return err
	}
	roles, err := client.Roles()
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
		return generateCredentialProcessConfig(destination)
	}
	role := args[0]
	credentials, err := creds.GetCredentials(role, noIpRestrict, assumeRole, "")
	if err != nil {
		return err
	}
	printCredentialProcess(credentials)
	return nil
}

func printCredentialProcess(credentials *creds.AwsCredentials) {
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
		log.Error(err)
	}
	fmt.Printf(string(b))
}
