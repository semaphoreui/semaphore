package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

type minimalRunner struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func getRunners(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)
	runners, err := helpers.Store(r).GetGlobalRunners()

	if err != nil {
		panic(err)
	}

	if !currentUser.Admin {
		helpers.WriteErrorStatus(w, "You must be admin", http.StatusForbidden)
		return
	}

	var result = make([]minimalRunner, 0)

	for _, runner := range runners {
		result = append(result, minimalRunner{
			ID:     runner.ID,
			Name:   "",
			Active: false,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

//func addRunner(w http.ResponseWriter, r *http.Request) {
//	var user db.UserWithPwd
//	if !helpers.Bind(w, r, &user) {
//		return
//	}
//
//	editor := context.Get(r, "user").(*db.User)
//	if !editor.Admin {
//		log.Warn(editor.Username + " is not permitted to create users")
//		w.WriteHeader(http.StatusUnauthorized)
//		return
//	}
//
//	newUser, err := helpers.Store(r).CreateUser(user)
//
//	if err != nil {
//		log.Warn(editor.Username + " is not created: " + err.Error())
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	helpers.WriteJSON(w, http.StatusCreated, newUser)
//}

//func getRunnerMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		runnerID, err := helpers.GetIntParam("runner_id", w, r)
//
//		if err != nil {
//			return
//		}
//
//		runner, err := helpers.Store(r).GetGlobalRunner(runnerID)
//
//		if err != nil {
//			helpers.WriteError(w, err)
//			return
//		}
//
//		editor := context.Get(r, "runner").(*db.Runner)
//
//		if !editor.Admin && editor.ID != runner.ID {
//			log.Warn(editor.Username + " is not permitted to edit users")
//			w.WriteHeader(http.StatusUnauthorized)
//			return
//		}
//
//		context.Set(r, "_user", runner)
//		next.ServeHTTP(w, r)
//	})
//}

//func updateUser(w http.ResponseWriter, r *http.Request) {
//	targetUser := context.Get(r, "_user").(db.User)
//	editor := context.Get(r, "user").(*db.User)
//
//	var user db.UserWithPwd
//	if !helpers.Bind(w, r, &user) {
//		return
//	}
//
//	if !editor.Admin && editor.ID != targetUser.ID {
//		log.Warn(editor.Username + " is not permitted to edit users")
//		w.WriteHeader(http.StatusUnauthorized)
//		return
//	}
//
//	if editor.ID == targetUser.ID && targetUser.Admin != user.Admin {
//		log.Warn("User can't edit his own role")
//		w.WriteHeader(http.StatusUnauthorized)
//		return
//	}
//
//	if targetUser.External && targetUser.Username != user.Username {
//		log.Warn("Username is not editable for external users")
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	user.ID = targetUser.ID
//	if err := helpers.Store(r).UpdateUser(user); err != nil {
//		log.Error(err.Error())
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}
//
//func updateUserPassword(w http.ResponseWriter, r *http.Request) {
//	user := context.Get(r, "_user").(db.User)
//	editor := context.Get(r, "user").(*db.User)
//
//	var pwd struct {
//		Pwd string `json:"password"`
//	}
//
//	if !editor.Admin && editor.ID != user.ID {
//		log.Warn(editor.Username + " is not permitted to edit users")
//		w.WriteHeader(http.StatusUnauthorized)
//		return
//	}
//
//	if user.External {
//		log.Warn("Password is not editable for external users")
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	if !helpers.Bind(w, r, &pwd) {
//		return
//	}
//
//	if err := helpers.Store(r).SetUserPassword(user.ID, pwd.Pwd); err != nil {
//		util.LogWarning(err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}
//
//func deleteUser(w http.ResponseWriter, r *http.Request) {
//	user := context.Get(r, "_user").(db.User)
//	editor := context.Get(r, "user").(*db.User)
//
//	if !editor.Admin && editor.ID != user.ID {
//		log.Warn(editor.Username + " is not permitted to delete users")
//		w.WriteHeader(http.StatusUnauthorized)
//		return
//	}
//
//	if err := helpers.Store(r).DeleteUser(user.ID); err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}
