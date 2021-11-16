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

package errors

import (
	"fmt"

	"github.com/spf13/viper"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	NoCredentialsFoundInCache      = Error("no credentials found in cache")
	NoTokenFoundInCache            = Error("no token found in cache")
	NoDefaultRoleSet               = Error("no default role set")
	BrowserOpenError               = Error("could not launch browser, open link manually")
	CredentialRetrievalError       = Error("failed to retrieve credentials from broker")
	InvalidJWT                     = Error("JWT is invalid")
	InvalidArn                     = Error("requested ARN is invalid")
	MutualTLSCertNeedsRefreshError = Error("mTLS cert needs to be refreshed")
	MultipleMatchingRoles          = Error("more than one matching role for search string")
	NoMatchingRoles                = Error("no matching roles for search string")
	MalformedRequestError          = Error("malformed request sent to broker")
	UnexpectedResponseType         = Error("received an unexpected response type")
)

func HandleError(err error) {
	if !viper.GetBool("errors.custom_messages_enabled") || viper.GetString("errors.base_help_url") == "" {
		return
	}
	base_url := viper.GetString("errors.base_help_url")
	current_error_url := base_url + viper.GetString("errors.help_url_suffix.default")
	switch err {
	case NoMatchingRoles:
		current_error_url = base_url + viper.GetString("errors.help_url_suffix.no_matching_roles")
		fmt.Println("It looks like you are missing access to that role.")
	case MutualTLSCertNeedsRefreshError:
		current_error_url = ""
		if viper.GetString("mtls_settings.refresh_command") != "" {
			fmt.Println("Refreshing your MTLS certificate now...")
			// TODO: figure out best way to do this part
		} else {
			fmt.Println("mtls_settings.old_cert_message")
		}
	case MultipleMatchingRoles:
		//TODO
	}

	if current_error_url != "" {
		fmt.Printf("Please visit %s for help with your error.\n", current_error_url)
	}
}
