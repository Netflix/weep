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
	"strings"

	"github.com/netflix/weep/creds"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export [role_name]",
	Short: exportShortHelp,
	Long:  exportLongHelp,
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
	role := args[0]
	creds, err := creds.GetCredentials(role, noIpRestrict, assumeRole, "")
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

func printExport(creds *creds.AwsCredentials) {
	if isFish() {
		// fish has a different way of setting variables than bash/zsh and others
		fmt.Printf("set -x AWS_ACCESS_KEY_ID %s && set -x AWS_SECRET_ACCESS_KEY %s && set -x AWS_SESSION_TOKEN %s\n",
			creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
	} else {
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s && export AWS_SECRET_ACCESS_KEY=%s && export AWS_SESSION_TOKEN=%s\n",
			creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
	}
}
