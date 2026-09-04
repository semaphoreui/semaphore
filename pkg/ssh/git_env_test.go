package ssh

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withSshConfig(t *testing.T, mode util.SshStrictHostKeyChecking, knownHosts, configPath string) {
	t.Helper()

	original := util.Config
	t.Cleanup(func() { util.Config = original })

	util.Config = &util.ConfigType{
		Ssh: &util.SshConfig{
			StrictHostKeyChecking: mode,
			KnownHostsFile:        knownHosts,
		},
		SshConfigPath: configPath,
	}
}

func gitSSHCommand(t *testing.T, env []string) string {
	t.Helper()

	for _, e := range env {
		if strings.HasPrefix(e, "GIT_SSH_COMMAND=") {
			return strings.TrimPrefix(e, "GIT_SSH_COMMAND=")
		}
	}
	return ""
}

// GIT_SSH_COMMAND must name the ssh executable exactly once. A second "ssh" is
// read by ssh as the host to connect to, which broke every SSH repository on
// the default configuration.
func TestGetGitEnv_SshCommandHasSingleExecutable(t *testing.T) {
	modes := []util.SshStrictHostKeyChecking{
		util.SshStrictHostKeyCheckingNo,
		util.SshStrictHostKeyCheckingYes,
		util.SshStrictHostKeyCheckingAcceptNew,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			withSshConfig(t, mode, "/tmp/known_hosts", "")

			installation := AccessKeyInstallation{SSHAgent: &Agent{SocketFile: "/tmp/agent.sock"}}
			cmd := gitSSHCommand(t, installation.GetGitEnv())

			require.NotEmpty(t, cmd, "GIT_SSH_COMMAND must be set when an agent is running")
			assert.Equal(t, "ssh", strings.Fields(cmd)[0], "must start with the ssh executable")
			assert.NotContains(t, cmd, "ssh ssh", "the executable must not be repeated")

			// Every remaining field is a flag or a flag's value, never a bare word
			// that ssh would treat as a hostname.
			assert.True(t, strings.HasPrefix(strings.Fields(cmd)[1], "-"),
				"expected an option after the executable, got %q", cmd)
		})
	}
}

func TestGetGitEnv_IncludesAgentSocketAndOptions(t *testing.T) {
	withSshConfig(t, util.SshStrictHostKeyCheckingNo, "", "")

	installation := AccessKeyInstallation{SSHAgent: &Agent{SocketFile: "/tmp/agent.sock"}}
	env := installation.GetGitEnv()

	assert.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
	assert.Contains(t, env, "SSH_AUTH_SOCK=/tmp/agent.sock")
	assert.Equal(t,
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
		gitSSHCommand(t, env))
}

func TestGetGitEnv_AppendsSshConfigPath(t *testing.T) {
	withSshConfig(t, util.SshStrictHostKeyCheckingNo, "", "/etc/semaphore/ssh_config")

	installation := AccessKeyInstallation{SSHAgent: &Agent{SocketFile: "/tmp/agent.sock"}}

	assert.Equal(t,
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -F /etc/semaphore/ssh_config",
		gitSSHCommand(t, installation.GetGitEnv()))
}

// Without an agent there is no key to offer, so no GIT_SSH_COMMAND is set.
func TestGetGitEnv_NoAgent(t *testing.T) {
	withSshConfig(t, util.SshStrictHostKeyCheckingNo, "", "")

	installation := AccessKeyInstallation{}
	env := installation.GetGitEnv()

	assert.Equal(t, []string{"GIT_TERMINAL_PROMPT=0"}, env)
}
