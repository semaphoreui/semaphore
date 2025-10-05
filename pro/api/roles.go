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

func (c *RolesController) GetRole(w http.ResponseWriter, r *http.Request) {
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
