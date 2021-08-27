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

package config

import (
	"embed"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	EmbeddedConfigs    embed.FS
	EmbeddedConfigFile string // To be set by ldflags at compile time
)

// ReadEmbeddedConfig attempts to read the embedded mTLS config and create a tls.Config
func ReadEmbeddedConfig() error {
	if EmbeddedConfigFile == "" {
		return EmbeddedConfigDisabledError
	}
	f, err := EmbeddedConfigs.Open(EmbeddedConfigFile)
	if err != nil {
		return errors.Wrap(err, "could not open embedded config")
	}
	defer f.Close()

	v := viper.New()
	err = viper.ReadConfig(f)
	if err != nil {
		return errors.Wrap(err, "could not read embedded config")
	}
	if err = viper.MergeConfigMap(v.AllSettings()); err != nil {
		return errors.Wrap(err, "could not merge embedded config")
	}
	return nil
}
