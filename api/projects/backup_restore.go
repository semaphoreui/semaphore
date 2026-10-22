package projects

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	projectService "github.com/semaphoreui/semaphore/services/project"
	log "github.com/sirupsen/logrus"
)

// BackupController serves project backup/restore. Workflows live outside
// db.Store (Pro feature, see db.WorkflowManager), so the workflow store is
// injected and threaded into the backup/restore routines. A restore creates
// integrations too, so it is held to the same license check as the
// integrations API (see IntegrationController).
type BackupController struct {
	workflowStore db.WorkflowManager
	pro           proFeatureGate
}

func NewBackupController(workflowStore db.WorkflowManager, subscription pro_interfaces.SubscriptionService) *BackupController {
	return &BackupController{workflowStore: workflowStore, pro: newProFeatureGate(subscription)}
}

// workflowIntegrationsAllowed refuses, before anything is created, a backup
// holding an integration that starts a workflow when the user lacks the
// workflow_integrations_management feature. The answer is written on refusal.
func (c *BackupController) workflowIntegrationsAllowed(w http.ResponseWriter, r *http.Request, user *db.User, backup *projectService.BackupFormat) bool {
	runsWorkflow := false
	for _, integration := range backup.Integration {
		if integration.Workflow != nil {
			runsWorkflow = true
			break
		}
	}
	if !runsWorkflow {
		return true
	}

	features, err := c.pro.features(user)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "backup_api",
		}).WithError(err).Error("Failed to read the subscription plan")
		helpers.WriteErrorStatus(w, "Failed to read the subscription plan", http.StatusInternalServerError)
		return false
	}
	if !features.WorkflowIntegrationsManagement {
		helpers.WriteErrorStatus(w, workflowIntegrationsProOnly, http.StatusForbidden)
		return false
	}
	return true
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
		w.WriteHeader(http.StatusForbidden)
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
		log.WithFields(log.Fields{
			"context": "backup_api",
		}).WithError(err).Error("failed to unmarshal backup")
		helpers.WriteError(w, err)
		return
	}

	store := helpers.Store(r)

	if err := backup.Verify(); err != nil {
		log.WithFields(log.Fields{
			"context": "backup_api",
		}).WithError(err).Error("failed to verify backup")
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !c.workflowIntegrationsAllowed(w, r, user, &backup) {
		return
	}

	var p *db.Project
	p, err := backup.Restore(*user, store, c.workflowStore)

	if err != nil {
		log.WithFields(log.Fields{
			"context": "backup_api",
		}).WithError(err).Error("failed to restore backup")
	}

	// Without a project nothing was created; with one, a failed step is a partial success.
	if err != nil && p == nil {
		helpers.WriteError(w, err)
		return
	}

	integrationAliases := len(backup.IntegrationAliases)
	for _, integration := range backup.Integration {
		integrationAliases += len(integration.Aliases)
	}

	event := audit.Event{
		Kind:      audit.ResourceProjectBackupRestore,
		Target:    audit.ResourceTarget(audit.TargetProject, p.ID, p.Name),
		ProjectID: p.ID,
	}
	metadata := audit.BackupRestoreMetadata{Objects: map[string]int{
		"templates":           len(backup.Templates),
		"repositories":        len(backup.Repositories),
		"host_configs":        len(backup.HostConfigs),
		"keys":                len(backup.Keys),
		"views":               len(backup.Views),
		"inventories":         len(backup.Inventories),
		"environments":        len(backup.Environments),
		"integrations":        len(backup.Integration),
		"integration_aliases": integrationAliases,
		"schedules":           len(backup.Schedules),
		"secret_storages":     len(backup.SecretStorages),
		"roles":               len(backup.Roles),
		"runners":             len(backup.Runners),
		"workflows":           len(backup.Workflows),
	}}

	if err != nil {
		event.Reason = audit.ReasonRestoreFailed
		metadata.Partial = true
	}
	event.Metadata = metadata
	helpers.Audit(r).Record(r.Context(), event)

	if err != nil {
		if _, ok := errors.AsType[*common_errors.ValidationError](err); ok {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, p)
}
