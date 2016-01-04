package models

import (
	"time"
)

type User struct {
	ID       string    `db:"id" json:"id"`
	Created  time.Time `db:"created" json:"created"`
	Username string    `db:"username" json:"username"`
	Name     string    `db:"name" json:"name"`
	Email    string    `db:"email" json:"email"`
	Password string    `db:"password" json:"password"`
}
