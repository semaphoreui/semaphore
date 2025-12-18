package db

import (
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/util"
)

// TemplateAdditionalRepository represents an additional repository attached to a template
type TemplateAdditionalRepository struct {
	ID           int        `db:"id" json:"id" backup:"-"`
	TemplateID   int        `db:"template_id" json:"template_id" backup:"-"`
	RepositoryID int        `db:"repository_id" json:"repository_id" binding:"required"`
	Path         string     `db:"path" json:"path" binding:"required"`
	GitBranch    *string    `db:"git_branch" json:"git_branch,omitempty"`
	Position     int        `db:"position" json:"position"`
	Repository   *Repository `db:"-" json:"repository,omitempty" backup:"-"`
}

// Validate validates the TemplateAdditionalRepository
func (tar *TemplateAdditionalRepository) Validate() error {
	// Remove leading and trailing slashes
	tar.Path = strings.Trim(tar.Path, "/")

	if tar.Path == "" {
		return &ValidationError{"additional repository path cannot be empty"}
	}

	// Validate repository type - only git/ssh/https repos allowed
	if tar.Repository != nil {
		repoType := tar.Repository.GetType()
		if repoType != RepositoryGit && repoType != RepositorySSH && repoType != RepositoryHTTP {
			return &ValidationError{"additional repositories must be git, ssh, or https type"}
		}
	}

	// Path validation: alphanumeric, dash, underscore only
	// Prevents directory traversal attacks
	pathRegex := regexp.MustCompile(`^[a-zA-Z0-9_/-]+$`)
	if !pathRegex.MatchString(tar.Path) {
		return &ValidationError{"path must contain only alphanumeric characters, dashes, underscores, and forward slashes"}
	}

	if len(tar.Path) > 255 {
		return &ValidationError{"path cannot exceed 255 characters"}
	}

	// Check for directory traversal attempts
	if strings.Contains(tar.Path, "..") || strings.Contains(tar.Path, "\\") {
		return &ValidationError{"path cannot contain '..' or '\\'"}
	}

	// Reserved paths
	reservedPaths := []string{".", "..", "tmp", "cache", "logs", "log"}
	for _, reserved := range reservedPaths {
		if strings.EqualFold(tar.Path, reserved) {
			return &ValidationError{"path '" + reserved + "' is reserved"}
		}
	}

	return nil
}

// GetFullPath returns the full filesystem path for this additional repository
func (tar *TemplateAdditionalRepository) GetFullPath(projectID int, templateID int) string {
	if tar.Repository == nil {
		return ""
	}
	basePath := util.Config.GetProjectTmpDir(projectID)
	repoDir := "repository_" + strconv.Itoa(tar.RepositoryID) + "_template_" + strconv.Itoa(templateID)
	return path.Join(basePath, repoDir, "repos", tar.Path)
}
