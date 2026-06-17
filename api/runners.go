package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	log "github.com/sirupsen/logrus"
)

type runnerWithToken struct {
	db.Runner
	Token string `json:"token"`
}

// GlobalRunnerController handles CRUD for global (non-project) runners.
type GlobalRunnerController struct {
	runnerService server.RunnerService
}

func NewGlobalRunnerController(runnerService server.RunnerService) *GlobalRunnerController {
	return &GlobalRunnerController{
		runnerService: runnerService,
	}
}

func (c *GlobalRunnerController) GetRunners(w http.ResponseWriter, r *http.Request) {
	runners, err := helpers.Store(r).GetAllRunners(false, false, db.RunnerFilterIgnoreTags, nil)

	if err != nil {
		panic(err)
	}

	var result = make([]db.Runner, 0)

	result = append(result, runners...)

	for i := range result {
		result[i].Registered = result[i].IsRegistered()
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

func (c *GlobalRunnerController) AddRunner(w http.ResponseWriter, r *http.Request) {
	var runner db.Runner
	if !helpers.Bind(w, r, &runner) {
		return
	}

	runner.ProjectID = nil

	newRunner, err := c.runnerService.CreateRunner(runner)

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

func (c *GlobalRunnerController) RunnerMiddleware(next http.Handler) http.Handler {
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

		r = helpers.SetContextValue(r, "runner", &runner)
		next.ServeHTTP(w, r)
	})
}

func (c *GlobalRunnerController) GetRunner(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(*db.Runner)

	runner.Registered = runner.IsRegistered()

	helpers.WriteJSON(w, http.StatusOK, runner)
}

func (c *GlobalRunnerController) UpdateRunner(w http.ResponseWriter, r *http.Request) {
	oldRunner := helpers.GetFromContext(r, "runner").(*db.Runner)

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

func (c *GlobalRunnerController) ClearRunnerCache(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(*db.Runner)

	store := helpers.Store(r)

	err := store.ClearRunnerCache(*runner)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *GlobalRunnerController) DeleteRunner(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(*db.Runner)

	store := helpers.Store(r)

	err := store.DeleteGlobalRunner(runner.ID)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *GlobalRunnerController) RegenerateRegistrationToken(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(*db.Runner)

	token, err := c.runnerService.RegenerateRegistrationToken(*runner)

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"registration_token": token,
		"runner_id":          runner.ID,
	})
}

func (c *GlobalRunnerController) GetRunnerTags(w http.ResponseWriter, r *http.Request) {
	tags, err := helpers.Store(r).GetGlobalRunnerTags()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tags)
}

func (c *GlobalRunnerController) SetRunnerActive(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(*db.Runner)

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
