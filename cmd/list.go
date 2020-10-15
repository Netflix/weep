package cmd

import (
	"fmt"
	"github.com/netflix/weep/consoleme"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available roles",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	roles, err := client.Roles()
	if err != nil {
		return err
	}
	fmt.Println("Roles:")
	for i := range roles {
		fmt.Println("  ", roles[i])
	}
	return nil
}
