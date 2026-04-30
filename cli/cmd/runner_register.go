package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var runnerRegisterArgs struct {
	stdinRegistrationToken    bool
	registrationTokenFilePath string
	projectID                 int
}

func init() {
	runnerRegisterCmd.PersistentFlags().BoolVar(&runnerRegisterArgs.stdinRegistrationToken, "stdin-registration-token", false, "Read registration token from stdin")
	runnerRegisterCmd.PersistentFlags().StringVar(&runnerRegisterArgs.registrationTokenFilePath, "registration-token-file", "", "Read registration token from a file")
	runnerRegisterCmd.PersistentFlags().IntVar(&runnerRegisterArgs.projectID, "project-id", 0, "Project ID for project-level runner (global runner if not provided)")
	runnerCmd.AddCommand(runnerRegisterCmd)
}

func readRegistrationTokenFromFile(path string) {
	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if len(tokenBytes) == 0 {
		panic("Empty token")
	}

	util.Config.Runner.RegistrationToken = strings.TrimSpace(string(tokenBytes))
}

func initRunnerRegistrationToken() {
	if runnerRegisterArgs.registrationTokenFilePath != "" {
		readRegistrationTokenFromFile(runnerRegisterArgs.registrationTokenFilePath)
		return
	}

	if util.Config.Runner.RegistrationTokenFile != "" {
		readRegistrationTokenFromFile(util.Config.Runner.RegistrationTokenFile)
		return
	}

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

	if runnerRegisterArgs.projectID > 0 {
		util.Config.Runner.ProjectID = &runnerRegisterArgs.projectID
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
