//go:build !pro

package projects

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

func GetRunners(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	runners, err := helpers.Store(r).GetRunners(project.ID, false)

	if err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusOK, runners)
}

func AddRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func RunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func UpdateRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func DeleteRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func SetRunnerActive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
