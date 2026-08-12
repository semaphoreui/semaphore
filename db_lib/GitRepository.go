package db_lib

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// gitRetryDelay is the delay before the first retry. It doubles on each
// following attempt.
const gitRetryDelay = time.Second

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

	// retryDelay overrides gitRetryDelay so tests do not have to wait for it.
	retryDelay time.Duration
}

func (r GitRepository) GetFullPath() string {
	if r.TmpDirName != "" {
		return path.Join(util.Config.GetProjectTmpDir(r.Repository.ProjectID), r.TmpDirName)
	}
	return r.Repository.GetFullPath(r.TemplateID)
}

func (r GitRepository) ValidateRepo() error {
	_, err := os.Stat(r.GetFullPath())
	return err
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

	attempts := util.Config.GitAttempts
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 1; ; attempt++ {
		if err = op(); err == nil || attempt >= attempts {
			return
		}

		delay := base << (attempt - 1)
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
			os.RemoveAll(r.GetFullPath())
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
