package api

import (
	"database/sql"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/ansible-semaphore/mulekick"
	"github.com/gorilla/context"
	"golang.org/x/crypto/bcrypt"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []db.User
	if _, err := db.Mysql.Select(&users, "select * from user"); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	if err := mulekick.Bind(w, r, &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	editor := context.Get(r, "user").(*db.User)
	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user.Created = db.GetParsedTime(time.Now())

	if err := db.Mysql.Insert(&user); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, user)
}

func getUserMiddleware(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetIntParam("user_id", w, r)
	if err != nil {
		return
	}

	var user db.User
	if err := db.Mysql.SelectOne(&user, "select * from user where id=?", userID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	editor := context.Get(r, "user").(*db.User)
	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	context.Set(r, "_user", user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	oldUser := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	var user db.User
	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	if !editor.Admin && editor.ID != oldUser.ID {
		log.Warn(editor.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if editor.ID == oldUser.ID && oldUser.Admin != user.Admin {
		log.Warn("User can't edit his own role")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if oldUser.External && oldUser.Username != user.Username {
		log.Warn("Username is not editable for external LDAP users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := db.Mysql.Exec("update user set name=?, username=?, email=?, alert=?, admin=? where id=?", user.Name, user.Username, user.Email, user.Alert, user.Admin, oldUser.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	var pwd struct {
		Pwd string `json:"password"`
	}

	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.External {
		log.Warn("Password is not editable for external LDAP users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := mulekick.Bind(w, r, &pwd); err != nil {
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(pwd.Pwd), 11)
	util.LogWarning(err)
	if _, err := db.Mysql.Exec("update user set password=? where id=?", string(password), user.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to delete users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if _, err := db.Mysql.Exec("delete from project__user where user_id=?", user.ID); err != nil {
		panic(err)
	}
	if _, err := db.Mysql.Exec("delete from user where id=?", user.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
