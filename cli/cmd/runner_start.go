package cmd

import (
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerStartCmd)
}

func runRunner() {
	util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

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
