package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// TestAgent_Listen_CreatesSocketDir tests that Listen() creates the socket's
// parent directory if it doesn't exist (e.g. project tmp dir not yet created).
func TestAgent_Listen_CreatesSocketDir(t *testing.T) {
	// Not t.TempDir(): its path exceeds the ~104-byte unix socket limit on macOS.
	tmp, err := os.MkdirTemp("/tmp", "ssh")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmp) }) //nolint:errcheck

	agent := Agent{
		SocketFile: filepath.Join(tmp, "project_2", "agent.sock"),
	}

	err = agent.Listen()
	require.NoError(t, err)
	require.NoError(t, agent.Close())
}

// TestAgent_Close_WithNilListener tests that Close() doesn't panic when listener is nil
func TestAgent_Close_WithNilListener(t *testing.T) {
	// Create agent with nil listener (simulates failed initialization)
	agent := Agent{}

	// This should not panic
	err := agent.Close()
	if err != nil {
		t.Errorf("Expected no error when closing agent with nil listener, got: %v", err)
	}
}

// TestAgent_Close_WithNilDone tests that Close() doesn't panic when done channel is nil
func TestAgent_Close_WithNilDone(t *testing.T) {
	// Create agent with nil done channel
	agent := Agent{
		done: nil,
	}

	// This should not panic
	err := agent.Close()
	if err != nil {
		t.Errorf("Expected no error when closing agent with nil done channel, got: %v", err)
	}
}

// TestAgent_Close_WithAllNil tests that Close() doesn't panic when both fields are nil
func TestAgent_Close_WithAllNil(t *testing.T) {
	// Create completely empty agent (simulates NewAgent() result)
	agent := NewAgent()

	// This should not panic
	err := agent.Close()
	if err != nil {
		t.Errorf("Expected no error when closing empty agent, got: %v", err)
	}
}

// TestAgent_Close_FailedInitialization simulates the exact scenario from issue #3232
// where agent initialization fails but the agent is still assigned to installation
func TestAgent_Close_FailedInitialization(t *testing.T) {
	// Simulate the scenario described in the issue:
	// 1. StartSSHAgent() fails during Listen() but returns incomplete agent
	// 2. Install() method assigns the incomplete agent to installation.SSHAgent
	// 3. Later, destroyKeys() calls Destroy() which calls Close() on incomplete agent

	// Create an agent that would be returned by StartSSHAgent() if Listen() failed
	incompleteAgent := Agent{
		Keys: []AgentKey{
			{
				Key:        []byte("test-private-key"),
				Passphrase: []byte(""),
			},
		},
		SocketFile: "/tmp/test-socket.sock",
		// listener and done are nil because Listen() failed
	}

	// This simulates the destroyKeys() -> Destroy() -> Close() call chain
	// that was causing the panic
	err := incompleteAgent.Close()
	if err != nil {
		t.Errorf("Expected no error when closing incomplete agent, got: %v", err)
	}
}

// TestGetGitEnv_SshCommandHasOneSshPrefix covers every host-key checking mode:
// gitHostKeyCheckingOpts must return options only, because GetGitEnv prepends
// "ssh". A second one is read by ssh as the host to connect to and every SSH
// repository fails with "Could not resolve hostname ssh".
func TestGetGitEnv_SshCommandHasOneSshPrefix(t *testing.T) {
	original := util.Config
	t.Cleanup(func() { util.Config = original })

	modes := []util.SshStrictHostKeyChecking{
		util.SshStrictHostKeyCheckingYes,
		util.SshStrictHostKeyCheckingNo,
		util.SshStrictHostKeyCheckingAcceptNew,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			util.Config = &util.ConfigType{Ssh: &util.SshConfig{
				StrictHostKeyChecking: mode,
				KnownHostsFile:        filepath.Join(t.TempDir(), "known_hosts"),
			}}

			key := AccessKeyInstallation{SSHAgent: &Agent{SocketFile: "/tmp/agent.sock"}}

			var cmd string
			for _, env := range key.GetGitEnv() {
				if strings.HasPrefix(env, "GIT_SSH_COMMAND=") {
					cmd = strings.TrimPrefix(env, "GIT_SSH_COMMAND=")
				}
			}

			require.NotEmpty(t, cmd)
			assert.NotContains(t, cmd, "ssh ssh")

			fields := strings.Fields(cmd)
			require.Greater(t, len(fields), 1)
			assert.Equal(t, "ssh", fields[0])
			assert.True(t, strings.HasPrefix(fields[1], "-"),
				"the argument after ssh must be an option, got %q", fields[1])
		})
	}
}

// A string key holds a bare value with no login and no private key, so no
// role that authenticates a user can use it. The installer must say so instead
// of handing an empty installation to the executor.
func TestKeyInstaller_Install_StringKeyIsRejectedForUserRoles(t *testing.T) {
	key := db.AccessKey{Name: "token", Type: db.AccessKeyString, String: "s3cr3t"}

	tests := []struct {
		name string
		role db.AccessKeyRole
	}{
		{"ansible user", db.AccessKeyRoleAnsibleUser},
		{"ansible become user", db.AccessKeyRoleAnsibleBecomeUser},
		{"ansible password vault", db.AccessKeyRoleAnsiblePasswordVault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installation, err := KeyInstaller{}.Install(key, tt.role, task_logger.NopLogger{})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "not supported")
			assert.Nil(t, installation.SSHAgent)
			assert.Empty(t, installation.Password, "the value must not leak into the installation")
		})
	}
}

// For the git role the installer only starts an agent for an ssh key; any
// other type is left to the git client, which rejects what it cannot use.
func TestKeyInstaller_Install_StringKeyForGitStartsNoAgent(t *testing.T) {
	key := db.AccessKey{Name: "token", Type: db.AccessKeyString, String: "s3cr3t"}

	installation, err := KeyInstaller{}.Install(key, db.AccessKeyRoleGit, task_logger.NopLogger{})

	require.NoError(t, err)
	assert.Nil(t, installation.SSHAgent)
	assert.Empty(t, installation.Login)
	assert.Empty(t, installation.Password)
}

// TestStartSSHAgent_UsesConfiguredSocketsDir checks that dirs.ssh_agent_sockets
// moves the agent socket out of tmp_path, and that the project tmp folder is
// still created 0755.
func TestStartSSHAgent_UsesConfiguredSocketsDir(t *testing.T) {
	original := util.Config
	t.Cleanup(func() { util.Config = original })

	// Not t.TempDir(): its path exceeds the ~104-byte unix socket limit on macOS.
	socketsDir, err := os.MkdirTemp("/tmp", "sock")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(socketsDir) }) //nolint:errcheck

	tmpPath, err := os.MkdirTemp("/tmp", "tmp")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpPath) }) //nolint:errcheck

	util.Config = &util.ConfigType{
		TmpPath: tmpPath,
		Dirs:    &util.ConfigDirs{SSHAgentSockets: socketsDir},
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	block, err := gossh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)

	projectID := 3
	key := db.AccessKey{ID: 5, ProjectID: &projectID}
	key.SshKey.PrivateKey = string(pem.EncodeToMemory(block))

	agent, err := StartSSHAgent(key, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = agent.Close() })

	assert.Equal(t, filepath.Join(socketsDir, "project_3"), filepath.Dir(agent.SocketFile))

	// The project tmp folder is still created 0755, as when the socket lived
	// there, so InstallHostConfigs does not make it 0700 first. The umask
	// trims the mode, so compare with a directory made the same way (0755).
	refDir := filepath.Join(t.TempDir(), "ref")
	require.NoError(t, os.MkdirAll(refDir, 0o755))
	refInfo, err := os.Stat(refDir)
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(tmpPath, "project_3"))
	require.NoError(t, err)
	assert.Equal(t, refInfo.Mode().Perm(), info.Mode().Perm())
}
