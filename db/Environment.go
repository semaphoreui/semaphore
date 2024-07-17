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
	ID        int                 `db:"id" json:"id"`
	Name      string              `db:"name" json:"name" binding:"required"`
	ProjectID int                 `db:"project_id" json:"project_id"`
	Password  *string             `db:"password" json:"password"`
	JSON      string              `db:"json" json:"json" binding:"required"`
	ENV       *string             `db:"env" json:"env" binding:"required"`
	Secrets   []EnvironmentSecret `db:"-" json:"secrets"`
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

func (env *Environment) Validate() error {
	if env.Name == "" {
		return &ValidationError{"Environment name can not be empty"}
	}

	if !json.Valid([]byte(env.JSON)) {
		return &ValidationError{"Extra variables must be valid JSON"}
	}

	if env.ENV != nil && !json.Valid([]byte(*env.ENV)) {
		return &ValidationError{"Environment variables must be valid JSON"}
	}

	return nil
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
