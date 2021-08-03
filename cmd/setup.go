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
	"embed"

	"github.com/spf13/cobra"
)

var (
	doItForMe   bool
	SetupExtras embed.FS
)

func init() {
	setupCmd.PersistentFlags().BoolVarP(&doItForMe, "write", "w", false, "install all the things (probably requires root, definitely requires trust)")
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: setupShortHelp,
	Long:  setupLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Setup(cmd, doItForMe)
		return err
	},
}
