package cmd

import (
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

type runnerArgs struct {
	unregister bool
}

var targetRunnerArgs runnerArgs

func init() {
	runnerCmd.PersistentFlags().BoolVar(&targetRunnerArgs.unregister, "unregister", false, "Unregister runner form the server")
	rootCmd.AddCommand(runnerCmd)
}

func runRunner() {
	util.ConfigInit(configPath)

	taskPool := runners.JobPool{}

	if targetRunnerArgs.unregister {
		taskPool.Unregister()
	} else {
		taskPool.Run()
	}
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
