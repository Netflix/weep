package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Print setup information for Weep",
	Run: func(cmd *cobra.Command, args []string) {
		PrintSetup()
	},
}
