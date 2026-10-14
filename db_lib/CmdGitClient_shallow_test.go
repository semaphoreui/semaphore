package db_lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitOut runs git in dir and returns its trimmed stdout.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return strings.TrimSpace(string(out))
}

// gitCommitFile writes content to name in dir and commits it, returning the hash.
func gitCommitFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
	gitOut(t, dir, "add", name)
	gitOut(t, dir, "commit", "-qm", "add "+name)
	return gitOut(t, dir, "rev-parse", "HEAD")
}

// With ShallowClone a clone holds one commit, can still be pulled when the
// upstream moves on, and can check out an older commit a task asks for by
// fetching it by hash.
func TestCmdGitClient_ShallowClone(t *testing.T) {
	setupGitClientTest(t)

	upstream := t.TempDir()
	gitInit(t, upstream)
	first := gitOut(t, upstream, "rev-parse", "HEAD")
	second := gitCommitFile(t, upstream, "second", "2")

	// git ignores --depth for a plain local path, so the upstream is reached
	// through file:// like any other remote.
	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, "file://"+upstream, "main")
	repo.ShallowClone = true
	repo.TmpDirName = "shallow"
	repo.Client = client

	require.NoError(t, client.Clone(repo))

	dir := repo.GetFullPath()
	assert.FileExists(t, filepath.Join(dir, ".git", "shallow"), "clone must be shallow")
	assert.Equal(t, "1", gitOut(t, dir, "rev-list", "--count", "HEAD"))
	assert.Equal(t, second, gitOut(t, dir, "rev-parse", "HEAD"))

	t.Run("pull after a new upstream commit", func(t *testing.T) {
		third := gitCommitFile(t, upstream, "third", "3")

		assert.True(t, client.CanBePulled(repo))
		require.NoError(t, client.Pull(repo))
		assert.Equal(t, third, gitOut(t, dir, "rev-parse", "HEAD"))
		assert.FileExists(t, filepath.Join(dir, ".git", "shallow"), "pull must keep the clone shallow")
	})

	t.Run("checkout of a commit outside the shallow history", func(t *testing.T) {
		require.NoError(t, client.Checkout(repo, first))
		assert.Equal(t, first, gitOut(t, dir, "rev-parse", "HEAD"))
		assert.NoFileExists(t, filepath.Join(dir, "second"))
	})
}

// Without ShallowClone the clone carries the whole history.
func TestCmdGitClient_FullCloneByDefault(t *testing.T) {
	setupGitClientTest(t)

	upstream := t.TempDir()
	gitInit(t, upstream)
	gitCommitFile(t, upstream, "second", "2")

	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, "file://"+upstream, "main")
	repo.TmpDirName = "full"

	require.NoError(t, client.Clone(repo))

	dir := repo.GetFullPath()
	assert.NoFileExists(t, filepath.Join(dir, ".git", "shallow"))
	assert.Equal(t, "2", gitOut(t, dir, "rev-list", "--count", "HEAD"))
}
