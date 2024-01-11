package projects

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
)

func GetWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid WebhookExtractValue ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var value db.WebhookExtractValue
	value, err = helpers.Store(r).GetWebhookExtractValue(extractor.ID, value_id)

	helpers.WriteJSON(w, http.StatusOK, value)
}

func GetWebhookExtractValues(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	values, err := helpers.Store(r).GetWebhookExtractValues(extractor.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, values)
}

func AddWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	var value db.WebhookExtractValue

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

	_, err := helpers.Store(r).CreateWebhookExtractValue(value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	var value db.WebhookExtractValue
	value, err = helpers.Store(r).GetWebhookExtractValue(extractor.ID, value_id)

	if !helpers.Bind(w, r, &value) {
		return
	}

	err = helpers.Store(r).UpdateWebhookExtractValue(value)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetWebhookExtractValueRefs(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var value db.WebhookExtractValue
	value, err = helpers.Store(r).GetWebhookExtractValue(extractor.ID, value_id)

	refs, err := helpers.Store(r).GetWebhookExtractValueRefs(value.ExtractorID, value.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func DeleteWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook Extract Value failed to be deleted",
		})
		return
	}

	err = helpers.Store(r).DeleteWebhookExtractValue(extractor.ID, value_id)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook Extract Value failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
