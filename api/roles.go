package api

import "net/http"

type RolesController struct {
}

func NewRolesController() *RolesController {
	return &RolesController{}
}

func (c *RolesController) GetRoles(w http.ResponseWriter, r *http.Request) {
}

func (c *RolesController) AddRole(w http.ResponseWriter, r *http.Request) {
}

func (c *RolesController) UpdateRole(w http.ResponseWriter, r *http.Request) {
}

func (c *RolesController) DeleteRole(w http.ResponseWriter, r *http.Request) {
}
