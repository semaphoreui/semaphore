package projects

import (
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"net/http"
)

type terraformInventoryController struct{}

func NewTerraformInventoryController(terraformRepo db.TerraformStore) pro_interfaces.TerraformInventoryController {
	return &terraformInventoryController{}
}

func (c *terraformInventoryController) GetTerraformInventoryAliases(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []string{})
}

func (c *terraformInventoryController) AddTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *terraformInventoryController) GetTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *terraformInventoryController) DeleteTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *terraformInventoryController) SetTerraformInventoryAliasAccessKey(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *terraformInventoryController) GetTerraformInventoryStates(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []string{})
}

func (c *terraformInventoryController) GetTerraformInventoryLatestState(w http.ResponseWriter, r *http.Request) {
	helpers.WriteErrorStatus(w, "No state found", http.StatusNotFound)
}

func (c *terraformInventoryController) GetTerraformInventoryState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *terraformInventoryController) DeleteTerraformInventoryState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
