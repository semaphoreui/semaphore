package db

import (
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type RepositoryType string

const (
	RepositoryGit   RepositoryType = "git"
	RepositorySSH   RepositoryType = "ssh"
	RepositoryHTTPS RepositoryType = "https"
	RepositoryFile  RepositoryType = "file"
	RepositoryLocal RepositoryType = "local"
)

// Repository is the model for code stored in a git repository
type Repository struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	GitURL    string `db:"git_url" json:"git_url" binding:"required"`
	GitBranch string `db:"git_branch" json:"git_branch" binding:"required"`
	SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required"`

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

func (r Repository) GetFullPath(templateID int) string {
	if r.GetType() == RepositoryLocal {
		return r.GetGitURL()
	}
	return path.Join(util.Config.TmpPath, r.GetDirName(templateID))
}

func (r Repository) GetGitURL() string {
	url := r.GitURL

	if r.GetType() == RepositoryHTTPS {
		auth := ""
		switch r.SSHKey.Type {
		case AccessKeyLoginPassword:
			auth = r.SSHKey.LoginPassword.Login + ":" + r.SSHKey.LoginPassword.Password
		case AccessKeyPAT:
			auth = r.SSHKey.PAT
		}
		if auth != "" {
			auth += "@"
		}
		url = "https://" + auth + r.GitURL[8:]
	}

	return url
}

func (r Repository) GetType() RepositoryType {
	if strings.HasPrefix(r.GitURL, "/") {
		return RepositoryLocal
	}

	re := regexp.MustCompile(`^(\w+)://`)
	m := re.FindStringSubmatch(r.GitURL)
	if m == nil {
		return RepositorySSH
	}

	return RepositoryType(m[1])
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
