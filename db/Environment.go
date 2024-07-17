package db

import (
	"encoding/json"
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

func (secret *EnvironmentSecret) Validate() error {
	return nil
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
