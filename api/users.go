package api

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	util2 "github.com/ansible-semaphore/semaphore/api/util"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := util2.GetStore(r).GetAllUsers()

	if err != nil {
		panic(err)
	}

	util2.WriteJSON(w, http.StatusOK, users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := util2.Bind(w, r, &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	editor := context.Get(r, "user").(*models.User)
	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	newUser, err := util2.GetStore(r).CreateUser(user)

	if err != nil {
		log.Warn(editor.Username + " is not created: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	util2.WriteJSON(w, http.StatusCreated, newUser)
}

func getUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := util2.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := util2.GetStore(r).GetUserById(userID)

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
	if err := util2.Bind(w, r, &user); err != nil {
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

	if err := util2.GetStore(r).UpdateUser(oldUser.ID, user); err != nil {
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

	if err := util2.Bind(w, r, &pwd); err != nil {
		return
	}

	if err := util2.GetStore(r).SetUserPassword(user.ID, pwd.Pwd); err != nil {
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

	if err := util2.GetStore(r).DeleteUser(user.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
