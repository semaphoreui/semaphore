package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/stretchr/testify/assert"
)

func TestGetRepositoryBranchNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "simple branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"def456\trefs/heads/develop",
			},
			expected: []string{"main", "develop"},
		},
		{
			name: "branch names with slashes",
			input: []string{
				"abc123\trefs/heads/env/test",
				"def456\trefs/heads/feature/my-feature",
			},
			expected: []string{"env/test", "feature/my-feature"},
		},
		{
			name: "mixed branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"def456\trefs/heads/env/test",
				"ghi789\trefs/heads/release/v1.0",
			},
			expected: []string{"main", "env/test", "release/v1.0"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "skip lines without tab",
			input: []string{
				"invalid line",
				"abc123\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "skip non-heads refs",
			input: []string{
				"abc123\trefs/tags/v1.0",
				"def456\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "trailing whitespace in ref path",
			input: []string{
				"abc123\trefs/heads/env/test\n",
			},
			expected: []string{"env/test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRepositoryBranchNames(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d branches, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, branch := range result {
				if branch != tt.expected[i] {
					t.Errorf("branch[%d]: expected %q, got %q", i, tt.expected[i], branch)
				}
			}
		})
	}
}

// TestGitProxyOpts covers the ssh options used to reach the git server of a
// repository through its proxy.
func TestGitProxyOpts(t *testing.T) {
	user := "ansible-proxy"
	port := 2222

	newRepo := func(keyID *int) db.Repository {
		return db.Repository{Proxy: &db.Proxy{
			Type:     db.ProxySSH,
			Host:     "bastion.example.org",
			User:     &user,
			Port:     &port,
			SSHKeyID: keyID,
		}}
	}

	t.Run("no proxy means no options", func(t *testing.T) {
		assert.Empty(t, gitProxyOpts(db.Repository{}, ssh.AccessKeyInstallation{}))
	})

	t.Run("proxy adds a ProxyCommand jump", func(t *testing.T) {
		assert.Equal(t,
			[]string{"-o", `"ProxyCommand=ssh -o StrictHostKeyChecking=no -W %h:%p -p 2222 ansible-proxy@bastion.example.org"`},
			gitProxyOpts(newRepo(nil), ssh.AccessKeyInstallation{}))
	})

	// The proxy key must reach the jump host only. IdentityAgent on the outer
	// ssh would apply to the git server too, which uses a different key.
	t.Run("the proxy key agent is scoped to the jump", func(t *testing.T) {
		keyID := 7
		installation := ssh.AccessKeyInstallation{SSHAgent: &ssh.Agent{SocketFile: "/tmp/proxy.sock"}}

		opts := gitProxyOpts(newRepo(&keyID), installation)

		assert.Equal(t,
			[]string{"-o", `"ProxyCommand=ssh -o IdentityAgent=/tmp/proxy.sock -o StrictHostKeyChecking=no -W %h:%p -p 2222 ansible-proxy@bastion.example.org"`},
			opts)
		assert.NotContains(t, opts[1][:len(opts[1])-1], "ProxyJump",
			"ProxyJump would authenticate the jump with the git key agent")
	})

	t.Run("a proxy without a port omits -p", func(t *testing.T) {
		repo := db.Repository{Proxy: &db.Proxy{Type: db.ProxySSH, Host: "bastion.example.org"}}

		assert.Equal(t,
			[]string{"-o", `"ProxyCommand=ssh -o StrictHostKeyChecking=no -W %h:%p bastion.example.org"`},
			gitProxyOpts(repo, ssh.AccessKeyInstallation{}))
	})
}
