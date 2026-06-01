package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// NewProjectRunnerController creates a new ProjectRunnerController instance.
func NewProjectRunnerController(subscriptionService pro_interfaces.SubscriptionService) pro_interfaces.ProjectRunnerController {
	return &ProjectRunnerControllerImpl{}
}

type ProjectRunnerControllerImpl struct {
}

func (c *ProjectRunnerControllerImpl) GetRunners(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)

	projectRunners, err := store.GetRunners(project.ID, false, db.RunnerFilterIgnoreTags, nil)
	if err != nil {
		panic(err)
	}

	globalRunners, err := store.GetAllRunners(false, true, db.RunnerFilterHasAnyTag, nil)
	if err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusOK, mergeProjectRunnersList(projectRunners, globalRunners))
}

// mergeProjectRunnersList returns project runners plus active global runners that have tags.
func mergeProjectRunnersList(projectRunners, globalRunners []db.Runner) []db.Runner {
	result := make([]db.Runner, 0, len(projectRunners)+len(globalRunners))
	result = append(result, projectRunners...)
	for _, runner := range globalRunners {
		if !runner.Active || len(runner.Tags) == 0 {
			continue
		}
		result = append(result, runner)
	}
	return result
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
	helpers.WriteJSON(w, http.StatusOK, []any{})
}
