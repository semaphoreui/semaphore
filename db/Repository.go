package db

import (
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	RepositoryGit   = "git"
	RepositorySSH   = "ssh"
	RepositoryHTTPS = "https"
	RepositoryFile  = "file"
)

// Repository is the model for code stored in a git repository
type Repository struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	GitURL    string `db:"git_url" json:"git_url" binding:"required"`
	GitBranch string `db:"git_branch" json:"git_branch" binding:"required"`
	SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required"`
	Removed   bool   `db:"removed" json:"removed"`

	SSHKey AccessKey `db:"-" json:"-"`
}

func (r Repository) ClearCache() error {
	dir, err := os.Open(util.Config.TmpPath)
	if err != nil {
		return err
	}

	files, err := dir.ReadDir(0)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(), r.getDirNamePrefix()) {
			err = os.RemoveAll(path.Join(util.Config.TmpPath, f.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r Repository) getDirNamePrefix() string {
	return "repository_" + strconv.Itoa(r.ID) + "_"
}

func (r Repository) GetDirName(templateID int) string {
	return r.getDirNamePrefix() + strconv.Itoa(templateID)
}

func (r Repository) GetPath(templateID int) string {
	return path.Join(util.Config.TmpPath, r.GetDirName(templateID))
}

func (r Repository) GetGitURL() string {
	url := r.GitURL
	if r.getSchema() == RepositoryHTTPS && r.SSHKey.Type == AccessKeyLoginPassword {
		url = "https://" + r.SSHKey.LoginPassword.Login + ":" + r.SSHKey.LoginPassword.Password + "@" + r.GitURL[5:]
	}
	return url
}

func (r Repository) getSchema() string {
	re := regexp.MustCompile(`^(\w+)://`)
	m := re.FindStringSubmatch(r.GitURL)
	if m == nil {
		return RepositoryFile
	}
	return m[1]
}

func (r Repository) Validate() error {
	if r.Name == "" {
		return &ValidationError{"repository name can't be empty"}
	}

	if r.GitURL == "" {
		return &ValidationError{"repository url can't be empty"}
	}

	if r.GitBranch == "" {
		return &ValidationError{"repository branch can't be empty"}
	}

	return nil
}
