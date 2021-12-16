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

	"github.com/netflix/weep/pkg/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	legacyExportCmd.PersistentFlags().StringVarP(&roleRefreshARN, "role", "z", "", "role")
	legacyExportCmd.PersistentFlags().StringVar(&shellInfo, "shell", "bash", "--shell=sh|bash|ksh|zsh|fish|csh|tcsh")
	legacyCmd.AddCommand(legacyExportCmd)
}

// wrapper to allow for backwards-compatibility for commands that implement the export functionality currently
var legacyExportCmd = &cobra.Command{
	Use:    "export [profile]",
	Short:  exportShortHelp,
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			profileName = args[0]
		} else {
			profileName = "default"
		}

		if roleRefreshARN == "" {
			// roleRefreshARN is not present, have to go through aws-profiles to see if a role matches
			awsProfiles := viper.GetStringMapString("aws-profiles")
			for name, role := range awsProfiles {
				if name == profileName {
					roleRefreshARN = role
					break
				}
			}
		}
		if roleRefreshARN == "" {
			return fmt.Errorf("unable to find profile %s in 'aws-profiles' property. You can also run with -r role_name <optional_profile_name>", profileName)
		}

		argsPass := []string{roleRefreshARN}
		// explicit shell flag, rather than guessing which shell to use
		useShellFlag = true
		shells := []string{"sh", "bash", "ksh", "zsh", "fish", "csh", "tcsh"}
		if !util.StringInSlice(shellInfo, shells) {
			return fmt.Errorf("shell must be one of %s", shells)
		}
		return exportCmd.RunE(exportCmd, argsPass)
	},
}
