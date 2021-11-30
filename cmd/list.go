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
	"strings"

	"github.com/spf13/viper"

	"github.com/netflix/weep/pkg/logging"

	"github.com/netflix/weep/pkg/creds"
	"github.com/netflix/weep/pkg/util"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.PersistentFlags().BoolVarP(&extendedInfo, "extended-info", "e", false, "include additional information about roles such as associated apps")
	listCmd.PersistentFlags().BoolVarP(&shortInfo, "short-info", "s", false, "only display the role ARNs")
	listCmd.PersistentFlags().StringVarP(&accountFilter, "account", "a", "", "filter by aws account number or account name")
	listCmd.PersistentFlags().BoolVar(&showAll, "all", true, "show user profiles as well as instance profiles")
	listCmd.PersistentFlags().BoolVar(&showInstanceProfilesOnly, "instance", false, "show only instance roles")
	listCmd.PersistentFlags().BoolVar(&showConfiguredProfilesOnly, "profiles", false, "show only configured roles")
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
	roles, err := client.RolesExtended()
	if err != nil {
		return "", err
	}
	var rolesData [][]string

	awsProfilesARNs := make(map[string]bool)
	awsProfiles := viper.GetStringMapString("aws-profiles")
	if showConfiguredProfilesOnly {
		for _, arn := range awsProfiles {
			awsProfilesARNs[arn] = true
		}
	}
	for _, role := range roles {
		if accountFilter != "" {
			// accountFilter could be account number OR account friendly name
			if accountFilter != role.AccountNumber && accountFilter != role.AccountName {
				// matches neither, skip this role
				continue
			}
		}
		if showInstanceProfilesOnly {
			if !strings.HasSuffix(role.Arn, "InstanceProfile") {
				continue
			}
		}
		if showConfiguredProfilesOnly {
			if _, ok := awsProfilesARNs[role.Arn]; !ok {
				continue
			}
		}
		if shortInfo {
			rolesData = append(rolesData, []string{role.Arn})
			continue
		}
		curData := []string{role.AccountName, role.RoleName, role.Arn}
		if extendedInfo {
			var namesb strings.Builder
			var ownersb strings.Builder
			for _, app := range role.Apps.AppDetails {
				namesb.WriteString(app.Name)
				namesb.WriteString("\n")
				ownersb.WriteString(app.Owner)
				ownersb.WriteString("\n")
			}
			appNames := namesb.String()
			ownerNames := ownersb.String()
			if len(appNames) > 0 {
				curData = append(curData, appNames[:len(appNames)-1])
				curData = append(curData, ownerNames[:len(ownerNames)-1])
			}
		}
		rolesData = append(rolesData, curData)
	}
	var headers []string
	if shortInfo {
		headers = []string{"Role ARN"}
	} else {
		headers = []string{"Account Name", "Role Name", "Role ARN"}
		if extendedInfo {
			headers = append(headers, "App", "App Owner")
		}
	}
	rolesString := util.RenderTabularData(headers, rolesData)
	return rolesString, nil
}

func runList(cmd *cobra.Command, args []string) error {
	rolesData, err := roleList()
	if err != nil {
		logging.LogError(err, "Error generating roles for weep list")
		return err
	}
	cmd.SetOut(os.Stdout)
	cmd.Println(rolesData)
	return nil
}
