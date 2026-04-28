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
	name                   string
	tags                   []string
	webhook                string
	nameSet                bool
	webhookSet             bool
	tagsSet                bool
}

func init() {
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.stdinRegistrationToken, "stdin-registration-token", false, "Read registration token from stdin")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.name, "name", "", "Runner name to register with")
	runnerRegisterCmd.PersistentFlags().StringSliceVar(&runnerRegisterArgs.tags, "tags", nil, "Runner tags (comma-separated or repeat the flag)")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.webhook, "webhook", "", "Runner webhook URL")
	runnerCmd.AddCommand(runnerRegisterCmd)
}

func initRunnerRegistrationToken() {
	if !runnerRegisterArgs.stdinRegistrationToken {
		return
	}

	tokenBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if len(tokenBytes) == 0 {
		panic("Empty token")
	}

	util.Config.Runner.RegistrationToken = strings.TrimSpace(string(tokenBytes))
}

func applyRunnerRegisterFlags(cmd *cobra.Command) {
	if cmd.PersistentFlags().Changed("name") {
		util.Config.Runner.Name = runnerRegisterArgs.name
	}
	if cmd.PersistentFlags().Changed("webhook") {
		util.Config.Runner.Webhook = runnerRegisterArgs.webhook
	}
	if cmd.PersistentFlags().Changed("tags") {
		util.Config.Runner.Tags = runnerRegisterArgs.tags
	}
}

func registerRunner(cmd *cobra.Command) {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	initRunnerRegistrationToken()

	applyRunnerRegisterFlags(cmd)

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
		registerRunner(cmd)
	},
}
