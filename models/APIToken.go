package models

import (
	"github.com/ansible-semaphore/semaphore/database"

	"time"
)

type APIToken struct {
	ID      string    `db:"id" json:"id"`
	Created time.Time `db:"created" json:"created"`
	Expired bool      `db:"expired" json:"expired"`
	UserID  int       `db:"user_id" json:"user_id"`
}

func init() {
	database.Mysql.AddTableWithName(APIToken{}, "user__token").SetKeys(false, "id")
}
