package lib

import (
	"os"

	"github.com/ansible-semaphore/semaphore/db"
)

type GitRepositoryDirType int

const (
	GitRepositoryTmpDir GitRepositoryDirType = iota
	GitRepositoryRepoDir
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
	TemplateID int
	Repository db.Repository
	Logger     Logger
	Client     GitClient
}

func (r GitRepository) GetFullPath() (path string) {
	path = r.Repository.GetFullPath(r.TemplateID)
	return
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
