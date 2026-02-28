package db

import (
	"fmt"
)

type AccessKeyType string
type AccessKeyOwner string

type AccessKeySourceStorageType string

const (
	AccessKeySSH           AccessKeyType = "ssh"
	AccessKeyNone          AccessKeyType = "none"
	AccessKeyLoginPassword AccessKeyType = "login_password"
	AccessKeyString        AccessKeyType = "string"
)
const (
	AccessKeyEnvironment   AccessKeyOwner = "environment"
	AccessKeyVariable      AccessKeyOwner = "variable"
	AccessKeySecretStorage AccessKeyOwner = "vault"
	AccessKeyShared        AccessKeyOwner = ""
)
const (
	AccessKeySourceStorageVault AccessKeySourceStorageType = "vault"
	AccessKeySourceStorageEnv   AccessKeySourceStorageType = "env"
	AccessKeySourceStorageFile  AccessKeySourceStorageType = "file"
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

	// UserID is an ID of a user which owns the access key.
	UserID *int `db:"user_id" json:"-" backup:"-"`

	Empty bool `db:"-" json:"empty,omitempty"`

	Owner AccessKeyOwner `db:"owner" json:"owner,omitempty"`

	// SourceStorageID represents the ID of the source storage associated with the access key, used for reference purposes.
	SourceStorageID *int `db:"source_storage_id" json:"source_storage_id,omitempty" backup:"-"`

	// SourceStorageKey is an optional reference to a specific storage key associated with the source storage.
	// For example, for HashiCorp Vault, this is the path to the secret.
	// If SourceStorageID is nil, this field is references to an environment variable.
	SourceStorageKey  *string                     `db:"source_storage_key" json:"source_storage_key,omitempty"`
	SourceStorageType *AccessKeySourceStorageType `db:"source_storage_type" json:"source_storage_type,omitempty"`
}

func (key *AccessKey) IsNativelyReadOnly() bool {
	if key.SourceStorageType == nil {
		return false
	}

	return *key.SourceStorageType == AccessKeySourceStorageFile || *key.SourceStorageType == AccessKeySourceStorageEnv
}

func (key *AccessKey) IsEmpty() bool {
	if key == nil {
		return true
	}

	if key.Type == AccessKeyNone {
		return false
	}

	if key.SourceStorageType != nil {
		switch *key.SourceStorageType {
		case AccessKeySourceStorageEnv, AccessKeySourceStorageFile:
			return key.SourceStorageKey == nil || *key.SourceStorageKey == ""
		case AccessKeySourceStorageVault:
			return key.SourceStorageID == nil
		default:
			return true
		}
	}

	if key.Secret != nil && *key.Secret != "" {
		return false
	}

	switch key.Type {
	case AccessKeyString:
		return key.String == ""
	case AccessKeySSH:
		return key.SshKey.PrivateKey == ""
	case AccessKeyLoginPassword:
		return key.LoginPassword.Password == ""
	default:
		return true
	}
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

	//if !validateSecretFields {
	//	return nil
	//}

	//switch key.Type {
	//case AccessKeySSH:
	//	if key.SshKey.PrivateKey == "" {
	//		return fmt.Errorf("private key can not be empty")
	//	}
	//case AccessKeyLoginPassword:
	//	if key.LoginPassword.Password == "" {
	//		return fmt.Errorf("password can not be empty")
	//	}
	//}

	return nil
}

func (key *AccessKey) IsEnvironmentVariable() bool {
	return key.SourceStorageID == nil && key.SourceStorageKey != nil
}
