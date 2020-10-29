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

var (
	Config WeepConfig
)

type MetaDataPath struct {
	Path string `mapstructure:"path"`
	Data string `mapstructure:"data"`
}

type MetaDataConfig struct {
	Routes []MetaDataPath `mapstructure:"routes"`
}

type MtlsSettings struct {
	Cert     string   `mapstructure:"cert"`
	Key      string   `mapstructure:"key"`
	CATrust  string   `mapstructure:"catrust"`
	Insecure bool     `mapstructure:"insecure"`
	Darwin   []string `mapstructure:"darwin"`
	Linux    []string `mapstructure:"linux"`
	Windows  []string `mapstructure:"windows"`
}

type ChallengeSettings struct {
	User string `mapstructure:"user"`
}

type WeepConfig struct {
	MetaData             MetaDataConfig    `mapstructure:"metadata"`
	ConsoleMeUrl         string            `mapstructure:"consoleme_url"`
	MtlsSettings         MtlsSettings      `mapstructure:"mtls_settings"`
	ChallengeSettings    ChallengeSettings `mapstructure:"challenge_settings"`
	AuthenticationMethod string            `mapstructure:"authentication_method"`
}
