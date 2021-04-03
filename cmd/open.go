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
	"errors"

	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/util"
	"github.com/spf13/cobra"
)

func init() {
	openCmd.PersistentFlags().BoolVarP(&noOpen, "no-open", "x", false, "don't automatically open links")
	rootCmd.AddCommand(openCmd)
}

var openCmd = &cobra.Command{
	Use:          "open <arn>",
	Short:        openShortHelp,
	Long:         openLongHelp,
	RunE:         runOpen,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
}

func runOpen(cmd *cobra.Command, args []string) error {
	arn_parsed, err := util.ArnParse(args[0])

	if err != nil {
		return err
	}
	if (arn_parsed.Service == "sns" || arn_parsed.Service == "sqs") && arn_parsed.Region == "" {
		return errors.New("Resource type sns and sqs require region in the arn")
	}
	var resourceURL string
	client, err := creds.GetClient(region)
	if err != nil {
		return err
	}
	resourceURL, err = client.GetResourceURL(args[0])
	if err != nil {
		return err
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
