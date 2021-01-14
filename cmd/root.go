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
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/netflix/weep/config"

	"github.com/kardianos/service"

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
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", viper.GetString("log_file"), "log file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn)")
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

func initLogging() {
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

	log.Debug("configuring logging")

	// Set the log format.  Default to Text
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}

	logDir := filepath.Dir(logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		log.Debugf("attempting to create log directory %s", logDir)
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			log.Errorf("could not create log directory")
		}
	}

	var w io.Writer
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("could not open %s for logging, defaulting to stderr: %v", logFile, err)
		log.SetOutput(os.Stderr)
		w = os.Stderr
	} else if service.Interactive() {
		w = io.MultiWriter(os.Stderr, file)
	} else {
		w = file
	}
	log.SetOutput(w)
	log.Debug("logging configured")
}
