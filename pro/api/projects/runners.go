package projects

import (
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"net/http"
)

// NewProjectRunnerController creates a new ProjectRunnerController instance.
func NewProjectRunnerController() pro_interfaces.ProjectRunnerController {
	return &ProjectRunnerControllerImpl{}
}

type ProjectRunnerControllerImpl struct {
}

func (c *ProjectRunnerControllerImpl) GetRunners(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	runners, err := helpers.Store(r).GetRunners(project.ID, false, nil)

	if err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusOK, runners)
}

func (c *ProjectRunnerControllerImpl) AddRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) RunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (c *ProjectRunnerControllerImpl) GetRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) UpdateRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) DeleteRunner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) SetRunnerActive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) ClearRunnerCache(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *ProjectRunnerControllerImpl) GetRunnerTags(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	tags, err := helpers.Store(r).GetRunnerTags(project.ID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tags)
}
