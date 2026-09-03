package db_lib

import (
	"errors"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flakyGitClient fails the first failures calls to Clone and Pull, then succeeds.
// Everything else is unused by the retry tests.
type flakyGitClient struct {
	failures int
	calls    int
}

func (c *flakyGitClient) attempt() error {
	c.calls++
	if c.calls <= c.failures {
		return errors.New("the remote end hung up unexpectedly")
	}
	return nil
}

func (c *flakyGitClient) Clone(GitRepository) error            { return c.attempt() }
func (c *flakyGitClient) Pull(GitRepository) error             { return c.attempt() }
func (c *flakyGitClient) Checkout(GitRepository, string) error { return nil }
func (c *flakyGitClient) CanBePulled(GitRepository) bool       { return true }

func (c *flakyGitClient) GetLastCommitMessage(GitRepository) (string, error) { return "", nil }
func (c *flakyGitClient) GetLastCommitHash(GitRepository) (string, error)    { return "", nil }
func (c *flakyGitClient) GetLastRemoteCommitHash(GitRepository) (string, error) {
	return "", nil
}
func (c *flakyGitClient) GetRemoteBranches(GitRepository) ([]string, error) { return nil, nil }
func (c *flakyGitClient) CloneLocal(GitRepository, string, string) error    { return nil }

// setupGitRetryTest points util.Config at a temp dir with the given attempt budget.
func setupGitRetryTest(t *testing.T, attempts int) {
	t.Helper()
	original := util.Config
	t.Cleanup(func() { util.Config = original })
	util.Config = &util.ConfigType{
		TmpPath:     t.TempDir(),
		Process:     &util.ConfigProcess{},
		GitAttempts: attempts,
	}
}

func newRetryTestRepo(client GitClient) GitRepository {
	return GitRepository{
		Logger: task_logger.NopLogger{},
		Client: client,
		// Keep the test fast; the production delay starts at a second.
		retryDelay: time.Millisecond,
	}
}

func TestGitRepositoryRetry(t *testing.T) {
	tests := []struct {
		name          string
		attempts      int
		failures      int
		expectedCalls int
		expectedError bool
	}{
		{"succeeds without retrying", 4, 0, 1, false},
		{"recovers from a transient failure", 4, 1, 2, false},
		{"uses the whole attempt budget", 4, 3, 4, false},
		{"gives up once the budget is spent", 3, 99, 3, true},
		{"one attempt does not retry", 1, 99, 1, true},
		// A config value below one is meaningless; one attempt is still made
		// rather than none.
		{"a budget below one still tries once", 0, 99, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGitRetryTest(t, tt.attempts)

			t.Run("pull", func(t *testing.T) {
				client := &flakyGitClient{failures: tt.failures}

				err := newRetryTestRepo(client).Pull()

				assert.Equal(t, tt.expectedCalls, client.calls)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})

			t.Run("clone", func(t *testing.T) {
				client := &flakyGitClient{failures: tt.failures}

				err := newRetryTestRepo(client).Clone()

				assert.Equal(t, tt.expectedCalls, client.calls)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}

// TestGitRepositoryRetryBacksOff checks the delay grows between attempts, so an
// unavailable git server is not hammered at a fixed rate.
func TestGitRepositoryRetryBacksOff(t *testing.T) {
	setupGitRetryTest(t, 4)

	client := &flakyGitClient{failures: 3}
	repo := newRetryTestRepo(client)
	repo.retryDelay = 20 * time.Millisecond

	start := time.Now()
	require.NoError(t, repo.Pull())
	elapsed := time.Since(start)

	// 20ms + 40ms + 80ms of waiting between the four attempts.
	assert.GreaterOrEqual(t, elapsed, 140*time.Millisecond)
}

// TestGitRetryDelayFor covers the growth of the backoff and its ceiling. Without
// the ceiling the doubling overflows the int64 of a Duration and goes negative,
// which time.Sleep returns from immediately: the backoff would become a loop of
// immediate requests against an already struggling git server.
func TestGitRetryDelayFor(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"first retry waits the base delay", 1, time.Second},
		{"second doubles", 2, 2 * time.Second},
		{"third doubles again", 3, 4 * time.Second},
		{"growth stops at the ceiling", 7, time.Minute},
		{"a huge attempt count stays at the ceiling", 40, time.Minute},
		// 1s << 34 is negative, and 1s << 64 is zero.
		{"an attempt count which would overflow stays at the ceiling", 35, time.Minute},
		{"an attempt count which would shift to zero stays at the ceiling", 65, time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := gitRetryDelayFor(gitRetryDelay, tt.attempt)

			assert.Equal(t, tt.expected, delay)
			assert.Greater(t, delay, time.Duration(0), "a non-positive delay does not wait at all")
		})
	}
}
