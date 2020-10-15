package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	cfgFile   string
	logLevel  string
	logFormat string

	rootCmd = &cobra.Command{
		Use:   "weep",
		Short: "weep helps you get the most out of ConsoleMe credentials",
		Long:  "TBD",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.weep.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "log_format", "", "log format (json or tty)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "log_level", "", "log level (debug, info, warn)")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	err := viper.ReadInConfig()
	if err == nil {
		log.Debug("Found config")
		err = viper.Unmarshal(&config.Config)
		if err != nil {
			log.Fatalf("unable to decode into struct, %v", err)
		}
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
