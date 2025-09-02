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
	hostname               string
	enabled                bool
	disabled               bool
	projectName            string
	webhook                string
	tags                   string
}

func init() {
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.stdinRegistrationToken, "stdin-registration-token", false, "Read registration token from stdin")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.hostname, "hostname", "", "Runner hostname or name")
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.enabled, "enabled", false, "Enable runner immediately (default true)")
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.disabled, "disabled", false, "Disable runner on registration")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.projectName, "project-name", "", "Associate runner with specific project")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.webhook, "webhook", "", "Webhook URL for the runner")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.tags, "tags", "", "Comma-separated list of tags for the runner")
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

func registerRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	initRunnerRegistrationToken()

	// Apply CLI flags to override config values
	if runnerRegisterArgs.hostname != "" {
		util.Config.Runner.Name = runnerRegisterArgs.hostname
	}

	// Handle enabled/disabled flags with proper precedence
	if runnerRegisterArgs.disabled {
		enabled := false
		util.Config.Runner.Active = &enabled
	} else if runnerRegisterArgs.enabled {
		enabled := true
		util.Config.Runner.Active = &enabled
	} else if util.Config.Runner.Active == nil {
		// Default to enabled if not specified
		enabled := true
		util.Config.Runner.Active = &enabled
	}

	if runnerRegisterArgs.projectName != "" {
		util.Config.Runner.ProjectName = runnerRegisterArgs.projectName
	}

	if runnerRegisterArgs.webhook != "" {
		util.Config.Runner.Webhook = runnerRegisterArgs.webhook
	}

	if runnerRegisterArgs.tags != "" {
		util.Config.Runner.Tag = runnerRegisterArgs.tags
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
