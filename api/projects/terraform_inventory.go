//go:build !pro

package projects

import (
	"github.com/semaphoreui/semaphore/api/helpers"
	"net/http"
)

func GetTerraformInventoryAliases(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []string{})
}

func AddTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func GetTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func DeleteTerraformInventoryAlias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func SetTerraformInventoryAliasAccessKey(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func GetTerraformInventoryStates(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []string{})
}

func GetTerraformInventoryLatestState(w http.ResponseWriter, r *http.Request) {
	helpers.WriteErrorStatus(w, "No state found", http.StatusNotFound)
}

func GetTerraformInventoryState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func DeleteTerraformInventoryState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
