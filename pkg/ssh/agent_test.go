package ssh

import (
	"testing"
)

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
