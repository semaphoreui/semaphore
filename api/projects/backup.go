package projects

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	projectService "github.com/ansible-semaphore/semaphore/services/project"
	"github.com/gorilla/context"
)

func GetBackup(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	store := helpers.Store(r)

	backup, err := projectService.GetBackup(project.ID, store)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, backup)
}
