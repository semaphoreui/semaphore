package db

type SecretStorageType string

const (
	SecretStorageTypeLocal SecretStorageType = "local"
	SecretStorageTypeVault SecretStorageType = "vault"
)

type SecretStorage struct {
	ID   int               `db:"id" json:"id"`
	Name string            `db:"name" json:"name"`
	Type SecretStorageType `db:"type" json:"type"`
}
