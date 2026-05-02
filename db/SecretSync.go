package db

import "time"

// SecretSync describes a unit of remote-storage synchronization.
// A row with EnvironmentID == nil represents storage-level sync (imports
// access keys). A row with EnvironmentID set represents env-scoped sync
// (imports environment variables for that variable group).
type SecretSync struct {
	ID            int  `db:"id" json:"id" backup:"-"`
	ProjectID     int  `db:"project_id" json:"project_id" backup:"-"`
	StorageID     int  `db:"storage_id" json:"storage_id" backup:"-"`
	EnvironmentID *int `db:"environment_id" json:"environment_id,omitempty" backup:"-"`

	SyncEnabled bool `db:"sync_enabled" json:"sync_enabled"`
	// SyncInterval is the auto-sync period in minutes. Zero disables auto-sync.
	SyncInterval     int        `db:"sync_interval" json:"sync_interval"`
	LastSyncedAt     *time.Time `db:"last_synced_at" json:"last_synced_at,omitempty"`
	LastSyncFailedAt *time.Time `db:"last_sync_failed_at" json:"last_sync_failed_at,omitempty"`

	Paths []SecretSyncPath `db:"-" json:"paths"`
}

type SecretSyncPath struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	SyncID    int    `db:"sync_id" json:"sync_id" backup:"-"`
	Path      string `db:"path" json:"path"`
	Prefix    string `db:"prefix" json:"prefix"`
	Separator string `db:"separator" json:"separator"`
}

// SecretStorageSyncPath is retained as an alias for backward compatibility
// with callers that predate the SecretSync refactor.
type SecretStorageSyncPath = SecretSyncPath
