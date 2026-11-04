package projects

import (
	"errors"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
)

// workflowPermsEnterpriseOnly is the 403 body when a grant is written on a plan without custom roles.
const workflowPermsEnterpriseOnly = "Workflow permissions are only available for Enterprise users."

// WorkflowPermController grants roles on one workflow, the twin of the template permission handlers.
type WorkflowPermController struct {
	workflowRoles db.WorkflowRoleManager
	pro           proFeatureGate
}

func NewWorkflowPermController(workflowRoles db.WorkflowRoleManager, subscription pro_interfaces.SubscriptionService) *WorkflowPermController {
	return &WorkflowPermController{workflowRoles: workflowRoles, pro: newProFeatureGate(subscription)}
}

// grantsAllowed answers the request itself when the plan has no custom roles.
func (c *WorkflowPermController) grantsAllowed(w http.ResponseWriter, r *http.Request, projectID int) bool {
	features, err := c.pro.features(helpers.UserFromContext(r))
	if err != nil {
		log.WithFields(log.Fields{
			"context":    "workflow_perms",
			"project_id": projectID,
		}).WithError(err).Error("Failed to read the subscription plan")
		helpers.WriteErrorStatus(w, "Failed to read the subscription plan", http.StatusInternalServerError)
		return false
	}
	if !features.CustomRolesManagement {
		helpers.WriteErrorStatus(w, workflowPermsEnterpriseOnly, http.StatusForbidden)
		return false
	}
	return true
}

func (c *WorkflowPermController) GetWorkflowPerms(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	perms, err := c.workflowRoles.GetWorkflowRoles(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, perms)
}

func (c *WorkflowPermController) AddWorkflowPerm(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	if !c.grantsAllowed(w, r, workflow.ProjectID) {
		return
	}

	var perm db.WorkflowRolePerm
	if !helpers.Bind(w, r, &perm) {
		return
	}
	if perm.Permissions&^db.WorkflowGrantPermissions != 0 {
		helpers.WriteErrorStatus(w, "Permissions are not workflow permissions", http.StatusBadRequest)
		return
	}
	// A verb on a workflow includes seeing it.
	if perm.Permissions != 0 {
		perm.Permissions |= db.CanReadProject
	}

	role, err := helpers.Store(r).GetProjectOrSharedRoleBySlug(workflow.ProjectID, perm.RoleSlug)
	if errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, "Role does not exist in this project", http.StatusBadRequest)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	perm.RoleID = role.ID
	perm.ProjectID = workflow.ProjectID
	perm.WorkflowID = workflow.ID

	newPerm, err := c.workflowRoles.CreateWorkflowRole(perm)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	newPerm.RoleSlug = role.Slug

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.IAMWorkflowPermissionCreate,
		Target:    &audit.Target{Type: audit.TargetWorkflowPermission, ID: strconv.Itoa(newPerm.ID)},
		ProjectID: workflow.ProjectID,
		Metadata: audit.WorkflowPermissionMetadata{
			WorkflowID:  workflow.ID,
			Role:        audit.NewRoleRef(role),
			Permissions: audit.PermissionNames(newPerm.Permissions),
		},
	})

	helpers.WriteJSON(w, http.StatusCreated, newPerm)
}

func (c *WorkflowPermController) UpdateWorkflowPerm(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	permID, ok := helpers.GetIntParamOrAbort("perm_id", w, r)
	if !ok {
		return
	}
	if !c.grantsAllowed(w, r, workflow.ProjectID) {
		return
	}

	var perm db.WorkflowRolePerm
	if !helpers.Bind(w, r, &perm) {
		return
	}
	if perm.Permissions&^db.WorkflowGrantPermissions != 0 {
		helpers.WriteErrorStatus(w, "Permissions are not workflow permissions", http.StatusBadRequest)
		return
	}
	// A verb on a workflow includes seeing it.
	if perm.Permissions != 0 {
		perm.Permissions |= db.CanReadProject
	}

	perm.ID = permID
	perm.ProjectID = workflow.ProjectID
	perm.WorkflowID = workflow.ID

	// The update changes only permissions, so the role comes from the stored grant.
	stored, err := c.workflowRoles.GetWorkflowRole(workflow.ProjectID, workflow.ID, permID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	if err = c.workflowRoles.UpdateWorkflowRole(perm); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.IAMWorkflowPermissionUpdate,
		Target:    &audit.Target{Type: audit.TargetWorkflowPermission, ID: strconv.Itoa(permID)},
		ProjectID: workflow.ProjectID,
		Metadata: audit.WorkflowPermissionMetadata{
			WorkflowID:  workflow.ID,
			Role:        audit.NewRoleRef(db.Role{ID: stored.RoleID, Slug: stored.RoleSlug}),
			Permissions: audit.PermissionNames(perm.Permissions),
		},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (c *WorkflowPermController) DeleteWorkflowPerm(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	permID, ok := helpers.GetIntParamOrAbort("perm_id", w, r)
	if !ok {
		return
	}
	if !c.grantsAllowed(w, r, workflow.ProjectID) {
		return
	}

	err := c.workflowRoles.DeleteWorkflowRole(workflow.ProjectID, workflow.ID, permID)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.IAMWorkflowPermissionDelete,
		Target:    &audit.Target{Type: audit.TargetWorkflowPermission, ID: strconv.Itoa(permID)},
		ProjectID: workflow.ProjectID,
		Metadata:  audit.WorkflowPermissionMetadata{WorkflowID: workflow.ID},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (c *WorkflowPermController) GetWorkflowPerm(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	permID, ok := helpers.GetIntParamOrAbort("perm_id", w, r)
	if !ok {
		return
	}

	perm, err := c.workflowRoles.GetWorkflowRole(workflow.ProjectID, workflow.ID, permID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, perm)
}
