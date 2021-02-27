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

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:    "completion [bash|zsh|fish|powershell]",
	Short:  completionShortHelp,
	Long:   completionLongHelp,
	Hidden: true,
}

var bashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate Bash completion",
	Long: `To load completions:

$ source <(weep completion bash)

To load completions for each session, execute once:

Linux:

  $ weep completion bash > /etc/bash_completion.d/weep

MacOS:

  $ weep completion bash > /usr/local/etc/bash_completion.d/weep
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

var zshCompletionCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate Zsh completion",
	Long: `To load completions:

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for each session, execute once:

$ weep completion zsh > "${fpath[1]}/_weep"

You will need to start a new shell for this setup to take effect.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

var fishCompletionCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate Fish completion",
	Long: `To load completions:

$ weep completion fish | source

To load completions for each session, execute once:

$ weep completion fish > ~/.config/fish/completions/weep.fish
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
			log.Fatal(err)
		}
	},
}

var powershellCompletionCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generate Powershell completion",
	Long: `To load completions:

You will need PowerShell >= version 5.0 to make this work, in order to use PSReadLine.

If you don't have a PowerShell profile, or would like to make a new one:

Create a $PROFILE file if needed

PS> if(!(Test-Path -Path $PROFILE)) {
	New-Item -ItemType File -Path $PROFILE -Force
}

If you want to use an existing PowerShell $PROFILE file:

Open your $PROFILE file with an editor (notepad used for example)

PS> notepad $PROFILE

Add the navigable menu of the available options when you hit Tab

Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete

Now, you can utilise completion for weep:

PS> weep completion powershell | Out-String | Invoke-Expression

To load completions for every new session, execute once and source this file from your 
PowerShell profile:

PS> weep completion powershell > weep.ps1

You will need to start a new shell for this setup to take effect.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Root().GenPowerShellCompletion(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	completionCmd.AddCommand(bashCompletionCmd)
	completionCmd.AddCommand(zshCompletionCmd)
	completionCmd.AddCommand(fishCompletionCmd)
	completionCmd.AddCommand(powershellCompletionCmd)
	rootCmd.AddCommand(completionCmd)
}
