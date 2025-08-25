package projects

import (
	"fmt"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	"net/http"
)

type SecretStorageController struct {
	secretRepo           db.SecretStorageRepository
	secretStorageService server.SecretStorageService
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

func NewSecretStorageController(
	secretRepo db.SecretStorageRepository,
	secretStorageService server.SecretStorageService,

) *SecretStorageController {
	return &SecretStorageController{
		secretRepo:           secretRepo,
		secretStorageService: secretStorageService,
	}
}

func (c *SecretStorageController) GetRefs(w http.ResponseWriter, r *http.Request) {
	key := helpers.GetFromContext(r, "secretStorage").(db.SecretStorage)
	refs, err := helpers.Store(r).GetSecretStorageRefs(key.ProjectID, key.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func (c *SecretStorageController) GetSecretStorages(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	storages, err := c.secretStorageService.GetSecretStorages(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
	}

	helpers.WriteJSON(w, http.StatusOK, storages)
}

func (c *SecretStorageController) GetSecretStorage(w http.ResponseWriter, r *http.Request) {
	storage := helpers.GetFromContext(r, "secretStorage").(db.SecretStorage)

	helpers.WriteJSON(w, http.StatusOK, storage)
}

func (c *SecretStorageController) Update(w http.ResponseWriter, r *http.Request) {
	oldStorage := helpers.GetFromContext(r, "secretStorage").(db.SecretStorage)

	var storage db.SecretStorage
	if !helpers.Bind(w, r, &storage) {
		return
	}

	if storage.ID != oldStorage.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Secret storage id in URL and in body must be the same",
		})
		return
	}

	if storage.ProjectID != oldStorage.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "You can not move secret storage to other project",
		})
		return
	}

	err := c.secretStorageService.Update(storage)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldStorage.ProjectID,
		ObjectType:  db.EventSchedule,
		ObjectID:    oldStorage.ID,
		Description: fmt.Sprintf("Secret storage with ID %d has been updated", storage.ID),
	})

	helpers.WriteJSON(w, http.StatusOK, storage)
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

	newStorage, err := c.secretStorageService.Create(storage)

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

func (c *SecretStorageController) Remove(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	storageID, err := helpers.GetIntParam("storage_id", w, r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = c.secretStorageService.Delete(project.ID, storageID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusNoContent, nil)
}
