package cmd

import (
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/spf13/cobra"
	"os"
)

func createRunnerJobPool() *runners.JobPool {
	return runners.NewJobPool(&ssh.KeyInstaller{})
}

func init() {
	rootCmd.AddCommand(runnerCmd)
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}
