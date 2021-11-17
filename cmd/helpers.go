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
	"os"
	"strconv"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/netflix/weep/pkg/creds"
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
	rolesExtended, err := client.RolesExtended()
	if err != nil {
		return "", err
	}
	var roles []string
	var rolesSearch []string
	maxLen := 12
	for _, role := range rolesExtended {
		if len(role.AccountName) > maxLen {
			maxLen = len(role.AccountName)
		}
	}
	maxLenS := strconv.Itoa(maxLen)
	for _, role := range rolesExtended {
		account := role.AccountName
		if account == "Unknown" {
			account = role.AccountNumber
		}
		account = fmt.Sprintf("%-"+maxLenS+"s", account)
		roles = append(roles, account+"\t"+role.RoleName)
		// So users can search <account friendly name> <role> or <role> <account friendly name>
		rolesSearch = append(rolesSearch, role.AccountName+role.Arn+role.AccountName)
	}

	// Prompt the user
	prompt := promptui.Select{
		Label: "You can search for role name or account name/number or a combination of the two, e.g. prod appname",
		Items: roles,
		Size:  10,
		Searcher: func(input string, index int) bool {
			// filter out all spaces
			input = strings.ReplaceAll(input, " ", "")
			return fuzzy.MatchNormalizedFold(input, rolesSearch[index])
		},
		StartInSearchMode: true,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return rolesExtended[idx].Arn, nil
}

// InteractiveAccountsPrompt will present the user with a fuzzy-searchable list of accounts if
// - We are currently attached to an interactive tty
// - The user has not disabled them through the WEEP_DISABLE_INTERACTIVE_PROMPTS option
func InteractiveAccountsPrompt(query string, region string, client *creds.Client) (string, error) {

	if !isRunningInTerminal() {
		return "", fmt.Errorf("cannot prompt for input")
	}

	if os.Getenv("WEEP_DISABLE_INTERACTIVE_PROMPTS") == "1" {
		return "", fmt.Errorf("interactive prompts are disabled")
	}

	// If a client was not provided, create one using the provided region
	if client == nil {
		var err error
		client, err = creds.GetClient(region)
		if err != nil {
			return "", err
		}
	}

	// Retrieve the list of accounts
	accounts, err := client.GetAccounts(query)
	if err != nil {
		return "", err
	}

	var accountsSearchString []string
	var accountsDisplay []string
	maxLen := 12
	for _, account := range accounts {
		if len(account.AccountName) > maxLen {
			maxLen = len(account.AccountName)
		}
	}
	maxLenS := strconv.Itoa(maxLen)
	for _, account := range accounts {
		account.AccountName = fmt.Sprintf("%-"+maxLenS+"s", account.AccountName)
		accountsDisplay = append(accountsDisplay, account.AccountName+"\t"+account.AccountNumber)
		// So users can search <account friendly name> <role> or <role> <account friendly name>
		accountsSearchString = append(accountsSearchString, account.AccountName+account.AccountNumber+account.AccountName)
	}

	// Prompt the user
	prompt := promptui.Select{
		Label: "You can search for account name or number or a combination of the two, e.g. aws 123",
		Items: accountsDisplay,
		Size:  10,
		Searcher: func(input string, index int) bool {
			// filter out all spaces
			input = strings.ReplaceAll(input, " ", "")
			return fuzzy.MatchNormalizedFold(input, accountsSearchString[index])
		},
		StartInSearchMode: true,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return accountsDisplay[idx], nil
}

func isRunningInTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
