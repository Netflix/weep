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

	"github.com/netflix/weep/pkg/logging"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:    "search [account]",
	Short:  searchShortHelp,
	Long:   searchLongHelp,
	Hidden: true,
}

var accountSearchCmd = &cobra.Command{
	Use:   "account",
	Short: "Search for an account through ConsoleMe",
	Long:  searchLongHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) == 1 {
			query = args[0]
		}
		account, err := InteractiveAccountsPrompt(query, region, nil)
		if err != nil {
			logging.LogError(err, "Error getting account")
			return err
		}
		fmt.Println(account)
		return nil
	},
}

func init() {
	searchCmd.AddCommand(accountSearchCmd)
	rootCmd.AddCommand(searchCmd)
}
