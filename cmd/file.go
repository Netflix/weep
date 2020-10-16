package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/consoleme"
	"github.com/netflix/weep/util"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"os"
	"path"
)

var (
	fileDestination  string
	fileNoIPRestrict bool
	fileProfileName  string
	fileRole         string
)

func init() {
	fileCmd.PersistentFlags().BoolVarP(&fileNoIPRestrict, "no-ip", "n", false, "remove IP restrictions")
	fileCmd.PersistentFlags().StringVarP(&fileDestination, "output", "o", getDefaultFile(), "output file for credentials")
	fileCmd.PersistentFlags().StringVarP(&fileProfileName, "profile", "p", "consoleme", "profile name")
	rootCmd.AddCommand(fileCmd)
}

var fileCmd = &cobra.Command{
	Use:   "file [role_name]",
	Short: "retrieve credentials and save them to a credentials file",
	Args:  cobra.ExactArgs(1),
	RunE:  runFile,
}

func runFile(cmd *cobra.Command, args []string) error {
	fileRole = args[0]
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	creds, err := client.GetRoleCredentials(fileRole, fileNoIPRestrict)
	if err != nil {
		return err
	}
	err = writeCredFile(creds)
	if err != nil {
		return err
	}
	return nil
}

func getDefaultFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Printf("couldn't get default directory!")
		os.Exit(1)
	}
	return path.Join(home, ".aws", "credentials")
}

func writeCredFile(creds consoleme.AwsCredentials) error {
	var credentialsINI *ini.File
	var err error

	if util.FileExists(fileDestination) {
		credentialsINI, err = ini.Load(fileDestination)
		if err != nil {
			return err
		}
	} else {
		credentialsINI = ini.Empty()
	}

	credentialsINI.Section(fileProfileName).Key("aws_access_key_id").SetValue(creds.AccessKeyId)
	credentialsINI.Section(fileProfileName).Key("aws_secret_access_key").SetValue(creds.SecretAccessKey)
	credentialsINI.Section(fileProfileName).Key("aws_session_token").SetValue(creds.SessionToken)
	err = credentialsINI.SaveTo(fileDestination)
	if err != nil {
		return err
	}

	return nil
}
