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
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/netflix/weep/util"

	"github.com/mattn/go-isatty"

	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:               "weep",
		Short:             "weep helps you get the most out of ConsoleMe credentials",
		Long:              "Weep is a CLI tool that manages AWS access via ConsoleMe for local development.",
		DisableAutoGenTag: true,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.weep.yaml)")
	rootCmd.PersistentFlags().StringSliceVarP(&assumeRole, "assume-role", "A", make([]string, 0), "one or more roles to assume after retrieving credentials")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "log format (json or tty)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn)")
	rootCmd.PersistentFlags().BoolVarP(&runAsService, "svc", "s", false, "run weep as a service")
}

func Execute() {
	shutdown := make(chan os.Signal, 1)
	done = make(chan int, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	if runAsService {
		RunService()
	} else {
		if err := rootCmd.Execute(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName(".weep")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath(home + "/.config/weep/")
	}

	if err := config.ReadEmbeddedConfig(); err != nil {
		log.Debugf("unable to read embedded config: %v; falling back to config file", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && config.EmbeddedConfigFile != "" {
			log.Debugf("no config file found, trying to use embedded config")
		} else if isatty.IsTerminal(os.Stdout.Fd()) {
			err = util.FirstRunPrompt()
			if err != nil {
				log.Fatalf("config bootstrap failed: %v", err)
			}
		} else {
			log.Debugf("unable to read config file: %v", err)
		}
	}

	log.Debugf("found config at %s", viper.ConfigFileUsed())
	if err := viper.Unmarshal(&config.Config); err != nil {
		log.Fatalf("unable to decode config into struct: %v", err)
	}
}

func initLogging() {
	// Set the log format.  Default to Text
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}

	// Set the log level and default to INFO
	switch logLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
