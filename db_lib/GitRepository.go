package db_lib

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

const (
	// gitRetryDelay is the delay before the first retry. It doubles on each
	// following attempt, up to gitRetryMaxDelay.
	gitRetryDelay = time.Second

	// gitRetryMaxDelay caps the doubling. A git server which has been down for
	// a minute is not helped by waiting longer, and an uncapped delay grows past
	// what a Duration can hold.
	gitRetryMaxDelay = time.Minute
)

type GitRepositoryDirType int

const (
	GitRepositoryTmpPath GitRepositoryDirType = iota
	GitRepositoryFullPath
)

type GitClient interface {
	Clone(r GitRepository) error
	Pull(r GitRepository) error
	Checkout(r GitRepository, target string) error
	CanBePulled(r GitRepository) bool
	GetLastCommitMessage(r GitRepository) (msg string, err error)
	GetLastCommitHash(r GitRepository) (hash string, err error)
	GetLastRemoteCommitHash(r GitRepository) (hash string, err error)
	GetRemoteBranches(r GitRepository) ([]string, error)
}

type GitRepository struct {
	TmpDirName string
	TemplateID int
	Repository db.Repository
	Logger     task_logger.Logger
	Client     GitClient

	// HostConfigs are the credential mappings of the project, applied to every
	// host this repository reaches, including submodules.
	HostConfigs *ssh.HostConfigInstallation

	// retryDelay overrides gitRetryDelay so tests do not have to wait for it.
	retryDelay time.Duration
}

func (r GitRepository) GetFullPath() string {
	if r.TmpDirName != "" {
		return path.Join(util.Config.GetProjectTmpDir(r.Repository.ProjectID), r.TmpDirName)
	}
	return r.Repository.GetFullPath(r.TemplateID)
}

// ErrNotGitRepository is returned by ValidateRepo when the checkout directory
// exists but holds no git metadata, e.g. after a clone which was interrupted
// before git created it. Callers tell it apart from a missing directory
// (os.IsNotExist) because the fix is different: a missing directory is
// cloned, a broken one is removed first.
var ErrNotGitRepository = errors.New("directory is not a git repository")

// ValidateRepo reports whether the checkout directory holds a git repository.
// Checking the directory alone is not enough: Clone() creates it before git
// runs, so an interrupted clone leaves a directory which would otherwise be
// taken for a repository and then fail on every pull with
// "fatal: not a git repository".
func (r GitRepository) ValidateRepo() error {
	fullPath := r.GetFullPath()

	if _, err := os.Stat(fullPath); err != nil {
		return err
	}

	// .git is a directory for a clone and a file for a worktree or submodule;
	// either one marks a repository.
	if _, err := os.Stat(path.Join(fullPath, ".git")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrNotGitRepository, fullPath)
		}
		return err
	}

	return nil
}

// CloneOrPull brings the checkout directory up to date with the remote branch.
// A missing directory is cloned; a directory which is not a repository, or
// one whose pull fails, is removed and cloned from scratch.
func (r GitRepository) CloneOrPull() error {
	err := r.ValidateRepo()

	if err != nil {
		if !os.IsNotExist(err) {
			if err = os.RemoveAll(r.GetFullPath()); err != nil {
				return err
			}
		}
		return r.Clone()
	}

	if r.CanBePulled() {
		if err = r.Pull(); err == nil {
			return nil
		}
	}

	if err = os.RemoveAll(r.GetFullPath()); err != nil {
		return err
	}

	return r.Clone()
}

// gitRetryDelayFor returns the delay before the given attempt: base doubled once
// per attempt already made, up to gitRetryMaxDelay.
//
// The doubling is capped rather than shifted freely: a large attempt budget
// overflows the int64 of a Duration and a negative duration makes time.Sleep
// return at once, which would turn the backoff into a loop of immediate requests
// against the git server the retries are waiting for.
func gitRetryDelayFor(base time.Duration, attempt int) time.Duration {
	delay := base

	for i := 1; i < attempt && delay < gitRetryMaxDelay; i++ {
		delay *= 2
	}

	if delay > gitRetryMaxDelay {
		return gitRetryMaxDelay
	}

	return delay
}

// retry runs a git network operation again when it fails, for git servers which
// are intermittently unavailable. Retrying here rather than at the call sites
// keeps a transient outage from reaching updateRepository(), which reacts to a
// failed pull by deleting the local repository and cloning it from scratch.
func (r GitRepository) retry(name string, op func() error) (err error) {
	base := r.retryDelay
	if base == 0 {
		base = gitRetryDelay
	}

	attempts := max(util.Config.GitAttempts, 1)

	for attempt := 1; ; attempt++ {
		if err = op(); err == nil || attempt >= attempts {
			return
		}

		delay := gitRetryDelayFor(base, attempt)
		r.Logger.Log(fmt.Sprintf("Git %s failed (%s), retrying in %s", name, err, delay))
		time.Sleep(delay)
	}
}

func (r GitRepository) Clone() error {
	err := r.retry("clone", func() error {
		err := r.Client.Clone(r)
		if err != nil {
			// Remove any partial/corrupt clone so the next attempt starts fresh
			// instead of getting stuck on ValidateRepo() passing against a broken dir.
			_ = os.RemoveAll(r.GetFullPath())
		}
		return err
	})
	return err
}

func (r GitRepository) Pull() error {
	return r.retry("pull", func() error {
		return r.Client.Pull(r)
	})
}

func (r GitRepository) Checkout(target string) error {
	return r.Client.Checkout(r, target)
}

func (r GitRepository) CanBePulled() bool {
	return r.Client.CanBePulled(r)
}

func (r GitRepository) GetLastCommitMessage() (msg string, err error) {
	return r.Client.GetLastCommitMessage(r)
}

func (r GitRepository) GetLastCommitHash() (hash string, err error) {
	return r.Client.GetLastCommitHash(r)
}

func (r GitRepository) GetLastRemoteCommitHash() (hash string, err error) {
	return r.Client.GetLastRemoteCommitHash(r)
}

func (r GitRepository) GetRemoteBranches() ([]string, error) {
	return r.Client.GetRemoteBranches(r)
}
