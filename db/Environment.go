package db

import (
	"encoding/json"
	"errors"
	"strings"
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
	Secrets []EnvironmentSecret `db:"-" json:"secrets" backup:"-"`
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

func FillEnvironmentSecrets(store Store, env *Environment, deserializeSecret bool) error {
	keys, err := store.GetEnvironmentSecrets(env.ProjectID, env.ID)

	if err != nil {
		return err
	}

	for _, k := range keys {
		var secretName string
		var secretType EnvironmentSecretType

		if strings.HasPrefix(k.Name, string(EnvironmentSecretVar)+".") {
			secretType = EnvironmentSecretVar
			secretName = strings.TrimPrefix(k.Name, string(EnvironmentSecretVar)+".")
		} else if strings.HasPrefix(k.Name, string(EnvironmentSecretEnv)+".") {
			secretType = EnvironmentSecretEnv
			secretName = strings.TrimPrefix(k.Name, string(EnvironmentSecretEnv)+".")
		} else {
			secretType = EnvironmentSecretVar
			secretName = k.Name
		}

		if deserializeSecret {
			err = k.DeserializeSecret()
			if err != nil {
				return err
			}
		}

		env.Secrets = append(env.Secrets, EnvironmentSecret{
			ID:     k.ID,
			Name:   secretName,
			Type:   secretType,
			Secret: k.String,
		})
	}

	return nil
}
