package projects

import (
	"io"
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	projectService "github.com/ansible-semaphore/semaphore/services/project"
	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
)

func GetBackup(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	store := helpers.Store(r)

	backup, err := projectService.GetBackup(project.ID, store)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	str, err := backup.Marshal()
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(str))
}

func Restore(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	var backup projectService.BackupFormat

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, r.Body); err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	if err := backup.Unmarshal(buf.String()); err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	store := helpers.Store(r)
	if err := backup.Verify(); err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	var p *db.Project
	p, err := backup.Restore(*user, store)

	if err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, p)
}
