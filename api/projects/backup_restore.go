package projects

import (
	"io"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	projectService "github.com/semaphoreui/semaphore/services/project"
	log "github.com/sirupsen/logrus"
)

// BackupController serves project backup/restore. Workflows live outside
// db.Store (Pro feature, see db.WorkflowManager), so the workflow store is
// injected and threaded into the backup/restore routines.
type BackupController struct {
	workflowStore db.WorkflowManager
}

func NewBackupController(workflowStore db.WorkflowManager) *BackupController {
	return &BackupController{workflowStore: workflowStore}
}

func (c *BackupController) GetBackup(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	store := helpers.Store(r)

	backup, err := projectService.GetBackup(project.ID, store, c.workflowStore)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	str, err := backup.Marshal()
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceProjectBackupExport,
		Target:    audit.ResourceTarget(audit.TargetProject, project.ID, project.Name),
		ProjectID: project.ID,
	})

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(str))
}

func (c *BackupController) Restore(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	if !user.Admin && !util.Config.NonAdminCanCreateProject {
		log.Warn(user.Username + " is not permitted to restore the project")
		helpers.RecordDenied(r, "admin", 0)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var backup projectService.BackupFormat

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, r.Body); err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	str := buf.String()

	if err := backup.Unmarshal(str); err != nil {
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
	p, err := backup.Restore(*user, store, c.workflowStore)

	if err != nil {
		log.Error(err)
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceProjectBackupRestore,
		Target:    audit.ResourceTarget(audit.TargetProject, p.ID, p.Name),
		ProjectID: p.ID,
		Metadata: audit.BackupRestoreMetadata{Objects: map[string]int{
			"templates":       len(backup.Templates),
			"repositories":    len(backup.Repositories),
			"host_configs":    len(backup.HostConfigs),
			"keys":            len(backup.Keys),
			"views":           len(backup.Views),
			"inventories":     len(backup.Inventories),
			"environments":    len(backup.Environments),
			"integrations":    len(backup.Integration),
			"schedules":       len(backup.Schedules),
			"secret_storages": len(backup.SecretStorages),
			"roles":           len(backup.Roles),
			"runners":         len(backup.Runners),
			"workflows":       len(backup.Workflows),
		}},
	})

	helpers.WriteJSON(w, http.StatusOK, p)
}
