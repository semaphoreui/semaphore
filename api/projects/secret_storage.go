package projects

import (
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"net/http"
)

type SecretStorageController struct {
	secretRepo db.SecretStorageRepository
}

func NewSecretStorageController(secretRepo db.SecretStorageRepository) *SecretStorageController {
	return &SecretStorageController{
		secretRepo: secretRepo,
	}
}

func (c *SecretStorageController) GetSecretStorages(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	storages, err := c.secretRepo.GetSecretStorages(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
	}

	helpers.WriteJSON(w, http.StatusOK, storages)
}
