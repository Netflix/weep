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

	"github.com/netflix/weep/pkg/logging"
	"github.com/sirupsen/logrus"

	"github.com/netflix/weep/pkg/aws"

	"github.com/netflix/weep/pkg/creds"

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
	Args:  cobra.MaximumNArgs(1),
	RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
	// If a role was provided, use it, otherwise prompt
	role, err := InteractiveRolePrompt(args, region, nil)
	if err != nil {
		logging.LogError(err, "Error getting role")
		return err
	}
	logging.Log.WithFields(logrus.Fields{"role": role}).Infoln("Getting credentials")
	credentials, err := creds.GetCredentials(role, noIpRestrict, assumeRole, "")
	if err != nil {
		logging.LogError(err, "Error getting credentials")
		return err
	}
	if !useShellFlag {
		// user hasn't explicitly passed in a shell
		printExport(credentials)
	} else {
		printExportForShell(shellInfo, credentials)
	}
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

// User hasn't specified a shell, attempt to guess it
func printExport(creds *aws.Credentials) {
	if isFish() {
		// fish has a different way of setting variables than bash/zsh and others
		printExportForShell("fish", creds)
	} else {
		// defaults to bash
		printExportForShell("bash", creds)
	}
}

// Prints out the export command for a specific shell, as defined by user
func printExportForShell(shell string, creds *aws.Credentials) {
	fmt.Println(exportVar(shell, "AWS_ACCESS_KEY_ID", creds.AccessKeyId))
	fmt.Println(exportVar(shell, "AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey))
	fmt.Println(exportVar(shell, "AWS_SESSION_TOKEN", creds.SessionToken))
}

func exportVar(shell, name, value string) string {
	switch shell {
	case "fish":
		return fmt.Sprintf("set -gx %s %q;", name, value)
	case "csh", "tcsh":
		return fmt.Sprintf("setenv %s %q;", name, value)
	default: // "sh", "bash", "ksh", "zsh":
		return fmt.Sprintf("export %s=%q", name, value)
	}
}
