package db_lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// nopKeyInstaller is an AccessKeyInstaller that installs nothing (as for AccessKeyNone).
type nopKeyInstaller struct{}

func (nopKeyInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (ssh.AccessKeyInstallation, error) {
	return ssh.AccessKeyInstallation{}, nil
}

func gitInit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "config", "user.email", "t@t")
	gitRun(t, dir, "config", "user.name", "t")
	return dir
}

func gitAddFile(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0644))
	gitRun(t, dir, "add", name)
	gitRun(t, dir, "commit", "-qm", "add "+name)
}

func gitSubmoduleAdd(t *testing.T, dir, url, name string) {
	t.Helper()
	gitRun(t, dir, "-c", "protocol.file.allow=always", "submodule", "add", "--quiet", url, name)
	gitRun(t, dir, "commit", "-qm", "add submodule "+name)
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func gitRevParse(t *testing.T, dir, rev string) string {
	t.Helper()
	return strings.TrimSpace(gitRun(t, dir, "rev-parse", rev))
}

func newTestGitRepo(t *testing.T, gitURL, gitBranch, tmpDirName string) GitRepository {
	t.Helper()
	return GitRepository{
		TmpDirName: tmpDirName,
		Repository: db.Repository{
			ProjectID: 1,
			GitURL:    gitURL,
			GitBranch: gitBranch,
			SSHKey:    db.AccessKey{Type: db.AccessKeyNone},
		},
		Logger: task_logger.NopLogger{},
	}
}

// setupGitClientTest initializes util.Config with a temp dir and restores it afterwards.
func setupGitClientTest(t *testing.T) {
	t.Helper()
	original := util.Config
	t.Cleanup(func() { util.Config = original })
	util.Config = &util.ConfigType{TmpPath: t.TempDir(), Process: &util.ConfigProcess{}}
}

// TestCmdGitClient_OptionInjectionNeutralized proves that a GitURL crafted as a git
// option (e.g. "--upload-pack=/path/to/script") is NOT executed by the remote git
// operations, because "--end-of-options" is inserted before the URL positional.
func TestCmdGitClient_OptionInjectionNeutralized(t *testing.T) {
	setupGitClientTest(t)

	tmp := t.TempDir()
	marker := filepath.Join(tmp, "PWNED")
	evil := filepath.Join(tmp, "evil.sh")
	require.NoError(t, os.WriteFile(evil,
		[]byte("#!/bin/sh\necho pwned > "+marker+"\n"), 0755))

	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, "--upload-pack="+evil, "main", "")

	t.Run("GetLastRemoteCommitHash", func(t *testing.T) {
		_, err := client.GetLastRemoteCommitHash(repo)
		assert.Error(t, err, "injected upload-pack must not succeed")
		assert.NoFileExists(t, marker, "payload must not be executed")
	})

	require.NoError(t, os.RemoveAll(marker))

	t.Run("GetRemoteBranches", func(t *testing.T) {
		_, err := client.GetRemoteBranches(repo)
		assert.Error(t, err, "injected upload-pack must not succeed")
		assert.NoFileExists(t, marker, "payload must not be executed")
	})
}

// TestCmdGitClient_LegitRemoteOperations makes sure "--end-of-options" does not break
// normal remote operations against a valid repository.
func TestCmdGitClient_LegitRemoteOperations(t *testing.T) {
	setupGitClientTest(t)

	upstream := gitInit(t)
	gitAddFile(t, upstream, "initial.txt")

	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, upstream, "main", "")

	t.Run("GetLastRemoteCommitHash", func(t *testing.T) {
		hash, err := client.GetLastRemoteCommitHash(repo)
		require.NoError(t, err)
		assert.Len(t, hash, 40, "expected a full commit hash")
	})

	t.Run("GetRemoteBranches", func(t *testing.T) {
		branches, err := client.GetRemoteBranches(repo)
		require.NoError(t, err)
		assert.Equal(t, []string{"main"}, branches)
	})
}
