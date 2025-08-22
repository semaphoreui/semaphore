package db

import (
	"encoding/json"
	"errors"
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

type EnvironmentSecret struct {
	ID        int                        `json:"id"`
	Type      EnvironmentSecretType      `json:"type"`
	Name      string                     `json:"name"`
	Secret    string                     `json:"secret"`
	Operation EnvironmentSecretOperation `json:"operation"`
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
}

func (s *EnvironmentSecret) Validate() error {

	if s.Type == EnvironmentSecretVar || s.Type == EnvironmentSecretEnv {
		return nil
	}

	if s.Secret == "" {
		return errors.New("missing secret")
	}

	return errors.New("invalid environment secret type")
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
		err = &ValidationError{"Environment name can not be empty"}
		return
	}

	err = validateJSON(env.JSON, false)
	if err != nil {
		err = &ValidationError{"Extra variables " + err.Error()}
		return
	}

	if env.ENV == nil {
		return
	}

	err = validateJSON(*env.ENV, true)
	if err != nil {
		err = &ValidationError{"Environment variables " + err.Error()}
	}

	return
}
