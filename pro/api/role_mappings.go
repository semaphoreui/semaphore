package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

type RoleMappingsController struct{}

func NewRoleMappingsController(mappings db.RoleMappingStore, store db.Store) *RoleMappingsController {
	return &RoleMappingsController{}
}

func (c *RoleMappingsController) GetMappings(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []string{})
}

func (c *RoleMappingsController) AddMapping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *RoleMappingsController) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
