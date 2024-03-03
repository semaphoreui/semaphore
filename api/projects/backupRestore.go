package projects

import (
	"net/http"

	log "github.com/sirupsen/logrus"
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

func Restore(w http.ResponseWriter, r *http.Request) {
	var backup projectService.BackupFormat
	var p *db.Project
	var err error

	if !helpers.Bind(w, r, &backup) {
		helpers.WriteJSON(w, http.StatusBadRequest, backup)
		return
	}
	store := helpers.Store(r)
	if err = backup.Verify(); err != nil {
		log.Error(err)
		helpers.WriteError(w, (err))
		return
	}
	if p, err = backup.Restore(store); err != nil {
		log.Error(err)
		helpers.WriteError(w, (err))
		return
	}
	helpers.WriteJSON(w, http.StatusOK, p)
}
