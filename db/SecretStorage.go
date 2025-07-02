package db

type SecretStorageType string

const (
	SecretStorageVault SecretStorageType = "vault"
)

type SecretStorage struct {
	ID        int               `db:"id" json:"id" backup:"-"`
	ProjectID int               `db:"project_id" json:"project_id" backup:"-"`
	Name      string            `db:"name" json:"name"`
	Type      SecretStorageType `db:"type" json:"type"`
}
