package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoGitClient_CloneLocal covers the three things a task copy relies on: it
// lands on the requested commit, it takes its submodules from the shared checkout
// instead of the network, and it keeps working once the shared checkout is gone.
func TestGoGitClient_CloneLocal(t *testing.T) {
	client := GoGitClient{}

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
