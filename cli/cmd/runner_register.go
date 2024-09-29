package cmd

import (
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerRegisterCmd)
}

func registerRunner() {
	util.ConfigInit(configPath, noConfig)

	taskPool := runners.JobPool{}
	err := taskPool.Register()
	if err != nil {
		panic(err)
	}
}

var runnerRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register runner on the server",
	Run: func(cmd *cobra.Command, args []string) {
		registerRunner()
	},
}
