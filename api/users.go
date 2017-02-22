package api

import (
	"database/sql"
	"net/http"
	"time"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"golang.org/x/crypto/bcrypt"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if _, err := database.Mysql.Select(&users, "select * from user"); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	user.Created = time.Now()

	if err := database.Mysql.Insert(&user); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, user)
}

func getUserMiddleware(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetIntParam("user_id", w, r)
	if err != nil {
		return
	}

	var user models.User
	if err := database.Mysql.SelectOne(&user, "select * from user where id=?", userID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "_user", user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	oldUser := context.Get(r, "_user").(models.User)

	var user models.User
	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update user set name=?, username=?, email=? where id=?", user.Name, user.Username, user.Email, oldUser.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(models.User)
	var pwd struct {
		Pwd string `json:"password"`
	}

	if err := mulekick.Bind(w, r, &pwd); err != nil {
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(pwd.Pwd), 11)
	if _, err := database.Mysql.Exec("update user set password=? where id=?", string(password), user.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(models.User)

	if _, err := database.Mysql.Exec("delete from project__user where user_id=?", user.ID); err != nil {
		panic(err)
	}
	if _, err := database.Mysql.Exec("delete from user where id=?", user.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
