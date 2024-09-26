package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gorilla/context"
)

type minimalGlobalRunner struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Webhook          string `db:"webhook" json:"webhook"`
	MaxParallelTasks int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
}

func getGlobalRunners(w http.ResponseWriter, r *http.Request) {
	runners, err := helpers.Store(r).GetGlobalRunners()

	if err != nil {
		panic(err)
	}

	var result = make([]minimalGlobalRunner, 0)

	for _, runner := range runners {
		result = append(result, minimalGlobalRunner{
			ID:     runner.ID,
			Name:   "",
			Active: false,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

func addGlobalRunner(w http.ResponseWriter, r *http.Request) {
	var runner minimalGlobalRunner
	if !helpers.Bind(w, r, &runner) {
		return
	}

	editor := context.Get(r, "user").(*db.User)
	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	newRunner, err := helpers.Store(r).CreateRunner(db.Runner{
		Webhook:          runner.Webhook,
		MaxParallelTasks: runner.MaxParallelTasks,
	})

	if err != nil {
		log.Warn("Runner is not created: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newRunner)
}

func globalRunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runnerID, err := helpers.GetIntParam("runner_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "runner_id required",
			})
			return
		}

		store := helpers.Store(r)

		runner, err := store.GetGlobalRunner(runnerID)

		if err != nil {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "Runner not found",
			})
			return
		}

		context.Set(r, "runner", runner)
		next.ServeHTTP(w, r)
	})
}

func getGlobalRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(*db.Runner)

	helpers.WriteJSON(w, http.StatusOK, minimalGlobalRunner{
		Name:             "",
		Active:           true,
		Webhook:          runner.Webhook,
		MaxParallelTasks: runner.MaxParallelTasks,
	})
}

func updateGlobalRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(*db.Runner)

	store := helpers.Store(r)

	runner.ProjectID = nil

	err := store.UpdateRunner(*runner)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteGlobalRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(*db.Runner)

	store := helpers.Store(r)

	err := store.DeleteGlobalRunner(runner.ID)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setGlobalRunnerActive(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(*db.Runner)

	store := helpers.Store(r)

	var body struct {
		Active bool `json:"active"`
	}

	if !helpers.Bind(w, r, &body) {
		helpers.WriteErrorStatus(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	runner.Active = body.Active

	err := store.UpdateRunner(runner)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
