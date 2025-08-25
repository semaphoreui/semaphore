package db

import (
	"fmt"
	//"github.com/semaphoreui/semaphore/pkg/ssh"
	//"github.com/semaphoreui/semaphore/pkg/random"
	//"github.com/semaphoreui/semaphore/pkg/ssh"
	//"path"
)

type AccessKeyType string
type AccessKeyOwner string

const (
	AccessKeySSH           AccessKeyType = "ssh"
	AccessKeyNone          AccessKeyType = "none"
	AccessKeyLoginPassword AccessKeyType = "login_password"
	AccessKeyString        AccessKeyType = "string"
)
const (
	AccessKeyEnvironment AccessKeyOwner = "environment"
	AccessKeyVariable    AccessKeyOwner = "variable"
	AccessKeyVault       AccessKeyOwner = "vault"
	AccessKeyShared      AccessKeyOwner = ""
)

// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id" backup:"-"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'ssh/login_password/none'
	Type AccessKeyType `db:"type" json:"type" binding:"required"`

	ProjectID *int `db:"project_id" json:"project_id" backup:"-"`

	// Secret used internally, do not assign this field.
	// You should use methods SerializeSecret to fill this field.
	Secret *string `db:"secret" json:"-" backup:"-"`
	Plain  *string `db:"plain" json:"plain,omitempty"`

	String         string        `db:"-" json:"string"`
	LoginPassword  LoginPassword `db:"-" json:"login_password"`
	SshKey         SshKey        `db:"-" json:"ssh"`
	OverrideSecret bool          `db:"-" json:"override_secret,omitempty"`

	StorageID *int `db:"storage_id" json:"-" backup:"-"`

	// EnvironmentID is an ID of environment which owns the access key.
	EnvironmentID *int `db:"environment_id" json:"-" backup:"-"`

	// UserID is an ID of user which owns the access key.
	UserID *int `db:"user_id" json:"-" backup:"-"`

	Empty bool `db:"-" json:"empty,omitempty"`

	Owner AccessKeyOwner `db:"owner" json:"owner,omitempty"`

	SourceStorageID  *int    `db:"source_storage_id" json:"source_storage_id,omitempty" backup:"-"`
	SourceStorageKey *string `db:"source_storage_key" json:"source_storage_key,omitempty"`
}

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SshKey struct {
	Login      string `json:"login"`
	Passphrase string `json:"passphrase"`
	PrivateKey string `json:"private_key"`
}

type AccessKeyRole int

const (
	AccessKeyRoleAnsibleUser = iota
	AccessKeyRoleAnsibleBecomeUser
	AccessKeyRoleAnsiblePasswordVault
	AccessKeyRoleGit
)

func (key *AccessKey) Validate(validateSecretFields bool) error {
	if key.Name == "" {
		return fmt.Errorf("name can not be empty")
	}

	if !validateSecretFields {
		return nil
	}

	switch key.Type {
	case AccessKeySSH:
		if key.SshKey.PrivateKey == "" {
			return fmt.Errorf("private key can not be empty")
		}
	case AccessKeyLoginPassword:
		if key.LoginPassword.Password == "" {
			return fmt.Errorf("password can not be empty")
		}
	}

	return nil
}
