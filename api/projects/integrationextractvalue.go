package projects

import (
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
)

func GetIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid IntegrationExtractValue ID",
		})
		return
	}

	integration := context.Get(r, "integration").(db.Integration)
	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(project.ID, valueId, integration.ID)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Failed to get IntegrationExtractValue, %v", err),
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, value)
}

func GetIntegrationExtractValues(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	integration := context.Get(r, "integration").(db.Integration)
	values, err := helpers.Store(r).GetIntegrationExtractValues(project.ID, helpers.QueryParams(r.URL), integration.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, values)
}

func AddIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	integration := context.Get(r, "integration").(db.Integration)

	var value db.IntegrationExtractValue

	if !helpers.Bind(w, r, &value) {
		return
	}

	if value.IntegrationID != integration.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Extractor ID in body and URL must be the same",
		})
		return
	}

	if err := value.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newValue, err := helpers.Store(r).CreateIntegrationExtractValue(project.ID, value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newValue)
}

func UpdateIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	integration := context.Get(r, "integration").(db.Integration)

	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(project.ID, valueId, integration.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if !helpers.Bind(w, r, &value) {
		return
	}

	if value.ID != valueId {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Value ID in body and URL must be the same",
		})
		return
	}

	err = helpers.Store(r).UpdateIntegrationExtractValue(project.ID, value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetIntegrationExtractValueRefs(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	integration := context.Get(r, "integration").(db.Integration)
	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(project.ID, valueId, integration.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	refs, err := helpers.Store(r).GetIntegrationExtractValueRefs(project.ID, value.ID, value.IntegrationID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func DeleteIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	integration := context.Get(r, "integration").(db.Integration)

	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Extract Value failed to be deleted",
		})
		return
	}

	err = helpers.Store(r).DeleteIntegrationExtractValue(project.ID, valueId, integration.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Extract Value failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
