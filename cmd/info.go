package cmd

import (
	"fmt"
	"io"

	"github.com/netflix/weep/metadata"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:    "info",
	Short:  infoShortHelp,
	Long:   infoLongHelp,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		PrintWeepInfo(cmd.OutOrStderr())
	},
}

func PrintWeepInfo(w io.Writer) {
	encoder := yaml.NewEncoder(w)

	roles, err := roleList(true)
	if err != nil {
		fmt.Fprintln(w, "Failed to retrieve role list from ConsoleMe:")
		fmt.Fprintln(w, err)
	}
	fmt.Fprintln(w, roles)

	fmt.Fprintln(w, "Version")
	encoder.Encode(metadata.GetVersion())
	fmt.Fprint(w, "\n")

	fmt.Fprintln(w, "Configuration")
	encoder.Encode(viper.AllSettings())
	fmt.Fprint(w, "\n")

	fmt.Fprintln(w, "Host Info")
	encoder.Encode(metadata.GetInstanceInfo())

}
