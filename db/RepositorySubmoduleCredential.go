package db

import (
	"net/url"
	"reflect"
	"strings"

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
	if err := validateSubmoduleCredentialHost(c.Host); err != nil {
		return err
	}

	if c.AccessKeyID == 0 {
		return common_errors.NewValidationError("submodule credential access key is required")
	}

	return nil
}

// validateSubmoduleCredentialHost rejects anything that isn't a bare
// hostname, optionally with a ":port" suffix -- matching what
// resolveSubmoduleAccessKey compares against a submodule URL's host[:port].
// A scheme, path, query, fragment, userinfo, wildcard, or padding whitespace
// can never match a real submodule URL's host, so accepting it would let an
// admin save a mapping that silently never applies (the executor falls back
// to the main access key and the task fails).
func validateSubmoduleCredentialHost(host string) error {
	invalid := common_errors.NewValidationError("submodule credential host must be a hostname, optionally with a port")

	if host == "" || strings.TrimSpace(host) != host {
		return invalid
	}

	hostURL, err := url.Parse("//" + host)
	if err != nil ||
		hostURL.Hostname() == "" ||
		hostURL.Hostname() == "*" ||
		hostURL.Path != "" ||
		hostURL.RawQuery != "" ||
		hostURL.Fragment != "" ||
		hostURL.User != nil {
		return invalid
	}

	return nil
}
