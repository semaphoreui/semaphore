package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
)

// GlobalRolesController assigns global roles to users. The roles themselves are managed by the roles API.
type GlobalRolesController struct{}

func NewGlobalRolesController() *GlobalRolesController {
	return &GlobalRolesController{}
}

func (c *GlobalRolesController) GetGlobalPermissionCatalog(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, db.GlobalPermissionCatalog())
}

// targetUser reads the user named in the path. The route is only reachable with CanManageGlobalRoles, so
// nothing else is checked here.
func (c *GlobalRolesController) targetUser(w http.ResponseWriter, r *http.Request) (db.User, bool) {
	userID, ok := helpers.GetIntParamOrAbort("user_id", w, r)
	if !ok {
		return db.User{}, false
	}
	user, err := helpers.Store(r).GetUser(userID)
	if err != nil {
		helpers.WriteError(w, err)
		return db.User{}, false
	}
	return user, true
}

func (c *GlobalRolesController) GetGlobalRoleAssignments(w http.ResponseWriter, r *http.Request) {
	user, ok := c.targetUser(w, r)
	if !ok {
		return
	}
	assignments, err := helpers.Store(r).GetGlobalRoleAssignments(user.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, assignments)
}

func (c *GlobalRolesController) AddGlobalRoleAssignment(w http.ResponseWriter, r *http.Request) {
	user, ok := c.targetUser(w, r)
	if !ok {
		return
	}
	var input struct {
		RoleID int `json:"role_id"`
	}
	if !helpers.Bind(w, r, &input) {
		return
	}
	if input.RoleID <= 0 {
		helpers.WriteErrorStatus(w, "role_id is required", http.StatusBadRequest)
		return
	}
	assignment, err := helpers.Store(r).CreateGlobalRoleAssignment(user.ID, input.RoleID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.Audit(r).Record(r.Context(), globalRoleEvent(audit.IAMGlobalRoleAssign, user, assignment))
	helpers.WriteJSON(w, http.StatusCreated, assignment)
}

func (c *GlobalRolesController) DeleteGlobalRoleAssignment(w http.ResponseWriter, r *http.Request) {
	user, ok := c.targetUser(w, r)
	if !ok {
		return
	}
	assignmentID, ok := helpers.GetIntParamOrAbort("assignment_id", w, r)
	if !ok {
		return
	}
	assignment, err := helpers.Store(r).GetGlobalRoleAssignment(user.ID, assignmentID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	if err = helpers.Store(r).DeleteGlobalRoleAssignment(user.ID, assignmentID); err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.Audit(r).Record(r.Context(), globalRoleEvent(audit.IAMGlobalRoleUnassign, user, assignment))
	w.WriteHeader(http.StatusNoContent)
}

func (c *GlobalRolesController) GetEffectiveGlobalPermissions(w http.ResponseWriter, r *http.Request) {
	user, ok := c.targetUser(w, r)
	if !ok {
		return
	}
	assignments, err := helpers.Store(r).GetGlobalRoleAssignments(user.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, db.ExplainEffectiveGlobalPermissions(user.Admin, assignments))
}

func globalRoleEvent(kind audit.Kind, user db.User, assignment db.GlobalRoleAssignment) audit.Event {
	return audit.Event{
		Kind:   kind,
		Target: audit.UserTarget(user.ID, user.Username),
		Metadata: audit.GlobalRoleMetadata{
			Role:              audit.RoleRef{ID: assignment.RoleID, Slug: assignment.RoleSlug},
			GlobalPermissions: audit.GlobalPermissionNames(assignment.GlobalPermissions),
		},
	}
}
