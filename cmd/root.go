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
	"os/signal"
	"syscall"

	"github.com/netflix/weep/config"
	"github.com/netflix/weep/logging"

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
	log = logging.GetLogger()
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(updateLoggingConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.weep.yaml)")
	rootCmd.PersistentFlags().StringSliceVarP(&assumeRole, "assume-role", "A", make([]string, 0), "one or more roles to assume after retrieving credentials")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "log format (json or tty)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", viper.GetString("log_file"), "log file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn)")
}

func Run(initFunctions ...func()) {
	cobra.OnInitialize(initFunctions...)
	Execute()
}

func Execute() {
	shutdown = make(chan os.Signal, 1)
	done = make(chan int, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
	}
}

func initConfig() {
	if err := config.InitConfig(cfgFile); err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}
}

// updateLoggingConfig overrides the default logging settings based on the config and CLI args
func updateLoggingConfig() {
	err := logging.UpdateConfig(logLevel, logFormat, logFile)
	if err != nil {
		log.Errorf("failed to configure logger: %v", err)
	}
}
