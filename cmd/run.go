package cmd

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/netflix/weep/creds"
	"github.com/spf13/cobra"
)

var (
	runRole string
)

func init() {
	runCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	runCmd.PersistentFlags().StringVarP(&runRole, "role", "r", "", "role to use when running command")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a CLI command with Weep credentials",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRun,
}

func runRun(cmd *cobra.Command, args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	credentials, err := creds.GetCredentials(runRole, noIpRestrict, assumeRole)
	if err != nil {
		return err
	}
	env := os.Environ()

	env = append(env, "AWS_ACCESS_KEY_ID="+credentials.AccessKeyId)
	env = append(env, "AWS_SECRET_ACCESS_KEY="+credentials.SecretAccessKey)
	env = append(env, "AWS_SESSION_TOKEN="+credentials.SessionToken)

	if err := syscall.Exec(binary, args[0:], env); err != nil {
		return err
	}
	return nil
}
