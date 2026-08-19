package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func setupRunnerConfig() {
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
	if util.Config.Runner == nil {
		util.Config.Runner = &util.RunnerConfig{}
	}
}

func TestInitRunnerRegistrationToken_FromFlagFile(t *testing.T) {
	setupRunnerConfig()
	tmp := filepath.Join(t.TempDir(), "token.txt")
	os.WriteFile(tmp, []byte("  my-token-123\n"), 0644)

	util.Config.Runner.RegistrationToken = ""
	util.Config.Runner.RegistrationTokenFile = ""
	runnerRegisterArgs.registrationTokenFilePath = tmp
	runnerRegisterArgs.stdinRegistrationToken = false

	initRunnerRegistrationToken()

	assert.Equal(t, "my-token-123", util.Config.Runner.RegistrationToken)
}

func TestInitRunnerRegistrationToken_FromConfigFile(t *testing.T) {
	setupRunnerConfig()
	tmp := filepath.Join(t.TempDir(), "token.txt")
	os.WriteFile(tmp, []byte("config-token\n"), 0644)

	util.Config.Runner.RegistrationToken = ""
	util.Config.Runner.RegistrationTokenFile = tmp
	runnerRegisterArgs.registrationTokenFilePath = ""
	runnerRegisterArgs.stdinRegistrationToken = false

	initRunnerRegistrationToken()

	assert.Equal(t, "config-token", util.Config.Runner.RegistrationToken)
}

func TestInitRunnerRegistrationToken_FromStdin(t *testing.T) {
	setupRunnerConfig()
	r, w, _ := os.Pipe()
	w.WriteString("stdin-token\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	util.Config.Runner.RegistrationToken = ""
	util.Config.Runner.RegistrationTokenFile = ""
	runnerRegisterArgs.registrationTokenFilePath = ""
	runnerRegisterArgs.stdinRegistrationToken = true

	initRunnerRegistrationToken()

	assert.Equal(t, "stdin-token", util.Config.Runner.RegistrationToken)
}

func TestInitRunnerRegistrationToken_NoSource(t *testing.T) {
	setupRunnerConfig()
	util.Config.Runner.RegistrationToken = ""
	util.Config.Runner.RegistrationTokenFile = ""
	runnerRegisterArgs.registrationTokenFilePath = ""
	runnerRegisterArgs.stdinRegistrationToken = false

	initRunnerRegistrationToken()

	assert.Empty(t, util.Config.Runner.RegistrationToken)
}

func TestInitRunnerRegistrationToken_EmptyFile_Panics(t *testing.T) {
	setupRunnerConfig()
	tmp := filepath.Join(t.TempDir(), "empty.txt")
	os.WriteFile(tmp, []byte(""), 0644)

	util.Config.Runner.RegistrationToken = ""
	runnerRegisterArgs.registrationTokenFilePath = tmp

	assert.Panics(t, func() {
		initRunnerRegistrationToken()
	})
}

func TestInitRunnerRegistrationToken_FlagFileTakesPriority(t *testing.T) {
	setupRunnerConfig()
	flagFile := filepath.Join(t.TempDir(), "flag-token.txt")
	configFile := filepath.Join(t.TempDir(), "config-token.txt")
	os.WriteFile(flagFile, []byte("flag-token"), 0644)
	os.WriteFile(configFile, []byte("config-token"), 0644)

	util.Config.Runner.RegistrationToken = ""
	util.Config.Runner.RegistrationTokenFile = configFile
	runnerRegisterArgs.registrationTokenFilePath = flagFile
	runnerRegisterArgs.stdinRegistrationToken = true

	initRunnerRegistrationToken()

	assert.Equal(t, "flag-token", util.Config.Runner.RegistrationToken)
}
