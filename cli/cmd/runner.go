package cmd

import (
	"os"

	"github.com/semaphoreui/semaphore/pkg/ssh"
	proTasks "github.com/semaphoreui/semaphore/pro/services/tasks"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/spf13/cobra"
)

func createRunnerJobPool() *runners.JobPool {
	pool := runners.NewJobPool(&ssh.KeyInstaller{})
	// Workflow tasks run on this runner ship their outputs to the server; the
	// open-source build wires nil and captures nothing.
	pool.SetTaskOutputsCollector(proTasks.NewTaskOutputsCollector())
	return pool
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
