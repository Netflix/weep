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
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/netflix/weep/internal/creds"
	"os"
	"strings"
)

// InteractiveRolePrompt will present the user with a fuzzy-searchable list of roles if
// - We are currently attached to an interactive tty
// - The user has not disabled them through the WEEP_DISABLE_INTERACTIVE_PROMPTS option
func InteractiveRolePrompt(args []string, region string, client *creds.Client) (string, error) {
	// If a role was provided, just use that
	if len(args) > 0 {
		return args[0], nil
	}

	if !isRunningInTerminal() {
		return "", fmt.Errorf("no role provided, and cannot prompt for input")
	}

	if os.Getenv("WEEP_DISABLE_INTERACTIVE_PROMPTS") == "1" {
		return "", fmt.Errorf("no role provided, and interactive prompts are disabled")
	}

	// If a client was not provided, create one using the provided region
	if client == nil {
		var err error
		client, err = creds.GetClient(region)
		if err != nil {
			return "", err
		}
	}

	// Retrieve the list of roles
	roles, err := client.Roles()
	if err != nil {
		return "", err
	}

	// Prompt the user
	prompt := promptui.Select{
		Label: "Select Role",
		Items: roles,
		Size: 16,
		Searcher: func(input string, index int) bool {
			return fuzzy.MatchNormalized(strings.ToLower(input), strings.ToLower(roles[index]))
		},
		StartInSearchMode: true,
	}

	_, role, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return role, nil
}

func isRunningInTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
