package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
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
	roleID, err := helpers.GetIntParam("role_id", w, r)
	if err != nil {
		return
	}

	role, err := c.roleRepo.GetRole(roleID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, role)
}

func (c *RolesController) GetRoles(w http.ResponseWriter, r *http.Request) {

	roles, err := c.roleRepo.GetRoles()
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, roles)
}

func (c *RolesController) AddRole(w http.ResponseWriter, r *http.Request) {
	var role db.Role
	if !helpers.Bind(w, r, &role) {
		return
	}

	newRole, err := c.roleRepo.CreateRole(role)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newRole)
}

func (c *RolesController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := helpers.GetIntParam("role_id", w, r)
	if err != nil {
		return
	}

	var role db.Role
	if !helpers.Bind(w, r, &role) {
		return
	}

	role.ID = roleID

	if err := c.roleRepo.UpdateRole(role); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *RolesController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := helpers.GetIntParam("role_id", w, r)
	if err != nil {
		return
	}

	if err := c.roleRepo.DeleteRole(roleID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
