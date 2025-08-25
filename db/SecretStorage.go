package db

import "encoding/json"

type SecretStorageType string

const (
	SecretStorageTypeLocal SecretStorageType = "local"
	SecretStorageTypeVault SecretStorageType = "vault"
)

type SecretStorage struct {
	ID        int               `db:"id" json:"id" backup:"-"`
	ProjectID int               `db:"project_id" json:"project_id" backup:"-"`
	Name      string            `db:"name" json:"name"`
	Type      SecretStorageType `db:"type" json:"type"`
	Params    MapStringAnyField `db:"params" json:"params"`
	ReadOnly  bool              `db:"readonly" json:"readonly"`

	VaultToken string `db:"-" json:"vault_token,omitempty" backup:"-"`
}

type VaultSecretStorageParams struct {
	URL string `json:"url"`
}

func (s *SecretStorage) ExtractParams(target any) (err error) {
	content, err := json.Marshal(s.Params)
	if err != nil {
		return
	}

	switch target.(type) {
	case *VaultSecretStorageParams:
	default:
		err = &ValidationError{"invalid target type for extracting VaultSecretStorageParams"}
		return
	}

	err = json.Unmarshal(content, target)
	return
}
