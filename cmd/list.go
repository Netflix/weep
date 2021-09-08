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
	"os"
	"strconv"

	"github.com/netflix/weep/pkg/creds"
	"github.com/netflix/weep/pkg/util"
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

func roleList() (string, error) {
	client, err := creds.GetClient(region)
	if err != nil {
		return "", err
	}
	roles, err := client.Roles()
	if err != nil {
		return "", err
	}
	var rolesData [][]string
	// TODO: once corresponding ConsoleMe PR is merged, update to use actual data
	for i, role := range roles {
		rolesData = append(rolesData, []string{role, strconv.Itoa((i + 1) * 12345677), "Unknown"})
	}
	// TODO: once corresponding ConsoleMe PR is merged, update to use actual headers
	headers := []string{"Role ARN", "Account ID", "Account Friendly name"}
	rolesString := util.RenderTabularData(headers, rolesData)
	return rolesString, nil
}

func runList(cmd *cobra.Command, args []string) error {
	rolesData, err := roleList()
	if err != nil {
		return err
	}
	cmd.SetOut(os.Stdout)
	cmd.Println(rolesData)
	return nil
}
