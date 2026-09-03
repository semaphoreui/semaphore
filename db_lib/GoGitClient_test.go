package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoGitClient_CloneLocal_ChecksOutRequestedCommit(t *testing.T) {
	setupGitClientTest(t)

	upstream := t.TempDir()
	gitInit(t, upstream)
	oldHash := gitRevParse(t, upstream, "HEAD")

	require.NoError(t, os.WriteFile(filepath.Join(upstream, "f"), []byte("bye"), 0644))
	gitRun(t, upstream, "add", "f")
	gitRun(t, upstream, "commit", "-qm", "second")
	require.NotEqual(t, oldHash, gitRevParse(t, upstream, "HEAD"))

	client := GoGitClient{}
	taskCopy := newTestGitRepo(t, upstream, "main")
	taskCopy.TmpDirName = "task-copy"
	dest := taskCopy.GetFullPath()
	require.NoError(t, client.CloneLocal(taskCopy, upstream, oldHash))

	content, err := os.ReadFile(filepath.Join(dest, "f"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(content), "checked-out content must match the pinned OLD commit, not HEAD")
}

func TestGoGitClient_CloneLocal_TakesSubmodulesFromSourceWithoutRemote(t *testing.T) {
	setupGitClientTest(t)

	subUpstream := t.TempDir()
	gitInit(t, subUpstream)
	upstream := t.TempDir()
	gitInit(t, upstream)
	gitRun(t, upstream, "-c", "protocol.file.allow=always", "submodule", "add", "--quiet", subUpstream, "sub")
	gitRun(t, upstream, "commit", "-qm", "add submodule")
	commitHash := gitRevParse(t, upstream, "HEAD")

	require.NoError(t, os.RemoveAll(subUpstream))

	client := GoGitClient{}
	taskCopy := newTestGitRepo(t, upstream, "main")
	taskCopy.TmpDirName = "task-copy"
	dest := taskCopy.GetFullPath()

	require.NoError(t, client.CloneLocal(taskCopy, upstream, commitHash))

	content, err := os.ReadFile(filepath.Join(dest, "sub", "f"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(content))
}
