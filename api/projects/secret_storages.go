package projects

import (
	"fmt"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"net/http"
)

type SecretStorageController struct {
	secretRepo db.SecretStorageRepository
}

func SecretStorageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		storageID, err := helpers.GetIntParam("storage_id", w, r)
		if err != nil {
			return
		}

		key, err := helpers.Store(r).GetSecretStorage(project.ID, storageID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "secretStorage", key)
		next.ServeHTTP(w, r)
	})
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

func (c *SecretStorageController) GetSecretStorage(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	storages, err := c.secretRepo.GetSecretStorages(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
	}

	helpers.WriteJSON(w, http.StatusOK, storages)
}

func (c *SecretStorageController) Add(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	var storage db.SecretStorage

	if !helpers.Bind(w, r, &storage) {
		return
	}

	if storage.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	newStorage, err := c.secretRepo.CreateSecretStorage(storage)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newStorage.ProjectID,
		ObjectType:  db.EventKey,
		ObjectID:    newStorage.ID,
		Description: fmt.Sprintf("Secret storage %s has been created", storage.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newStorage)
}
