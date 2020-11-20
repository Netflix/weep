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
	"fmt"
	"os"
	"path"

	ini "gopkg.in/ini.v1"

	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/consoleme"
	"github.com/netflix/weep/util"
	"github.com/spf13/cobra"
)

func init() {
	fileCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	fileCmd.PersistentFlags().StringVarP(&destination, "output", "o", getDefaultCredentialsFile(), "output file for credentials")
	fileCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "default", "profile name")
	fileCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "overwrite existing profile without prompting")
	rootCmd.AddCommand(fileCmd)
}

var fileCmd = &cobra.Command{
	Use:   "file [role_name]",
	Short: "Retrieve credentials and save them to a credentials file",
	Args:  cobra.ExactArgs(1),
	RunE:  runFile,
}

func runFile(cmd *cobra.Command, args []string) error {
	role = args[0]
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	credentials, err := client.GetRoleCredentials(role, noIpRestrict)
	if err != nil {
		return err
	}
	err = writeCredentialsFile(credentials)
	if err != nil {
		return err
	}
	return nil
}

func getDefaultCredentialsFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Printf("couldn't get default directory!")
		os.Exit(1)
	}
	return path.Join(home, ".aws", "credentials")
}

func getDefaultAwsConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Printf("couldn't get default directory!")
		os.Exit(1)
	}
	return path.Join(home, ".aws", "config")
}

func shouldOverwriteCredentials() bool {
	if force {
		return true
	}
	userForce, err := util.PromptBool(fmt.Sprintf("Overwrite %s profile?", profileName))
	if err != nil {
		return false
	}
	return userForce
}

func writeCredentialsFile(credentials consoleme.AwsCredentials) error {
	var credentialsINI *ini.File
	var err error

	// Disable pretty format, but still put spaces around `=`
	ini.PrettyFormat = false
	ini.PrettyEqual = true

	if util.FileExists(destination) {
		credentialsINI, err = ini.Load(destination)
		if err != nil {
			return err
		}
	} else {
		credentialsINI = ini.Empty()
	}

	if _, err := credentialsINI.GetSection(profileName); err == nil {
		// section already exists, should we overwrite?
		if !shouldOverwriteCredentials() {
			// user says no, so we'll just bail out
			return fmt.Errorf("not overwriting %s profile", profileName)
		}
	}

	credentialsINI.Section(profileName).Key("aws_access_key_id").SetValue(credentials.AccessKeyId)
	credentialsINI.Section(profileName).Key("aws_secret_access_key").SetValue(credentials.SecretAccessKey)
	credentialsINI.Section(profileName).Key("aws_session_token").SetValue(credentials.SessionToken)
	err = credentialsINI.SaveTo(destination)
	if err != nil {
		return err
	}

	return nil
}
