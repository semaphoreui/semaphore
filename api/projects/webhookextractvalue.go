package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"github.com/gorilla/context"
)

func GetWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid WebhookExtractValue ID",
		})
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
    project := context.Get(r, "project").(db.Project)
    webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

    var value db.WebhookExtractValue

    if !helpers.Bind(w, r, &value) {
        return
    }

    if value.ExtractorID != extractor.ID {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
            "error": "Extractor ID in body and URL must be the same",
        })
        return
    }

    if err := value.Validate(); err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
            "error": err.Error(),
            })
        return
    }

    newValue, err := helpers.Store(r).CreateWebhookExtractValue(value)

    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    user := context.Get(r, "user").(*db.User)
    objType := db.EventWebhookExtractValue
    desc := "Webhook Extracted Value" + newValue.Name + " created"
    _, err = helpers.Store(r).CreateEvent(db.Event{
        UserID:      &user.ID,
        ProjectID:   &project.ID,
        WebhookID:   &webhook_id,
        ExtractorID: &extractor.ID,
        ObjectType:  &objType,
        ObjectID:    &value.ID,
        Description: &desc,
    })

    w.WriteHeader(http.StatusNoContent)
}

func UpdateWebhookExtractValue(w http.ResponseWriter, r *http.Request) {
    value_id, err := helpers.GetIntParam("value_id", w, r)

    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Invalid Value ID",
        })
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

    user := context.Get(r, "user").(*db.User)
    webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	desc := "WebhookExtractValue (" + value.String() + ") updated"

	objType := db.EventWebhookExtractValue

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		
		WebhookID:   &webhook_id,
		ExtractorID: &extractor.ID,
		Description: &desc,
		ObjectID:    &value.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetWebhookExtractValueRefs(w http.ResponseWriter, r *http.Request) {
	value_id, err := helpers.GetIntParam("value_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Value ID",
		})
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
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var value db.WebhookExtractValue
	value, err = helpers.Store(r).GetWebhookExtractValue(extractor.ID, value_id)

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

	user := context.Get(r, "user").(*db.User)
	project := context.Get(r, "project").(db.Project)
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	desc := "Webhook Extract Value (" + value.String() + ") deleted"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		WebhookID:   &webhook_id,
		ExtractorID: &extractor.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusNoContent)
}
