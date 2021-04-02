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

	"github.com/netflix/weep/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	linkCmd.PersistentFlags().BoolVarP(&noOpen, "no-open", "x", false, "don't automatically open links")
	rootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:          "link <arn>",
	Short:        linkShortHelp,
	Long:         linkLongHelp,
	RunE:         runLink,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
}

func runLink(cmd *cobra.Command, args []string) error {
	arn_parsed, err := util.ArnParse(args[0])

	if err != nil {
		return err
	}
	var resourceURL string
	if arn_parsed.Service == "sns" || arn_parsed.Service == "sqs" {
		if arn_parsed.Region == "" {
			return errors.New("Region is required for sns or sqs")
		}
		resourceURL = fmt.Sprintf("%s/policies/edit/%s/%s/%s/%s", viper.GetString("consoleme_url"), arn_parsed.AccountId, arn_parsed.Service, arn_parsed.Region, arn_parsed.Resource)
	} else if arn_parsed.Service == "iam" {
		service := "iamrole"
		resourceURL = fmt.Sprintf("%s/policies/edit/%s/%s/%s", viper.GetString("consoleme_url"), arn_parsed.AccountId, service, arn_parsed.Resource)
	} else if arn_parsed.Service == "s3" {
		// TODO: s3 doesn't have an account ID, how to solve?
	} else {

	}
	if noOpen {
		cmd.Printf("ConsoleMe Link: %s\n", resourceURL)
		cmd.Print("Skipping opening link automatically, please open above link manually!\n")
	} else {
		cmd.Printf("Opening browser to link: %s\n", resourceURL)
		err = util.OpenLink(resourceURL)
		if err != nil {
			cmd.PrintErrln(err.Error())
		}
	}
	return nil
}
