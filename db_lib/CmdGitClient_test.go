package db_lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestCmdGitClient_AppliesHostConfigs proves the credential mappings of the
// project reach the git command. Without this the whole feature is inert: the
// config is generated, and git is never told to use it.
func TestCmdGitClient_AppliesHostConfigs(t *testing.T) {
	tmp, err := os.MkdirTemp("/tmp", "hcwire")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmp) }) //nolint:errcheck

	original := util.Config
	t.Cleanup(func() { util.Config = original })
	util.Config = &util.ConfigType{
		TmpPath: tmp,
		Process: &util.ConfigProcess{},
		Ssh:     &util.SshConfig{StrictHostKeyChecking: util.SshStrictHostKeyCheckingNo},
	}

	keyDir := t.TempDir()
	require.NoError(t, exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "",
		"-f", filepath.Join(keyDir, "k")).Run())
	private, err := os.ReadFile(filepath.Join(keyDir, "k"))
	require.NoError(t, err)

	key := db.AccessKey{ID: 1, Type: db.AccessKeySSH, SshKey: db.SshKey{PrivateKey: string(private)}}

	installation, err := ssh.InstallHostConfigs(1, []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "github.com", SSHKey: key},
		{ID: 2, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/", SSHKey: key},
	}, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	client := CmdGitClient{keyInstaller: nopKeyInstaller{}}
	repo := GitRepository{
		Repository:  db.Repository{ProjectID: 1, GitURL: "git@github.com:acme/x.git", GitBranch: "main"},
		Logger:      task_logger.NopLogger{},
		Client:      client,
		HostConfigs: installation,
	}

	cmd := client.makeCmd(repo, GitRepositoryTmpPath, ssh.AccessKeyInstallation{})

	var sshCommand, gitParams string
	for _, env := range cmd.Env {
		if v, ok := strings.CutPrefix(env, "GIT_SSH_COMMAND="); ok {
			sshCommand = v
		}
		if v, ok := strings.CutPrefix(env, "GIT_CONFIG_PARAMETERS="); ok {
			gitParams = v
		}
	}

	require.NotEmpty(t, sshCommand, "git must be told to use the generated config")
	assert.Contains(t, sshCommand, "-F "+installation.SSHConfigPath())
	assert.Contains(t, gitParams, "insteadOf=https://github.com/acme/")
}
