package runners

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
)

func RunnerMiddleware(next http.Handler) http.Handler {
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

func GetRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(db.Runner)

	data := runners.RunnerState{}

	tasks := helpers.TaskPool(r).GetRunningTasks()

	for _, tsk := range tasks {
		if tsk.RunnerID != runner.ID {
			continue
		}

		if tsk.Task.Status == db.TaskRunningStatus {
			data.NewJobs = append(data.NewJobs, runners.JobData{
				Username:        tsk.Username,
				IncomingVersion: tsk.IncomingVersion,
				Task:            tsk.Task,
				Template:        tsk.Template,
				Inventory:       tsk.Inventory,
				Repository:      tsk.Repository,
				Environment:     tsk.Environment,
			})
		} else {
			data.CurrentJobs = append(data.CurrentJobs, runners.JobState{
				ID:     tsk.Task.ID,
				Status: tsk.Task.Status,
			})
		}
	}

	helpers.WriteJSON(w, http.StatusOK, data)
}

func UpdateRunner(w http.ResponseWriter, r *http.Request) {
	var body runners.RunnerProgress

	if !helpers.Bind(w, r, &body) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid format",
		})
		return
	}

	taskPool := helpers.TaskPool(r)

	if body.Jobs == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for _, job := range body.Jobs {
		tsk := taskPool.GetTask(job.ID)

		if tsk == nil {
			// TODO: log
			continue
		}

		for _, logRecord := range job.LogRecords {
			tsk.Log2(logRecord.Message, logRecord.Time)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func RegisterRunner(w http.ResponseWriter, r *http.Request) {
	var register runners.RunnerRegistration

	if !helpers.Bind(w, r, &register) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid format",
		})
		return
	}

	if util.Config.RunnerRegistrationToken == "" || register.RegistrationToken != util.Config.RunnerRegistrationToken {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid registration token",
		})
		return
	}

	runner, err := helpers.Store(r).CreateRunner(db.Runner{
		State: db.RunnerActive,
	})

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Unexpected error",
		})
		return
	}

	res := runners.RunnerConfig{
		RunnerID: runner.ID,
		Token:    runner.Token,
	}

	helpers.WriteJSON(w, http.StatusOK, res)
}
