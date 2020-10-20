package config

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/util"
	"github.com/spf13/viper"
)

// FirstRunPrompt gets user input to bootstrap a bare-minimum configuration.
func FirstRunPrompt() error {
	fmt.Println("Welcome to weep, the ConsoleMe CLI!")
	fmt.Println("We're going to ask a few questions to get you set up.")
	fmt.Println("Or, if you would prefer, you can manually create a config file.")
	fmt.Println("Learn more here: https://github.com/Netflix/weep#configuration")
	cmURL, err := promptConsoleMeURL()
	if err != nil {
		return err
	}
	viper.Set("consoleme_url", cmURL)

	authMethod, err := promptAuthMethod()
	if err != nil {
		return err
	}
	viper.Set("authentication_method", authMethod)

	if authMethod == "mtls" {
		cert, err := promptFilePath("mTLS certificate path", "")
		if err != nil {
			return err
		}
		viper.Set("mtls_settings.cert", cert)

		key, err := promptFilePath("mTLS key path", "")
		if err != nil {
			return err
		}
		viper.Set("mtls_settings.key", key)

		ca, err := promptFilePath("mTLS CA bundle path", "")
		if err != nil {
			return err
		}
		viper.Set("mtls_settings.cafile", ca)

		insecure, err := promptBool("Skip validation of mTLS hostname?")
		if err != nil {
			return err
		}
		viper.Set("mtls_settings.insecure", insecure)
	} else if authMethod == "challenge" {
		challengeUser, err := promptString("ConsoleMe username")
		if err != nil {
			return err
		}
		viper.Set("challenge_settings.user", challengeUser)
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	defaultConfig := path.Join(home, ".weep.yaml")
	saveLocation, err := promptFilePathNoValidate("Config destination", defaultConfig)
	if err != nil {
		return err
	}
	err = viper.SafeWriteConfigAs(saveLocation)
	if err != nil {
		return err
	}
	return nil
}

func promptConsoleMeURL() (string, error) {
	validateURL := func(input string) error {
		_, err := url.ParseRequestURI(input)
		if err != nil {
			return errors.New("invalid URL")
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

func promptFilePath(label, defaultValue string) (string, error) {
	validateFile := func(input string) error {
		if util.FileExists(input) {
			return nil
		} else {
			return fmt.Errorf("file not found: %s", input)
		}
	}
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFile,
		Default:  defaultValue,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptFilePathNoValidate(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
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

func promptString(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}
