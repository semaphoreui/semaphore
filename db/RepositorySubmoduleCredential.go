package db

import (
	"reflect"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// RepositorySubmoduleCredential maps a git host to an AccessKey that should be
// used to authenticate submodule clones targeting that host, for repositories
// whose submodules live on a different host/credentials than the main repo.
type RepositorySubmoduleCredential struct {
	ID           int `db:"id" json:"id" backup:"-"`
	ProjectID    int `db:"project_id" json:"project_id" backup:"-"`
	RepositoryID int `db:"repository_id" json:"repository_id" backup:"-"`
	// Host is matched exactly (case-insensitively) against the hostname of a
	// submodule's remote URL.
	Host        string `db:"host" json:"host" binding:"required"`
	AccessKeyID int    `db:"access_key_id" json:"access_key_id" binding:"required"`

	AccessKey AccessKey `db:"-" json:"-" backup:"-"`
}

var RepositorySubmoduleCredentialProps = ObjectProps{
	TableName:             "project__repository_submodule_credential",
	Type:                  reflect.TypeFor[RepositorySubmoduleCredential](),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "submodule_credential_id",
}

func (c RepositorySubmoduleCredential) Validate() error {
	if c.Host == "" || c.Host == "*" {
		return common_errors.NewValidationError("submodule credential host can't be empty")
	}

	if c.AccessKeyID == 0 {
		return common_errors.NewValidationError("submodule credential access key is required")
	}

	return nil
}
