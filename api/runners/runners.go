package runners

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

func RunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("X-Runner-Token")

		if token == "" {
			helpers.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		store := helpers.Store(r)

		runner, err := store.GetGlobalRunnerByToken(token)

		if err != nil {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "Runner not found",
			})
			return
		}

		if runner.Token != token {
			helpers.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		context.Set(r, "runner", runner)
		next.ServeHTTP(w, r)
	})
}

func GetRunner(w http.ResponseWriter, r *http.Request) {
	runner := context.Get(r, "runner").(db.Runner)

	data := runners.RunnerState{
		AccessKeys: make(map[int]db.AccessKey),
	}

	tasks := helpers.TaskPool(r).GetRunningTasks()

	for _, tsk := range tasks {
		if tsk.RunnerID != runner.ID {
			continue
		}

		if tsk.Task.Status == task_logger.TaskStartingStatus {

			data.NewJobs = append(data.NewJobs, runners.JobData{
				Username:            tsk.Username,
				IncomingVersion:     tsk.IncomingVersion,
				Task:                tsk.Task,
				Template:            tsk.Template,
				Inventory:           tsk.Inventory,
				InventoryRepository: tsk.Inventory.Repository,
				Repository:          tsk.Repository,
				Environment:         tsk.Environment,
			})

			if tsk.Inventory.SSHKeyID != nil {
				err := tsk.Inventory.SSHKey.DeserializeSecret()
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[*tsk.Inventory.SSHKeyID] = tsk.Inventory.SSHKey
			}

			if tsk.Inventory.BecomeKeyID != nil {
				err := tsk.Inventory.BecomeKey.DeserializeSecret()
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[*tsk.Inventory.BecomeKeyID] = tsk.Inventory.BecomeKey
			}

			if tsk.Template.Vaults != nil {
				for _, vault := range tsk.Template.Vaults {
					err := vault.Vault.DeserializeSecret()
					if err != nil {
						// TODO: return error
					}
					data.AccessKeys[vault.VaultKeyID] = *vault.Vault
				}
			}

			if tsk.Inventory.RepositoryID != nil {
				err := tsk.Inventory.Repository.SSHKey.DeserializeSecret()
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[tsk.Inventory.Repository.SSHKeyID] = tsk.Inventory.Repository.SSHKey
			}

			data.AccessKeys[tsk.Repository.SSHKeyID] = tsk.Repository.SSHKey

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

	runner := context.Get(r, "runner").(db.Runner)

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

		if tsk.RunnerID != runner.ID {
			// TODO: add error message
			continue
		}

		for _, logRecord := range job.LogRecords {
			tsk.LogWithTime(logRecord.Time, logRecord.Message)
		}

		tsk.SetStatus(job.Status)
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
		Webhook:          register.Webhook,
		MaxParallelTasks: register.MaxParallelTasks,
	})

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Unexpected error",
		})
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	res.Token = runner.Token

	helpers.WriteJSON(w, http.StatusOK, res)
}

func UnregisterRunner(w http.ResponseWriter, r *http.Request) {

	runner := context.Get(r, "runner").(db.Runner)

	err := helpers.Store(r).DeleteGlobalRunner(runner.ID)

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Unknown error",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
