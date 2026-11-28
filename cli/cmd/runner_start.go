package cmd

import (
	"time"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runnerStartArgs struct {
	autoRegister bool
}

func init() {
	runnerStartCmd.PersistentFlags().BoolVar(&runnerStartArgs.autoRegister, "auto-register", false, "Register new runner if not registered")
	runnerStartCmd.PersistentFlags().BoolVar(&runnerStartArgs.autoRegister, "register", false, "Alias of --auto-register")
	runnerCmd.AddCommand(runnerStartCmd)
}

func runRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	taskPool := createRunnerJobPool()

	// If --register is passed, try to register the runner if not already registered
	if runnerStartArgs.autoRegister {

		initRunnerRegistrationToken()

		if util.Config.Runner.Token == "" {

			for {
				err := taskPool.Register(configFile)

				if err == nil {
					break
				}

				log.WithFields(log.Fields{}).WithError(err).Info("Failed to register runner with token, will try again in 5 seconds")
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
