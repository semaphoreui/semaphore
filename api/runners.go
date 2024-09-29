package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gorilla/context"
)

//type minimalGlobalRunner struct {
//	ID               int    `json:"id"`
//	Name             string `json:"name"`
//	Active           bool   `json:"active"`
//	Webhook          string `db:"webhook" json:"webhook"`
//	MaxParallelTasks int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
//}

func getGlobalRunners(w http.ResponseWriter, r *http.Request) {
	runners, err := helpers.Store(r).GetGlobalRunners(false)

	if err != nil {
		panic(err)
	}

	var result = make([]db.Runner, 0)

	for _, runner := range runners {
		result = append(result, runner)
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

type runnerWithToken struct {
	db.Runner
	Token string `json:"token"`
}

func addGlobalRunner(w http.ResponseWriter, r *http.Request) {
	var runner db.Runner
	if !helpers.Bind(w, r, &runner) {
		return
	}

	runner.ProjectID = nil
	newRunner, err := helpers.Store(r).CreateRunner(runner)

	if err != nil {
		log.Warn("Runner is not created: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, runnerWithToken{
		Runner: newRunner,
		Token:  newRunner.Token,
	})
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

		context.Set(r, "runner", &runner)
		next.ServeHTTP(w, r)
	})
}

func getGlobalRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(*db.Runner)

	helpers.WriteJSON(w, http.StatusOK, runner)
}

func updateGlobalRunner(w http.ResponseWriter, r *http.Request) {
	oldRunner := context.Get(r, "runner").(*db.Runner)

	var runner db.Runner
	if !helpers.Bind(w, r, &runner) {
		return
	}

	store := helpers.Store(r)

	runner.ID = oldRunner.ID
	runner.ProjectID = nil

	err := store.UpdateRunner(runner)

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

	err := store.UpdateRunner(*runner)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
