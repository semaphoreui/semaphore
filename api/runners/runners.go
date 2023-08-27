package runners

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
)

func RunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		runner, err := helpers.RunnerPool(r).GetRunner(0)

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
	runner := context.Get(r, "runner").(tasks.RemoteRunner)

	// TODO: get runner data:
	// - template
	// - inventory
	// - environment
	// - keys

	helpers.WriteJSON(w, http.StatusOK, runner)
}

func UpdateRunner(w http.ResponseWriter, r *http.Request) {
	//
}

func RegisterRunner(w http.ResponseWriter, r *http.Request) {
	var register struct {
		RegistrationToken string `json:"registration_token" binding:"required"`
	}

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

	helpers.WriteJSON(w, http.StatusOK, runner)
}
