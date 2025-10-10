package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

type RolesController struct {
	roleRepo db.RoleRepository
}

func NewRolesController(roleRepo db.RoleRepository) *RolesController {
	return &RolesController{
		roleRepo: roleRepo,
	}
}

func (c *RolesController) GetGlobalRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) GetRoles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) AddRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// Project-specific role methods
func (c *RolesController) GetProjectRoles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) GetProjectAndGlobalRoles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) AddProjectRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) GetProjectRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) UpdateProjectRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RolesController) DeleteProjectRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
