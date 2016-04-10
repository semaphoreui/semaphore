package models

import (
	"time"

	"github.com/ansible-semaphore/semaphore/database"
)

type User struct {
	ID       int       `db:"id" json:"id"`
	Created  time.Time `db:"created" json:"created"`
	Username string    `db:"username" json:"username" binding:"required"`
	Name     string    `db:"name" json:"name" binding:"required"`
	Email    string    `db:"email" json:"email" binding:"required"`
	Password string    `db:"password" json:"-"`
}

func FetchUser(userID int) (*User, error) {
	var user User

	err := database.Mysql.SelectOne(&user, "select * from user where id=?", userID)
	return &user, err
}

func init() {
	database.Mysql.AddTableWithName(User{}, "user").SetKeys(true, "id")
}
