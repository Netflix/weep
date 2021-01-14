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
	Short: "Write all of your eligible roles as profiles in your AWS Config to source credentials from Weep",
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
	client, err := creds.GetClient()
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
	credentials, err := creds.GetCredentials(role, noIpRestrict, assumeRole)
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
