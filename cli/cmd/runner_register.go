package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var runnerRegisterArgs struct {
	stdinRegistrationToken bool
}

func init() {
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.stdinRegistrationToken, "stdin-registration-token", false, "Read registration token from stdin")
	runnerCmd.AddCommand(runnerRegisterCmd)
}

func registerRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	if runnerRegisterArgs.stdinRegistrationToken {
		tokenBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		if len(tokenBytes) == 0 {
			panic("Empty token")
		}

		util.Config.Runner.RegistrationToken = strings.TrimSpace(string(tokenBytes))
	}

	taskPool := createRunnerJobPool()

	err := taskPool.Register(configFile)

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
