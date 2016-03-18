package models

import (
	"time"

	"github.com/ansible-semaphore/semaphore/database"
)

type User struct {
	ID       int       `db:"id" json:"id"`
	Created  time.Time `db:"created" json:"created"`
	Username string    `db:"username" json:"username"`
	Name     string    `db:"name" json:"name"`
	Email    string    `db:"email" json:"email"`
	Password string    `db:"password" json:"password"`
}

func FetchUser(userID int) (*User, error) {
	var user User

	err := database.Mysql.SelectOne(&user, "select * from user where id=?", userID)
	return &user, err
}
