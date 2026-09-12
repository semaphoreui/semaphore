package db

import "time"

type SecretStorageType string

const (
	SecretStorageTypeLocal   SecretStorageType = "local"
	SecretStorageTypeVault   SecretStorageType = "vault"
	SecretStorageTypeOpenBao SecretStorageType = "openbao"
	SecretStorageTypeDvls    SecretStorageType = "dvls"
	SecretStorageTypeAwsSm   SecretStorageType = "aws_sm"
	SecretStorageTypeAzureKv SecretStorageType = "azure_kv"
)

type SecretStorage struct {
	ID        int               `db:"id" json:"id" backup:"-"`
	ProjectID int               `db:"project_id" json:"project_id" backup:"-"`
	Name      string            `db:"name" json:"name"`
	Type      SecretStorageType `db:"type" json:"type"`
	Params    MapStringAnyField `db:"params" json:"params"`
	ReadOnly  bool              `db:"readonly" json:"readonly"`

	// Sync fields are transfer-only; persisted in project__secret_sync.
	SyncEnabled      bool             `db:"-" json:"sync_enabled" backup:"sync_enabled"`
	SyncInterval     int              `db:"-" json:"sync_interval" backup:"sync_interval"`
	LastSyncedAt     *time.Time       `db:"-" json:"last_synced_at,omitempty" backup:"-"`
	LastSyncFailedAt *time.Time       `db:"-" json:"last_sync_failed_at,omitempty" backup:"-"`
	SyncPaths        []SecretSyncPath `db:"-" json:"sync_paths" backup:"-"`

	SourceStorageType *AccessKeySourceStorageType `db:"-" json:"source_storage_type,omitempty" backup:"-"`
	// Secret is a source value: literal secret for local storage,
	// env var name for "env", or file path for "file".
	Secret string `db:"-" json:"secret,omitempty" backup:"-"`
}
