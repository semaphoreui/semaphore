package db

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type EnvironmentSecretOperation string

const (
	EnvironmentSecretCreate EnvironmentSecretOperation = "create"
	EnvironmentSecretUpdate EnvironmentSecretOperation = "update"
	EnvironmentSecretDelete EnvironmentSecretOperation = "delete"
)

type EnvironmentSecretType string

const (
	EnvironmentSecretVar EnvironmentSecretType = "var"
	EnvironmentSecretEnv EnvironmentSecretType = "env"
)

func (t EnvironmentSecretType) GetAccessKeyOwner() AccessKeyOwner {
	switch t {
	case EnvironmentSecretVar:
		return AccessKeyVariable
	case EnvironmentSecretEnv:
		return AccessKeyEnvironment
	default:
		panic("unknown secret type: " + t)
	}
}

// EnvironmentSecret is one secret of a Variable Group: which Access Key reaches the group under which
// name and in which role. The value always lives in a `string` Access Key. The key is owned by the group
// when it was created from the group form with a value (AccessKey.Owner = variable/environment,
// AccessKey.EnvironmentID = the group); otherwise it is linked — a shared Key Store key or the owned key
// of another group.
type EnvironmentSecret struct {
	ID            int                   `db:"id" json:"id"`
	ProjectID     int                   `db:"project_id" json:"-"`
	EnvironmentID int                   `db:"environment_id" json:"-"`
	Type          EnvironmentSecretType `db:"type" json:"type"`
	Name          string                `db:"name" json:"name"`
	KeyID         int                   `db:"key_id" json:"key_id,omitempty"`

	// Secret is the value of a new or changed owned key. Input only: it is filled from the API request and
	// at task start, never when the group is read back.
	Secret    string                     `db:"-" json:"secret,omitempty"`
	Operation EnvironmentSecretOperation `db:"-" json:"operation,omitempty"`

	// Output only, for the form: the name of the linked key and whether this group owns it.
	KeyName string `db:"-" json:"key_name,omitempty"`
	Owned   bool   `db:"-" json:"owned"`
}

// EnvironmentSecretWithKey is a secret row together with the Access Key it points to.
type EnvironmentSecretWithKey struct {
	EnvironmentSecret
	Key AccessKey
}

// Environment is used to pass additional arguments, in json form to ansible
type Environment struct {
	ID        int     `db:"id" json:"id" backup:"-"`
	Name      string  `db:"name" json:"name" binding:"required"`
	ProjectID int     `db:"project_id" json:"project_id" backup:"-"`
	Password  *string `db:"password" json:"password"`
	JSON      string  `db:"json" json:"json" binding:"required"`
	ENV       *string `db:"env" json:"env" binding:"required"`

	// Secrets is a field which used to update secrets associated with the environment.
	Secrets []EnvironmentSecret `db:"-" json:"secrets,omitempty" backup:"-"`

	SecretStorageID        *int    `db:"secret_storage_id" json:"secret_storage_id,omitempty" backup:"-"`
	SecretStorageKeyPrefix *string `db:"secret_storage_key_prefix" json:"secret_storage_key_prefix,omitempty"`

	// Sync fields are transfer-only; persisted in project__secret_sync.
	SyncEnabled      bool             `db:"-" json:"sync_enabled"`
	SyncInterval     int              `db:"-" json:"sync_interval"`
	LastSyncedAt     *time.Time       `db:"-" json:"last_synced_at,omitempty"`
	LastSyncFailedAt *time.Time       `db:"-" json:"last_sync_failed_at,omitempty"`
	SyncPaths        []SecretSyncPath `db:"-" json:"sync_paths"`
}

func (s *EnvironmentSecret) Validate() error {
	switch s.Type {
	case EnvironmentSecretVar, EnvironmentSecretEnv:
	default:
		return common_errors.NewValidationError("invalid environment secret type")
	}

	switch s.Operation {
	case EnvironmentSecretCreate:
		if s.Name == "" {
			return common_errors.NewValidationError("secret name can not be empty")
		}
		if s.Secret == "" && s.KeyID == 0 {
			return common_errors.NewValidationError("secret " + s.Name + ": either a value or a key is required")
		}
		if s.Secret != "" && s.KeyID != 0 {
			return common_errors.NewValidationError("secret " + s.Name + ": a value and a key can not be set together")
		}
	case EnvironmentSecretUpdate:
		if s.Name == "" {
			return common_errors.NewValidationError("secret name can not be empty")
		}
		if s.Secret != "" && s.KeyID != 0 {
			return common_errors.NewValidationError("secret " + s.Name + ": a value and a key can not be set together")
		}
	case EnvironmentSecretDelete:
	default:
		return common_errors.NewValidationError("invalid environment secret operation")
	}

	return nil
}

// ValidateEnvironmentSecretKey tells whether an Access Key may back a Variable Group secret: only a
// `string` key of the project that is shared or owned by a group. Task-bound survey secrets and keys of
// a secret storage are not meant to be read by a group. Checked when a link is written and again when the
// value is resolved — the key may have been overwritten with another type since.
func ValidateEnvironmentSecretKey(key AccessKey) error {
	if key.Type != AccessKeyString {
		return common_errors.NewValidationError("key " + key.Name + " is " + string(key.Type) + ", only string keys can be used as a secret")
	}

	switch key.Owner {
	case AccessKeyShared, AccessKeyVariable, AccessKeyEnvironment:
	default:
		return common_errors.NewValidationError("key " + key.Name + " can not be used as a variable group secret")
	}

	if key.TaskID != nil {
		return common_errors.NewValidationError("key " + key.Name + " belongs to a task")
	}

	return nil
}

func validateJSON(s string, mustValuesBeScalar bool) error {
	if s == "" {
		return nil
	}

	var data map[string]any
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return errors.New("must be valid JSON")
	}

	for k, v := range data {
		if k == "" {
			return errors.New("key can not be empty")
		}

		if mustValuesBeScalar {
			switch v.(type) {
			case []any, map[string]any:
				return errors.New("values must be scalar")
			}
		}
	}

	return nil
}

func (env *Environment) Validate() (err error) {
	if env.Name == "" {
		err = common_errors.NewValidationError("Environment name can not be empty")
		return
	}

	err = validateJSON(env.JSON, false)
	if err != nil {
		err = common_errors.NewValidationError("Extra variables " + err.Error())
		return
	}

	if env.ENV == nil {
		return
	}

	err = validateJSON(*env.ENV, true)
	if err != nil {
		err = common_errors.NewValidationError("Environment variables " + err.Error())
	}

	return
}
