package cmd

import (
	"time"

	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var runnerStartArgs struct {
	register bool
}

func init() {
	runnerStartCmd.PersistentFlags().BoolVar(&runnerStartArgs.register, "register", false, "Register new runner if not registered")
	runnerCmd.AddCommand(runnerStartCmd)
}

func runRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	taskPool := createRunnerJobPool()

	// If --register is passed, try to register the runner if not already registered
	if runnerStartArgs.register {

		initRunnerRegistrationToken()

		if util.Config.Runner.Token == "" {

			for {
				err := taskPool.Register(configFile)

				if err == nil {
					break
				}

				time.Sleep(5 * time.Second)
			}

			_ = util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)
		}
	}

	taskPool.Run()
}

var runnerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
