package cmd

import (
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerStartCmd)
}

func runRunner() {
	util.ConfigInit(configPath, noConfig)

	taskPool := runners.JobPool{}

	taskPool.Run()
}

var runnerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
