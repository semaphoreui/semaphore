package api

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := context.Get(r, "store").(db.Store).GetAllUsers()

	if err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := util.Bind(w, r, &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	editor := context.Get(r, "user").(*models.User)
	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	newUser, err := context.Get(r, "store").(db.Store).CreateUser(user)

	if err != nil {
		log.Warn(editor.Username + " is not created: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	util.WriteJSON(w, http.StatusCreated, newUser)
}

func getUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := util.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := context.Get(r, "store").(db.Store).GetUserById(userID)

		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		editor := context.Get(r, "user").(*models.User)

		if !editor.Admin && editor.ID != user.ID {
			log.Warn(editor.Username + " is not permitted to edit users")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		context.Set(r, "_user", user)
		next.ServeHTTP(w, r)
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	oldUser := context.Get(r, "_user").(models.User)
	editor := context.Get(r, "user").(*models.User)

	var user models.User
	if err := util.Bind(w, r, &user); err != nil {
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

	if err := context.Get(r, "store").(db.Store).UpdateUser(oldUser.ID, user); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(models.User)
	editor := context.Get(r, "user").(*models.User)

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

	if err := util.Bind(w, r, &pwd); err != nil {
		return
	}

	if err := context.Get(r, "store").(db.Store).SetUserPassword(user.ID, pwd.Pwd); err != nil {
		util.LogWarning(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(models.User)
	editor := context.Get(r, "user").(*models.User)

	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to delete users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := context.Get(r, "store").(db.Store).DeleteUser(user.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
