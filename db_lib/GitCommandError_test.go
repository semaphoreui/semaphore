package db_lib

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitErrorDetail(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected string
	}{
		{"empty", "", ""},
		{"blank lines only", "\n  \n", ""},
		{"single line", "fatal: Authentication failed\n", "fatal: Authentication failed"},
		{
			"remote explanation and fatal line",
			"remote: HTTP Basic: Access denied\nfatal: Authentication failed for 'https://gitlab.com/g/r.git/'\n",
			"remote: HTTP Basic: Access denied\nfatal: Authentication failed for 'https://gitlab.com/g/r.git/'",
		},
		{"carriage returns split lines", "Receiving objects: 10%\rReceiving objects: 20%\r\nfatal: early EOF", "Receiving objects: 10%\nReceiving objects: 20%\nfatal: early EOF"},
		{"drops the clone progress line", "Cloning into 'repository_1_browse_b28b7af69320'...\nfatal: Authentication failed", "fatal: Authentication failed"},
		{"keeps the last lines", "1\n2\n3\n4\n5\n6\n7\n", "3\n4\n5\n6\n7"},
		{"invalid utf-8 is dropped", "fatal: bad \xff name", "fatal: bad  name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, gitErrorDetail(tt.stderr))
		})
	}
}

func TestGitCommandError(t *testing.T) {
	exitErr := &exec.ExitError{}

	t.Run("with stderr", func(t *testing.T) {
		repo := db.Repository{
			GitURL: "https://gitlab.com/group/repo.git",
			SSHKey: db.AccessKey{
				Type:          db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{Login: "oauth2", Password: "glpat-SECRET"},
			},
		}

		err := newGitCommandError(repo, []string{"ls-remote", "--heads"},
			"fatal: unable to access 'https://oauth2:glpat-SECRET@gitlab.com/group/repo.git/'\n", exitErr)

		assert.Equal(t, "ls-remote", err.Command)
		assert.Equal(t, "git ls-remote: fatal: unable to access 'https://oauth2:***@gitlab.com/group/repo.git/'", err.Error())
		assert.NotContains(t, err.Error(), "SECRET")

		var target *exec.ExitError
		assert.True(t, errors.As(err, &target), "the exit error must stay reachable")
	})

	t.Run("without stderr", func(t *testing.T) {
		err := newGitCommandError(db.Repository{}, []string{"clone"}, "", errors.New("exit status 128"))

		assert.Equal(t, "git clone: exit status 128", err.Error())
	})
}

func TestTailBuffer(t *testing.T) {
	t.Run("keeps everything under the limit", func(t *testing.T) {
		b := &tailBuffer{limit: 64}
		_, _ = b.Write([]byte("line 1\n"))
		_, _ = b.Write([]byte("line 2\n"))

		assert.Equal(t, "line 1\nline 2\n", b.String())
	})

	t.Run("keeps the tail and drops the partial first line", func(t *testing.T) {
		b := &tailBuffer{limit: 16}
		n, err := b.Write([]byte("https://u:SECRET@host\nfatal: failed\n"))

		require.NoError(t, err)
		assert.Equal(t, 36, n, "Write must report every byte as written")
		assert.Equal(t, "fatal: failed\n", b.String())
		assert.NotContains(t, b.String(), "CRET")
	})

	t.Run("drops a single partial line entirely", func(t *testing.T) {
		b := &tailBuffer{limit: 4}
		_, _ = b.Write([]byte("SECRET"))

		assert.Equal(t, "", b.String())
	})

	t.Run("nil buffer", func(t *testing.T) {
		var b *tailBuffer

		assert.Equal(t, "", b.String())
	})
}

// TestCmdGitClient_ErrorCarriesStderr runs real git commands that fail and
// checks that git's own explanation reaches the returned error, which is what
// the template form shows instead of a bare "exit status 128".
func TestCmdGitClient_ErrorCarriesStderr(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	setupGitClientTest(t)

	missing := filepath.Join(t.TempDir(), "missing.git")
	client := CreateCmdGitClient(nopKeyInstaller{})

	t.Run("output, as used by GetRemoteBranches", func(t *testing.T) {
		_, err := client.GetRemoteBranches(newTestGitRepo(t, missing, "main"))

		var gitErr *GitCommandError
		require.True(t, errors.As(err, &gitErr), "got %v", err)
		assert.Equal(t, "ls-remote", gitErr.Command)
		assert.Contains(t, gitErr.Detail, "fatal:")
		assert.Contains(t, gitErr.Error(), missing)

		var exitErr *exec.ExitError
		require.True(t, errors.As(err, &exitErr))
		assert.Empty(t, exitErr.Stderr, "unredacted stderr must not be kept on the exit error")
	})

	t.Run("run with NopLogger, as used by the playbooks endpoint", func(t *testing.T) {
		repo := newTestGitRepo(t, missing, "main")
		repo.TmpDirName = "clone_target"

		err := client.Clone(repo)

		var gitErr *GitCommandError
		require.True(t, errors.As(err, &gitErr), "got %v", err)
		assert.Equal(t, "clone", gitErr.Command)
		assert.Contains(t, gitErr.Detail, "fatal:")
	})

	t.Run("run with a task logger, which owns stderr", func(t *testing.T) {
		logger := &outputCaptureLogger{}
		repo := newTestGitRepo(t, missing, "main")
		repo.TmpDirName = "clone_target_logged"
		repo.Logger = logger

		err := client.Clone(repo)

		var gitErr *GitCommandError
		require.True(t, errors.As(err, &gitErr), "got %v", err)
		assert.Empty(t, gitErr.Detail, "stderr belongs to the task log")
		assert.True(t, strings.HasPrefix(gitErr.Error(), "git clone: exit status "), gitErr.Error())
		assert.Contains(t, logger.stderr.String(), "fatal:")
	})
}
