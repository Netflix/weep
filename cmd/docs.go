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
	"github.com/netflix/weep/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docCommand = &cobra.Command{
	Use:    "docs",
	Short:  docsShortHelp,
	Long:   docsLongHelp,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTree(rootCmd, "./docs/")
		if err != nil {
			logging.Log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docCommand)
}
