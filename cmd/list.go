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

	"github.com/netflix/weep/creds"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.PersistentFlags().BoolVarP(&showAll, "all", "a", false, "include all roles (console and application)")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: listShortHelp,
	Long:  listLongHelp,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := creds.GetClient(region)
	if err != nil {
		return err
	}
	roles, err := client.Roles(showAll)
	if err != nil {
		return err
	}
	fmt.Println("Roles:")
	for i := range roles {
		fmt.Println("  ", roles[i])
	}
	return nil
}
