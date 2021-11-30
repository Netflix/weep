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
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/netflix/weep/pkg/logging"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

func init() {
	// Set default configuration values here
	viper.SetTypeByDefaultValue(true)
	viper.SetDefault("authentication_method", "challenge")
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("feature_flags.consoleme_metadata", false)
	viper.SetDefault("log_file", getDefaultLogFile())
	viper.SetDefault("mtls_settings.old_cert_message", "mTLS certificate is too old, please refresh mtls certificate")
	viper.SetDefault("server.enforce_imdsv2", false)
	viper.SetDefault("server.http_timeout", 20)
	viper.SetDefault("server.address", "127.0.0.1")
	viper.SetDefault("server.port", 9091)
	viper.SetDefault("service.command", "serve")
	viper.SetDefault("service.run", []string{"service", "run"})
	viper.SetDefault("service.args", []string{})
	viper.SetDefault("service.flags", []string{})
	viper.SetDefault("swag.enable", false)
	viper.SetDefault("swag.use_mtls", false)
	viper.SetDefault("swag.url", "")

	// Set aliases for backward-compatibility
	viper.RegisterAlias("server.ecs_credential_provider_port", "server.port")
}

func getDefaultLogFile() string {
	switch os := runtime.GOOS; os {
	case "darwin":
		return filepath.Join("/", "tmp", "weep.log")
	case "linux":
		return filepath.Join("/", "tmp", "weep.log")
	case "windows":
		p, _ := filepath.Abs(filepath.FromSlash("/programdata/weep/weep.log"))
		return p
	default:
		return ""
	}
}

// initConfig reads in configs by precedence, with later configs overriding earlier:
//   - embedded
//   - /etc/weep/weep.yaml
//   - ~/.weep/weep.yaml
//   - ./weep.yaml
// If a config file is specified via CLI arg, it will be read exclusively and not merged with other
// configuration.
func InitConfig(filename string) error {
	viper.SetConfigType("yaml")

	// Read in explicitly defined config file
	if filename != "" {
		viper.SetConfigFile(filename)
		if err := viper.ReadInConfig(); err != nil {
			logging.Log.Errorf("could not open config file %s: %v", filename, err)
			return err
		}
		return nil
	}

	// Read embedded config if available
	if err := ReadEmbeddedConfig(); err != nil {
		logging.Log.Debugf("unable to read embedded config: %v", err)
	}

	configLocations := []string{
		"/etc/weep",
		"$HOME/.weep",
		".",
	}

	for _, dir := range configLocations {
		viper.SetConfigName("weep")
		viper.AddConfigPath(dir)
		_ = viper.MergeInConfig()
	}

	// TODO: revisit first-run setup
	//if err := viper.MergeInConfig(); err != nil {
	//	if _, ok := err.(viper.ConfigFileNotFoundError); ok && config.EmbeddedConfigFile != "" {
	//		logging.Log.Debugf("no config file found, trying to use embedded config")
	//	} else if isatty.IsTerminal(os.Stdout.Fd()) {
	//		err = util.FirstRunPrompt()
	//		if err != nil {
	//			logging.Log.Fatalf("config bootstrap failed: %v", err)
	//		}
	//	} else {
	//		logging.Log.Debugf("unable to read config file: %v", err)
	//	}
	//}

	if err := viper.Unmarshal(&Config); err != nil {
		return errors.Wrap(err, "unable to decode config into struct")
	}
	return nil
}

// SetUser saves the provided username to ~/.weep/weep.yaml
func SetUser(user string) error {
	// Create a temporary viper instance to isolate from main config
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("weep")
	v.AddConfigPath("$HOME/.weep")

	// Read in existing config if there is one so we don't overwrite it
	if err := v.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			break
		default:
			return err
		}
	}

	// Set user in the temp config then write to file
	v.Set("challenge_settings.user", user)

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	configPath := path.Join(home, ".weep/weep.yaml")

	if err := v.WriteConfigAs(configPath); err != nil {
		return err
	}
	return nil
}

func MtlsEnabled() bool {
	authMethod := viper.GetString("authentication_method")
	return authMethod == "mtls"
}

// BaseWebURL allows the ConsoleMe URL to be overridden for cases where the API
// and UI are accessed via different URLs
func BaseWebURL() string {
	if override := viper.GetString("consoleme_open_url_override"); override != "" {
		return override
	}
	return viper.GetString("consoleme_url")
}

func MergeExtraConfigFile(extraConfigFile string) error {
	f, err := os.Open(extraConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = viper.MergeConfig(f); err != nil {
		return errors.Wrap(err, "could not merge extra config")
	}

	return nil
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
