package db_lib

import (
	"os"
	"path"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
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

func (r GitRepository) Clone() error {
	err := r.Client.Clone(r)
	if err != nil {
		// Remove any partial/corrupt clone so the next attempt starts fresh
		// instead of getting stuck on ValidateRepo() passing against a broken dir.
		os.RemoveAll(r.GetFullPath())
	}
	return err
}

func (r GitRepository) Pull() error {
	return r.Client.Pull(r)
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
