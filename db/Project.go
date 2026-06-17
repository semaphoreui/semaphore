package db

import (
	"time"
)

// Project is the top level structure in Semaphore
type Project struct {
	ID                     int       `db:"id" json:"id" backup:"-"`
	Name                   string    `db:"name" json:"name" binding:"required"`
	Created                time.Time `db:"created" json:"created" backup:"-"`
	Alert                  bool      `db:"alert" json:"alert,omitempty"`
	AlertChat              *string   `db:"alert_chat" json:"alert_chat,omitempty"`
	AlertPushoverToken     *string   `db:"alert_pushover_token" json:"alert_pushover_token,omitempty"`
	MaxParallelTasks       int       `db:"max_parallel_tasks" json:"max_parallel_tasks,omitempty"`
	Type                   string    `db:"type" json:"type"`
	DefaultSecretStorageID *int      `db:"default_secret_storage_id" json:"default_secret_storage_id,omitempty" backup:"-"`
}
