package db

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/git"
	"github.com/semaphoreui/semaphore/util"
)

type RepositoryType string

const (
	RepositoryGit   RepositoryType = "git"
	RepositorySSH   RepositoryType = "ssh"
	RepositoryHTTP  RepositoryType = "https"
	RepositoryFile  RepositoryType = "file"
	RepositoryLocal RepositoryType = "local"
)

// Repository is the model for code stored in a git repository
type Repository struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	GitURL    string `db:"git_url" json:"git_url" binding:"required"`
	GitBranch string `db:"git_branch" json:"git_branch" binding:"required"`
	SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required" backup:"-"`

	SSHKey AccessKey `db:"-" json:"-" backup:"-"`

	// WorkingCopyPath is the task copy to read repository files from. Empty
	// when the task reads the shared per-template checkout.
	WorkingCopyPath string `db:"-" json:"-" backup:"-"`
}

func (r Repository) ClearCache() error {
	return util.ClearDir(util.Config.GetProjectTmpDir(r.ProjectID), true, r.getDirNamePrefix())
}

func (r Repository) getDirNamePrefix() string {
	return "repository_" + strconv.Itoa(r.ID) + "_"
}

func (r Repository) GetDirName(templateID int) string {
	return r.getDirNamePrefix() + "template_" + strconv.Itoa(templateID)
}

// GetHomePath returns the per-template "home" directory with a "_home" suffix.
// Currently this path is used for home-like directories such as ANSIBLE_HOME so
// that parallel tasks from different templates get isolated home directories
// (preventing concurrent ansible-galaxy writes to the same collections path),
// while keeping these artifacts separate from the repository files.
func (r Repository) GetHomePath(templateID int) string {
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetDirName(templateID)+"_home")
}

// GetInternalPath returns a per-template directory under the project tmp dir for Semaphore-owned
// metadata (e.g. galaxy requirements hashes). It is not a copy of the repository.
func (r Repository) GetInternalPath(templateID int) string {
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetDirName(templateID)+"_internal")
}

// GetFullPath returns the shared per-template checkout, the mirror kept up to
// date by pull/clone. Code that reads repository files for a running task must
// use GetWorkingCopyPath instead: a parallel task reads its own copy.
// The repository is cloned directly into the template directory
// (e.g. repository_15_template_114) without any subdirectory.
func (r Repository) GetFullPath(templateID int) string {
	if r.GetType() == RepositoryLocal {
		return r.GetGitURL(true)
	}
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetDirName(templateID))
}

// GetWorkingCopyPath returns the path to read repository files from for
// templateID: the WorkingCopyPath override when set, otherwise GetFullPath.
func (r Repository) GetWorkingCopyPath(templateID int) string {
	if r.WorkingCopyPath != "" {
		return r.WorkingCopyPath
	}
	return r.GetFullPath(templateID)
}

func (r Repository) GetTaskCopyPath(templateID, taskID int) string {
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID),
		r.GetDirName(templateID)+"_task_"+strconv.Itoa(taskID))
}

func (r Repository) GetGitURL(secure bool) string {
	url := r.GitURL

	if r.GetType() == RepositoryLocal {
		return util.NormalizeLocalFilesystemPath(url)
	}

	if secure {
		return url
	}

	if r.GetType() == RepositoryHTTP {
		auth := ""
		switch r.SSHKey.Type {
		case AccessKeyLoginPassword:
			if r.SSHKey.LoginPassword.Login == "" {
				auth = r.SSHKey.LoginPassword.Password
			} else {
				auth = r.SSHKey.LoginPassword.Login + ":" + r.SSHKey.LoginPassword.Password
			}
		}
		if auth != "" {
			auth += "@"
		}

		re := regexp.MustCompile(`^(https?)://`)
		m := re.FindStringSubmatch(url)
		var protocol string

		if m == nil {
			panic(fmt.Errorf("invalid git url: %s", url))
		}

		protocol = m[1]

		url = protocol + "://" + auth + r.GitURL[len(protocol)+3:]
	}

	return url
}

func (r Repository) GetType() RepositoryType {
	if strings.HasPrefix(r.GitURL, "/") {
		return RepositoryLocal
	}

	if util.IsWindowsLocalRepositoryPath(r.GitURL) {
		return RepositoryLocal
	}

	re := regexp.MustCompile(`^(\w+)://`)
	m := re.FindStringSubmatch(r.GitURL)
	if m == nil {
		return RepositorySSH
	}

	protocol := m[1]

	switch protocol {
	case "http", "https":
		return RepositoryHTTP
	default:
		return RepositoryType(protocol)
	}
}

func (r Repository) Validate() error {
	if r.Name == "" {
		return common_errors.NewValidationError("repository name can't be empty")
	}

	if r.GitURL == "" {
		return common_errors.NewValidationError("repository url can't be empty")
	}

	if err := ValidateGitURL(r.GitURL, "repository"); err != nil {
		return err
	}

	if r.GetType() != RepositoryLocal && r.GitBranch == "" {
		return common_errors.NewValidationError("repository branch can't be empty")
	}

	if err := git.ValidateGitBranch(r.GitBranch, "repository"); err != nil {
		return err
	}

	return nil
}
