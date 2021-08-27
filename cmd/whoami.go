package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/swag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:          "whoami",
	Short:        whoamiShortHelp,
	Long:         whoamiLongHelp,
	RunE:         runWhoami,
	SilenceUsage: true,
}

func runWhoami(cmd *cobra.Command, args []string) error {
	session := aws.GetSession()
	callerIdentity, err := aws.GetCallerIdentity(session)
	if err != nil {
		return err
	}
	var name string
	if viper.GetBool("swag.enable") {
		name, err = swag.AccountName(*callerIdentity.Account)
		if err != nil {
			cmd.Printf("Failed to get account info from SWAG: %v\n", err)
		}
	}
	role := roleFromArn(*callerIdentity.Arn)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintf(w, "Role:\t%s\n", role)
	if name != "" {
		fmt.Fprintf(w, "Account:\t%s (%s)\n", name, *callerIdentity.Account)
	} else {
		fmt.Fprintf(w, "Account:\t%s\n", *callerIdentity.Account)
	}
	fmt.Fprintf(w, "ARN:\t%s\n", *callerIdentity.Arn)
	fmt.Fprintf(w, "UserId:\t%s\n", *callerIdentity.UserId)
	w.Flush()

	return nil
}

func roleFromArn(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[1]
}
