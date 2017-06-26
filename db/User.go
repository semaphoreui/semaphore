package db

import (
	"time"
)

type User struct {
	ID        int       `db:"id" json:"id"`
	Created   time.Time `db:"created" json:"created"`
	Username  string    `db:"username" json:"username" binding:"required"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Email     string    `db:"email" json:"email" binding:"required"`
	Password  string    `db:"password" json:"-"`
	External  bool      `db:"external" json:"external"`
	Alert     bool      `db:"alert" json:"alert"`
	ExtraVars string    `db:"extra_vars" json:"extra_vars"`
	Vault     string    `db:"vault" json:"vault"`
}

func FetchUser(userID int) (*User, error) {
	var user User

	err := Mysql.SelectOne(&user, "select * from user where id=?", userID)
	return &user, err
}
