package models

import (
	"time"
)

// Project is the top level structure in Semaphore
type Project struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Created   time.Time `db:"created" json:"created"`
	Alert     bool      `db:"alert" json:"alert"`
	AlertChat string    `db:"alert_chat" json:"alert_chat"`
}
