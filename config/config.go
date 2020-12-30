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
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

func init() {
	// Set default configuration values here
	viper.SetTypeByDefaultValue(true)
	viper.SetDefault("log_file", getDefaultLogFile())
	viper.SetDefault("mtls_settings.old_cert_message", "mTLS certificate is too old, please refresh mtls certificate")
	viper.SetDefault("server.http_timeout", 20)
	viper.SetDefault("server.metadata_port", 9090)
	viper.SetDefault("server.ecs_credential_provider_port", 9091)
	viper.SetDefault("service.command", "ecs_credential_provider")
	viper.SetDefault("service.args", []string{})
}

func getDefaultLogFile() string {
	switch os := runtime.GOOS; os {
	case "darwin":
		return filepath.Join("tmp", "weep.log")
	case "linux":
		return filepath.Join("tmp", "weep.log")
	case "windows":
		path, _ := filepath.Abs(filepath.FromSlash("/programdata/weep/weep.log"))
		return path
	default:
		return ""
	}
}

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
