package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCmdGitClient_CloneLocal covers the three things a task copy relies on: it
// lands on the requested commit, it takes its submodules from the shared checkout
// instead of the network, and it keeps working once the shared checkout is gone.
func TestCmdGitClient_CloneLocal(t *testing.T) {
	client := CreateCmdGitClient(nopKeyInstaller{})

	t.Run("checks out the requested commit", func(t *testing.T) {
		setupGitClientTest(t)

		sharedCheckout := gitInit(t)
		gitAddFile(t, sharedCheckout, "requested.txt")
		requestedCommit := gitRevParse(t, sharedCheckout, "HEAD")
		gitAddFile(t, sharedCheckout, "tip.txt")

		taskCopy := newTestGitRepo(t, sharedCheckout, "main", "task-copy")

		require.NoError(t, client.CloneLocal(taskCopy, sharedCheckout, requestedCommit))

		assert.FileExists(t, filepath.Join(taskCopy.GetFullPath(), "requested.txt"))
		assert.NoFileExists(t, filepath.Join(taskCopy.GetFullPath(), "tip.txt"),
			"the copy must stop at the requested commit, not follow the branch tip")
		assert.Equal(t, requestedCommit, gitRevParse(t, taskCopy.GetFullPath(), "HEAD"))
		assert.FileExists(t, filepath.Join(sharedCheckout, "tip.txt"),
			"the shared checkout must not be moved by a task copy")
	})

	t.Run("takes submodules from the shared checkout", func(t *testing.T) {
		setupGitClientTest(t)

		submoduleRemote := gitInit(t)
		gitAddFile(t, submoduleRemote, "submodule.txt")

		sharedCheckout := gitInit(t)
		gitSubmoduleAdd(t, sharedCheckout, submoduleRemote, "sub")

		// Nothing can fetch the submodule any more, it has to come
		// from the shared checkout.
		require.NoError(t, os.RemoveAll(submoduleRemote))

		taskCopy := newTestGitRepo(t, sharedCheckout, "main", "task-copy")

		require.NoError(t, client.CloneLocal(taskCopy, sharedCheckout,
			gitRevParse(t, sharedCheckout, "HEAD")))

		assert.FileExists(t, filepath.Join(taskCopy.GetFullPath(), "sub", "submodule.txt"))
	})

	t.Run("survives removal of the shared checkout", func(t *testing.T) {
		setupGitClientTest(t)

		sharedCheckout := gitInit(t)
		gitAddFile(t, sharedCheckout, "initial.txt")
		requestedCommit := gitRevParse(t, sharedCheckout, "HEAD")

		taskCopy := newTestGitRepo(t, sharedCheckout, "main", "task-copy")

		require.NoError(t, client.CloneLocal(taskCopy, sharedCheckout, requestedCommit))

		// updateRepository throws the shared checkout away and clones again whenever
		// it cannot be pulled, so the task copy must not depend on it. Peeling HEAD to
		// a commit makes git read the object store, not just the ref file.
		require.NoError(t, os.RemoveAll(sharedCheckout))

		assert.Equal(t, requestedCommit, gitRevParse(t, taskCopy.GetFullPath(), "HEAD^{commit}"))
	})
}

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
