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
	"github.com/netflix/weep/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	loginCmd.PersistentFlags().BoolVarP(&noOpen, "no-open", "x", false, "print the link, but do not open a browser window")
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: loginShortHelp,
	Long:  loginLongHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	// If a role was provided, use it, otherwise prompt
	role, err := InteractiveRolePrompt(args, region, nil)
	if err != nil {
		return err
	}

	// Construct the URL and open/print it; default to HTTPS if not specified
	url := fmt.Sprintf("%s/role/%s", viper.GetString("consoleme_open_url_override"), role)

	if noOpen {
		cmd.Println(url)
	} else {
		err := util.OpenLink(url)
		if err != nil {
			return err
		}
	}

	return nil
}
