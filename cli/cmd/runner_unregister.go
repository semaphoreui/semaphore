package cmd

import (
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerUnregisterCmd)
}

func unregisterRunner() {
	util.ConfigInit(configPath, noConfig)

	taskPool := runners.JobPool{}
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
