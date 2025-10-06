package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/Digital-Data-Co/forge/util"
	"github.com/spf13/cobra"
)

var runnerRegisterArgs struct {
	stdinAPIToken bool
}

func init() {
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.stdinAPIToken, "stdin-api-token", false, "Read API token from stdin")
	runnerCmd.AddCommand(runnerRegisterCmd)
}

func initRunnerAPIToken() {
	if !runnerRegisterArgs.stdinAPIToken {
		return
	}

	tokenBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if len(tokenBytes) == 0 {
		panic("Empty token")
	}

	util.Config.Runner.APIToken = strings.TrimSpace(string(tokenBytes))
}

func registerRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	initRunnerAPIToken()

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
