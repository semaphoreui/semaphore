package db_lib

import (
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"path"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
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
		return path.Join(util.Config.TmpPath, r.TmpDirName)
	}
	return r.Repository.GetFullPath(r.TemplateID)
}

func (r GitRepository) ValidateRepo() error {
	_, err := os.Stat(r.GetFullPath())
	return err
}

func (r GitRepository) Clone() error {
	return r.Client.Clone(r)
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
