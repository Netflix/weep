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

import "os"

var (
	assumeRole        []string
	profileName       string
	destination       string
	destinationConfig string
	force             bool
	autoRefresh       bool
	generate          bool
	noIpRestrict      bool
	showAll           bool
	region            string
	listenAddr        string
	listenPort        int
	cfgFile           string
	logLevel          string
	logFile           string
	logFormat         string
	shutdown          chan os.Signal
	done              chan int
)

var completionShortHelp = "Generate completion script"
var completionLongHelp = `Generate shell completion script for Bash, Zsh, Fish, and Powershell.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/advanced-configuration/shell-completion
`
var credentialProcessShortHelp = "Retrieve credentials on the fly via the AWS SDK"
var credentialProcessLongHelp = `The credential_process command can be used by AWS SDKs to retrieve 
credentials from Weep on the fly. The --generate flag lets you automatically
generate an AWS configuration with profiles for all of your available roles, or 
you can manually update your configuration (see the link below to learn how).

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-process
`

var docsShortHelp = "Generate Markdown docs for CLI commands"
var docsLongHelp = ``

var exportShortHelp = "Retrieve credentials to be exported as environment variables"
var exportLongHelp = `The export command retrieves credentials for a role and prints a shell command to export 
the credentials to environment variables.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-export
`

var fileShortHelp = "Retrieve credentials and save them to a credentials file"
var fileLongHelp = `The file command writes role credentials to the AWS credentials file, usually 
~/.aws/credentials. Since these credentials are static, you’ll have to re-run the command
every hour to get new credentials.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-file
`

var listShortHelp = "List available roles"
var listLongHelp = `The list command prints out all of the roles you have access to via ConsoleMe. By default,
this command will only show console roles. Use the --all flag to also include application
roles.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/list-eligible-roles
`

var serveShortHelp = "Run a local ECS Credential Provider endpoint that serves and caches credentials for roles on demand"
var serveLongHelp = `The serve command runs a local webserver that serves the /ecs/ path. When the
AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable is set to a URL, the 
AWS CLI and SDKs will use that URL to retrieve credentials. For example, if 
you want to use credentials for a role called SuperCoolRole, you could do 
something like this:

AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:9091/ecs/SuperCoolRole \
        aws sts get-caller-identity

If you just want to use a single role, use the --role argument to specify which one and it 
will be served the same way credentials are served in an EC2 instance. There’s no need
to set an environment variable for this.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-provider
`

var serviceShortHelp = "Install or control weep as a system service"
var serviceLongHelp = `EXPERIMENTAL FEATURE
The service command lets you install Weep as a service on a Linux, macOS, or Windows
system.
`

var setupShortHelp = "Print setup information"
var setupLongHelp = ``

var versionShortHelp = "Print version information"
var versionLongHelp = ``
