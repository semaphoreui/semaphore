package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// GlobalRolesController is the community stub: global roles and their assignment to users are a Pro feature.
type GlobalRolesController struct{}

func NewGlobalRolesController(globalRoles db.GlobalRoleStore, store db.Store) *GlobalRolesController {
	return &GlobalRolesController{}
}

func (c *GlobalRolesController) GetGlobalPermissionCatalog(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, db.GlobalPermissionCatalog())
}

func (c *GlobalRolesController) GetGlobalRoles(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []db.GlobalRole{})
}

func (c *GlobalRolesController) GetGlobalRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) AddGlobalRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) UpdateGlobalRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) DeleteGlobalRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) GetGlobalRoleAssignments(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []db.GlobalRoleAssignment{})
}

func (c *GlobalRolesController) AddGlobalRoleAssignment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) DeleteGlobalRoleAssignment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *GlobalRolesController) GetEffectiveGlobalPermissions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
