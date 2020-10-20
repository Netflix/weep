package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "weep",
		Short: "weep helps you get the most out of ConsoleMe credentials",
		Long:  "TBD",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.weep.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "log-format", "", "log format (json or tty)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "log-level", "", "log level (debug, info, warn)")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
		log.Errorf("unable to read embedded config: %v; falling back to config file", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && config.EmbeddedConfigFile != "" {
			log.Debugf("no config file found, trying to use embedded config")
		} else if isatty.IsTerminal(os.Stdout.Fd()) {
			err = config.FirstRunPrompt()
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
