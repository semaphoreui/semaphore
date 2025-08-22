package api

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	"net/http"
)

type TerraformController struct {
	encryptionServices server.AccessKeyEncryptionService
}

func NewTerraformController(
	encryptionServices server.AccessKeyEncryptionService,
	terraformRepo db.TerraformStore,
	keyRepo db.AccessKeyManager,
) *TerraformController {
	return &TerraformController{
		encryptionServices: encryptionServices,
	}
}

func (c *TerraformController) TerraformInventoryAliasMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (c *TerraformController) GetTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *TerraformController) AddTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *TerraformController) LockTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *TerraformController) UnlockTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
