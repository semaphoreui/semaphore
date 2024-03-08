package db

import (
	"time"
)

// Project is the top level structure in Semaphore
type Project struct {
	ID               int       `db:"id" json:"id"`
	Name             string    `db:"name" json:"name" binding:"required"`
	Created          time.Time `db:"created" json:"created"`
	Alert            bool      `db:"alert" json:"alert"`
	AlertChat        *string   `db:"alert_chat" json:"alert_chat"`
	MaxParallelTasks int       `db:"max_parallel_tasks" json:"max_parallel_tasks"`
	Type             string    `db:"type" json:"type"`
}
