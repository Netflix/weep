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
	"strings"

	"github.com/netflix/weep/creds"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: listShortHelp,
	Long:  listLongHelp,
	RunE:  runList,
}

func roleList(all bool) (string, error) {
	client, err := creds.GetClient(region)
	if err != nil {
		return "", err
	}
	roles, err := client.Roles()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if all {
		sb.WriteString("Available Roles\n")
	} else {
		sb.WriteString("Available Console Roles\n")
	}
	for i := range roles {
		sb.WriteString("\t")
		sb.WriteString(roles[i])
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func runList(cmd *cobra.Command, args []string) error {
	roles, err := roleList(showAll)
	if err != nil {
		return err
	}
	cmd.Print(roles)
	return nil
}
