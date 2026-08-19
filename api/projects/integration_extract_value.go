package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

func GetIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid IntegrationExtractValue ID",
		})
		return
	}

	integration := helpers.GetFromContext(r, "integration").(db.Integration)
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
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration := helpers.GetFromContext(r, "integration").(db.Integration)
	values, err := helpers.Store(r).GetIntegrationExtractValues(project.ID, helpers.QueryParams(r.URL), integration.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, values)
}

func AddIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration := helpers.GetFromContext(r, "integration").(db.Integration)

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
	project := helpers.GetFromContext(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}

	integration := helpers.GetFromContext(r, "integration").(db.Integration)

	var newValue db.IntegrationExtractValue
	if !helpers.Bind(w, r, &newValue) {
		return
	}

	newValue.ID = valueId
	newValue.IntegrationID = integration.ID

	err = helpers.Store(r).UpdateIntegrationExtractValue(project.ID, newValue)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetIntegrationExtractValueRefs(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	integration := helpers.GetFromContext(r, "integration").(db.Integration)
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
	project := helpers.GetFromContext(r, "project").(db.Project)
	valueId, err := helpers.GetIntParam("value_id", w, r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}

	integration := helpers.GetFromContext(r, "integration").(db.Integration)

	err = helpers.Store(r).DeleteIntegrationExtractValue(project.ID, valueId, integration.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Integration Extract Value failed to be deleted",
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
