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

package main

import (
	"embed"
	"os"

	"github.com/netflix/weep/pkg/config"

	"github.com/netflix/weep/cmd"
)

//go:embed configs/*.yaml
var Configs embed.FS

//go:embed extras/*
var Extras embed.FS

func init() {
	cmd.SetupExtras = Extras
	config.EmbeddedConfigs = Configs
}

func main() {
	err := cmd.Execute()
	if err != nil {
		// err printing is handled by cobra
		os.Exit(1)
	}
}
