package cmd

import (
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerUnregisterCmd)
}

func unregisterRunner() {
	util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	taskPool := createRunnerJobPool()
	err := taskPool.Unregister()
	if err != nil {
		panic(err)
	}
}

var runnerUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister runner from the server",
	Run: func(cmd *cobra.Command, args []string) {
		unregisterRunner()
	},
}
