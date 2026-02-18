package pro_interfaces

import (
	"net/http"
)

type TerraformInventoryController interface {
	GetTerraformInventoryAliases(w http.ResponseWriter, r *http.Request)
	AddTerraformInventoryAlias(w http.ResponseWriter, r *http.Request)
	GetTerraformInventoryAlias(w http.ResponseWriter, r *http.Request)
	DeleteTerraformInventoryAlias(w http.ResponseWriter, r *http.Request)
	SetTerraformInventoryAliasAccessKey(w http.ResponseWriter, r *http.Request)
	GetTerraformInventoryStates(w http.ResponseWriter, r *http.Request)
	GetTerraformInventoryLatestState(w http.ResponseWriter, r *http.Request)
	GetTerraformInventoryState(w http.ResponseWriter, r *http.Request)
	DeleteTerraformInventoryState(w http.ResponseWriter, r *http.Request)
}
