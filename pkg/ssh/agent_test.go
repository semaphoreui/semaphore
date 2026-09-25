package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessKeyInstallation_GetGitEnv_HostKeyChecking(t *testing.T) {
	originalConfig := util.Config
	t.Cleanup(func() {
		util.Config = originalConfig
	})

	tests := []struct {
		name     string
		checking util.SshStrictHostKeyChecking
		expected string
	}{
		{
			name:     "disabled",
			checking: util.SshStrictHostKeyCheckingNo,
			expected: "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
		},
		{
			name:     "strict",
			checking: util.SshStrictHostKeyCheckingYes,
			expected: "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=yes -o UserKnownHostsFile=/tmp/known_hosts",
		},
		{
			name:     "accept new",
			checking: util.SshStrictHostKeyCheckingAcceptNew,
			expected: "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=/tmp/known_hosts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.Config = &util.ConfigType{
				Ssh: &util.SshConfig{
					KnownHostsFile:        "/tmp/known_hosts",
					StrictHostKeyChecking: tt.checking,
				},
			}
			installation := AccessKeyInstallation{
				SSHAgent: &Agent{SocketFile: "/tmp/agent.sock"},
			}

			env := installation.GetGitEnv()

			assert.Equal(t, []string{
				"GIT_TERMINAL_PROMPT=0",
				"SSH_AUTH_SOCK=/tmp/agent.sock",
				tt.expected,
			}, env)
		})
	}
}

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
