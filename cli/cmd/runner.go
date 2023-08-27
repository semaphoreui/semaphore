package cmd

import (
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runnerCmd)
}

func runRunner() {

	taskPool := runners.JobPool{}

	go taskPool.Run()
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
