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
	"path"
	"strconv"
	"time"

	"gopkg.in/ini.v1"

	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/util"
	"github.com/spf13/cobra"
)

func init() {
	fileCmd.PersistentFlags().StringVarP(&destination, "output", "o", getDefaultCredentialsFile(), "output file for credentials")
	fileCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "default", "profile name")
	fileCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "overwrite existing profile without prompting")
	fileCmd.PersistentFlags().BoolVarP(&autoRefresh, "refresh", "R", false, "automatically refresh credentials in file")
	rootCmd.AddCommand(fileCmd)
}

var fileCmd = &cobra.Command{
	Use:   "file [role_name]",
	Short: fileShortHelp,
	Long:  fileLongHelp,
	Args:  cobra.ExactArgs(1),
	RunE:  runFile,
}

func runFile(cmd *cobra.Command, args []string) error {
	role := args[0]
	err := updateCredentialsFile(role, profileName, destination, noIpRestrict, assumeRole)
	if err != nil {
		return err
	}
	if autoRefresh {
		log.Infof("starting automatic file refresh for %s", role)
		go fileRefresher(role, profileName, destination, noIpRestrict, assumeRole)
		<-shutdown
	}
	return nil
}

func updateCredentialsFile(role, profile, filename string, noIpRestrict bool, assumeRole []string) error {
	credentials, err := creds.GetCredentials(role, noIpRestrict, assumeRole, "")
	if err != nil {
		return err
	}
	err = writeCredentialsFile(credentials, profile, filename)
	if err != nil {
		return err
	}
	return nil
}

func fileRefresher(role, profile, filename string, noIpRestrict bool, assumeRole []string) {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-ticker.C:
			log.Debug("checking credentials")
			expiring, err := isExpiring(filename, profile, 10)
			if err != nil {
				log.Errorf("error checking credential expiration: %v", err)
			}
			if expiring {
				log.Info("credentials are expiring soon, refreshing...")
				err = updateCredentialsFile(role, profile, filename, noIpRestrict, assumeRole)
				if err != nil {
					log.Errorf("error updating credentials: %v", err)
				} else {
					log.Info("credentials refreshed!")
				}
			}
		}
	}
}

func getDefaultCredentialsFile() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("couldn't get default directory")
	}
	return path.Join(home, ".aws", "credentials")
}

func getDefaultAwsConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("couldn't get default directory")
	}
	return path.Join(home, ".aws", "config")
}

func shouldOverwriteCredentials() bool {
	if force || autoRefresh {
		return true
	}
	userForce, err := util.PromptBool(fmt.Sprintf("Overwrite %s profile?", profileName))
	if err != nil {
		return false
	}
	return userForce
}

func isExpiring(filename, profile string, thresholdMinutes int) (bool, error) {
	fileContents, err := ini.Load(filename)
	if err != nil {
		return false, err
	}
	section, err := fileContents.GetSection(profile)
	if err != nil {
		return true, err
	}
	expiration, err := section.GetKey("expiration")
	if err != nil {
		return true, err
	}
	expirationInt, err := expiration.Int64()
	if err != nil {
		return true, err
	}
	expirationTime := time.Unix(expirationInt, 0)
	diff := time.Duration(thresholdMinutes) * time.Minute
	timeUntilExpiration := expirationTime.Sub(time.Now()).Round(0)
	log.Debugf("%s until expiration, refresh threshold is %s", timeUntilExpiration, diff)
	if timeUntilExpiration < diff {
		log.Debug("will refresh")
		return true, nil
	}
	log.Debug("will not refresh")
	return false, nil
}

func writeCredentialsFile(credentials *creds.AwsCredentials, profile, filename string) error {
	var credentialsINI *ini.File
	var err error

	// Disable pretty format, but still put spaces around `=`
	ini.PrettyFormat = false
	ini.PrettyEqual = true

	if util.FileExists(filename) {
		credentialsINI, err = ini.Load(filename)
		if err != nil {
			return err
		}
	} else {
		credentialsINI = ini.Empty()
	}

	if _, err := credentialsINI.GetSection(profile); err == nil {
		// section already exists, should we overwrite?
		if !shouldOverwriteCredentials() {
			// user says no, so we'll just bail out
			return fmt.Errorf("not overwriting %s profile", profile)
		}
	}

	credentialsINI.Section(profile).Key("aws_access_key_id").SetValue(credentials.AccessKeyId)
	credentialsINI.Section(profile).Key("aws_secret_access_key").SetValue(credentials.SecretAccessKey)
	credentialsINI.Section(profile).Key("aws_session_token").SetValue(credentials.SessionToken)
	credentialsINI.Section(profile).Key("expiration").SetValue(strconv.FormatInt(credentials.Expiration.Unix(), 10))
	err = credentialsINI.SaveTo(filename)
	if err != nil {
		return err
	}

	return nil
}
