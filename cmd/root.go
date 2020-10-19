package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/mattn/go-isatty"
	"github.com/netflix/weep/util"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

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

	if err := config.ReadEmbeddedConfig(); err != nil {
		log.Errorf("unable to read embedded config: %v; falling back to config file", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && config.EmbeddedConfigFile != "" {
			log.Debug("no config file found, trying to use embedded config")
		} else if isatty.IsTerminal(os.Stdout.Fd()) {
			firstRun()
		} else {
			log.Fatalf("unable to read config file: %v", err)
		}
	}

	log.Debug("Found config")
	if err := viper.Unmarshal(&config.Config); err != nil {
		log.Fatalf("unable to decode config into struct: %v", err)
	}
}

func firstRun() {
	fmt.Println("Welcome to weep! It looks like this is your first time running.")
	fmt.Println("We're going to ask a few questions to get you set up.")
	fmt.Println("Or, if you would prefer, you can manually create a config file.")
	fmt.Println("Learn more here: https://github.com/Netflix/weep#configuration")
	cmURL, err := promptConsoleMeURL()
	if err != nil {
		// TODO: handle error
	}
	viper.Set("consoleme_url", cmURL)

	authMethod, err := promptAuthMethod()
	if err != nil {
		// TODO: handle error
	}
	viper.Set("authentication_method", authMethod)

	if authMethod == "mtls" {
		cert, err := promptFilePath("mTLS certificate path", "")
		if err != nil {
			// TODO: handle error
		}
		viper.Set("mtls_settings.cert", cert)

		key, err := promptFilePath("mTLS key path", "")
		if err != nil {
			// TODO: handle error
		}
		viper.Set("mtls_settings.key", key)

		ca, err := promptFilePath("mTLS CA bundle path", "")
		if err != nil {
			// TODO: handle error
		}
		viper.Set("mtls_settings.cafile", ca)

		insecure, err := promptBool("Skip validation of mTLS hostname?")
		if err != nil {
			// TODO: handle error
		}
		viper.Set("mtls_settings.insecure", insecure)
	} else if authMethod == "challenge" {

	}

	home, err := homedir.Dir()
	if err != nil {
		// TODO: handle error
	}
	defaultConfig := path.Join(home, ".weep.yaml")
	saveLocation, err := promptFilePathNoValidate("Config destination", defaultConfig)
	err = viper.SafeWriteConfigAs(saveLocation)
	if err != nil {
		// TODO: handle error
	}
}

func promptConsoleMeURL() (string, error) {
	validateURL := func(input string) error {
		_, err := url.ParseRequestURI(input)
		if err != nil {
			return errors.New("Invalid URL")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "ConsoleMe URL",
		Validate: validateURL,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptAuthMethod() (string, error) {
	prompt := promptui.Select{
		Label: "Authentication method",
		Items: []string{"challenge", "mtls"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptFilePath(label, default_value string) (string, error) {
	validateFile := func(input string) error {
		if util.FileExists(input) {
			return nil
		} else {
			return fmt.Errorf("File not found: %s", input)
		}
	}
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFile,
		Default:  default_value,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptFilePathNoValidate(label, default_value string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: default_value,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptBool(label string) (bool, error) {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"true", "false"},
	}

	index, _, err := prompt.Run()

	if err != nil {
		return false, err
	}

	return index == 0, nil
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
