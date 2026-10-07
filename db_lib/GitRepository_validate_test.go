package db_lib

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingGitClient records which operations CloneOrPull chose. Clone lays
// down a .git entry on success, the way a real clone does, so a later
// ValidateRepo on the same directory passes.
type recordingGitClient struct {
	calls       []string
	cloneErr    error
	pullErr     error
	canBePulled bool
}

func (c *recordingGitClient) Clone(r GitRepository) error {
	c.calls = append(c.calls, "clone")
	if c.cloneErr != nil {
		return c.cloneErr
	}
	return os.MkdirAll(path.Join(r.GetFullPath(), ".git"), 0755)
}

func (c *recordingGitClient) Pull(GitRepository) error {
	c.calls = append(c.calls, "pull")
	return c.pullErr
}

func (c *recordingGitClient) CanBePulled(GitRepository) bool {
	c.calls = append(c.calls, "canBePulled")
	return c.canBePulled
}

func (c *recordingGitClient) Checkout(GitRepository, string) error                  { return nil }
func (c *recordingGitClient) GetLastCommitMessage(GitRepository) (string, error)    { return "", nil }
func (c *recordingGitClient) GetLastCommitHash(GitRepository) (string, error)       { return "", nil }
func (c *recordingGitClient) GetLastRemoteCommitHash(GitRepository) (string, error) { return "", nil }
func (c *recordingGitClient) GetRemoteBranches(GitRepository) ([]string, error)     { return nil, nil }

// newValidateTestRepo returns a repository whose checkout lives in a scratch
// dir under the temp dir set up by setupGitRetryTest.
func newValidateTestRepo(client GitClient) GitRepository {
	return GitRepository{
		Logger:     task_logger.NopLogger{},
		Client:     client,
		Repository: db.Repository{ProjectID: 1, GitBranch: "main"},
		TmpDirName: "repository_1_browse_test",
		retryDelay: time.Millisecond,
	}
}

// layoutDir prepares the checkout directory in one of the states ValidateRepo
// must tell apart.
func layoutDir(t *testing.T, fullPath, layout string) {
	t.Helper()
	switch layout {
	case "missing":
	case "empty":
		require.NoError(t, os.MkdirAll(fullPath, 0755))
	case "files-without-git":
		require.NoError(t, os.MkdirAll(fullPath, 0755))
		require.NoError(t, os.WriteFile(path.Join(fullPath, "playbook.yml"), []byte("- hosts: all\n"), 0644))
	case "git-dir":
		require.NoError(t, os.MkdirAll(path.Join(fullPath, ".git"), 0755))
	case "git-file":
		require.NoError(t, os.MkdirAll(fullPath, 0755))
		require.NoError(t, os.WriteFile(path.Join(fullPath, ".git"), []byte("gitdir: ../.git/worktrees/x\n"), 0644))
	default:
		t.Fatalf("unknown layout %q", layout)
	}
}

func TestGitRepositoryValidateRepo(t *testing.T) {
	tests := []struct {
		name       string
		layout     string
		isNotExist bool
		isNotGit   bool
	}{
		{"missing directory", "missing", true, false},
		// The states behind "fatal: not a git repository": Clone() creates the
		// directory before git runs, so an interrupted clone leaves one of these.
		{"empty directory", "empty", false, true},
		{"directory with files but no .git", "files-without-git", false, true},
		{"clone with .git directory", "git-dir", false, false},
		{"worktree with .git file", "git-file", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGitRetryTest(t, 1)
			repo := newValidateTestRepo(&recordingGitClient{})
			layoutDir(t, repo.GetFullPath(), tt.layout)

			err := repo.ValidateRepo()

			if !tt.isNotExist && !tt.isNotGit {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tt.isNotExist, os.IsNotExist(err), "os.IsNotExist")
			assert.Equal(t, tt.isNotGit, errors.Is(err, ErrNotGitRepository), "ErrNotGitRepository")
		})
	}
}

func TestGitRepositoryCloneOrPull(t *testing.T) {
	tests := []struct {
		name          string
		layout        string
		canBePulled   bool
		pullErr       error
		cloneErr      error
		expectedCalls []string
		expectedError bool
	}{
		{
			name:          "missing directory is cloned",
			layout:        "missing",
			expectedCalls: []string{"clone"},
		},
		{
			// The reported bug: a leftover directory without .git was pulled
			// and failed on every retry with "fatal: not a git repository".
			name:          "directory without .git is recloned, never pulled",
			layout:        "files-without-git",
			canBePulled:   true,
			expectedCalls: []string{"clone"},
		},
		{
			name:          "empty directory is recloned, never pulled",
			layout:        "empty",
			canBePulled:   true,
			expectedCalls: []string{"clone"},
		},
		{
			name:          "repository is pulled",
			layout:        "git-dir",
			canBePulled:   true,
			expectedCalls: []string{"canBePulled", "pull"},
		},
		{
			name:          "diverged repository is recloned",
			layout:        "git-dir",
			canBePulled:   false,
			expectedCalls: []string{"canBePulled", "clone"},
		},
		{
			name:          "failed pull falls back to a fresh clone",
			layout:        "git-dir",
			canBePulled:   true,
			pullErr:       errors.New("exit status 128"),
			expectedCalls: []string{"canBePulled", "pull", "clone"},
		},
		{
			name:          "clone failure is reported",
			layout:        "missing",
			cloneErr:      errors.New("exit status 128"),
			expectedCalls: []string{"clone"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGitRetryTest(t, 1)
			client := &recordingGitClient{
				canBePulled: tt.canBePulled,
				pullErr:     tt.pullErr,
				cloneErr:    tt.cloneErr,
			}
			repo := newValidateTestRepo(client)
			layoutDir(t, repo.GetFullPath(), tt.layout)

			err := repo.CloneOrPull()

			assert.Equal(t, tt.expectedCalls, client.calls)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			// Whatever the starting state, the directory must now be a
			// repository, so the next request pulls instead of recloning.
			assert.NoError(t, repo.ValidateRepo())
		})
	}
}

func TestGitRepositoryCloneOrPull_RecloneStartsFromEmptyDir(t *testing.T) {
	setupGitRetryTest(t, 1)
	client := &recordingGitClient{}
	repo := newValidateTestRepo(client)
	layoutDir(t, repo.GetFullPath(), "files-without-git")

	require.NoError(t, repo.CloneOrPull())

	// git refuses to clone into a non-empty directory, so the stale contents
	// must have been removed before Clone ran.
	_, err := os.Stat(path.Join(repo.GetFullPath(), "playbook.yml"))
	assert.True(t, os.IsNotExist(err))
}
