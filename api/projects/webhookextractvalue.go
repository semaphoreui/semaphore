package projects

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
)

func GetIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid IntegrationExtractValue ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(extractor.ID, value_id)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Failed to get IntegrationExtractValue, %v", err),
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, value)
}

func GetIntegrationExtractValues(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	values, err := helpers.Store(r).GetIntegrationExtractValues(extractor.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, values)
}

func AddIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	var value db.IntegrationExtractValue

	if !helpers.Bind(w, r, &value) {
		return
	}

	if value.ExtractorID != extractor.ID {
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

	newValue, err := helpers.Store(r).CreateIntegrationExtractValue(value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newValue)
}

func UpdateIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(extractor.ID, value_id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if !helpers.Bind(w, r, &value) {
		return
	}

	err = helpers.Store(r).UpdateIntegrationExtractValue(value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetIntegrationExtractValueRefs(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var value db.IntegrationExtractValue
	value, err = helpers.Store(r).GetIntegrationExtractValue(extractor.ID, value_id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	refs, err := helpers.Store(r).GetIntegrationExtractValueRefs(value.ExtractorID, value.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func DeleteIntegrationExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Extract Value failed to be deleted",
		})
		return
	}

	err = helpers.Store(r).DeleteIntegrationExtractValue(extractor.ID, value_id)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Extract Value failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
