package projects

import (
	//	"strconv"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"github.com/gorilla/context"
)

func GetWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
	}
	
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var matcher db.WebhookMatcher
	matcher, err = helpers.Store(r).GetWebhookMatcher(extractor.ID, matcher_id)

	helpers.WriteJSON(w, http.StatusOK, matcher)
}

func GetWebhookMatcherRefs(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var matcher db.WebhookMatcher
	matcher, err = helpers.Store(r).GetWebhookMatcher(extractor.ID, matcher_id)

	refs, err := helpers.Store(r).GetWebhookMatcherRefs(matcher.ExtractorID, matcher.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}


func GetWebhookMatchers(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	matchers, err := helpers.Store(r).GetWebhookMatchers(extractor.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matchers)
}

func AddWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var matcher db.WebhookMatcher
	if !helpers.Bind(w, r, &matcher) {
		return
	}

	if matcher.ExtractorID != extractor.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
			"error": "Extractor ID in body and URL must be the same",
		})
		return
	}

	err := matcher.Validate()

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
			"error": err.Error(),
	    })
		return
	}

	newMatcher, err := helpers.Store(r).CreateWebhookMatcher(matcher)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
	project := context.Get(r, "project").(db.Project)

	objType := db.EventWebhookMatcher
	desc := "Webhook Matcher " + matcher.Name + " created"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		WebhookID:   &webhook_id,
		ExtractorID: &extractor.ID,
		ObjectType:  &objType,
		ObjectID:    &newMatcher.ID,
		Description: &desc,
	})

	helpers.WriteJSON(w, http.StatusOK, newMatcher)
}

func UpdateWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
	}
	extractor := context.Get(r, "extractor").(db.WebhookExtractor)

	var matcher db.WebhookMatcher

	if !helpers.Bind(w, r, &matcher) {
		return
	}

	log.Info(fmt.Sprintf("Updating API Matcher %v for Extractor %v, matcher ID: %v", matcher_id, extractor.ID, matcher.ID))

	err = helpers.Store(r).UpdateWebhookMatcher(matcher)
	log.Info(fmt.Sprintf("Err %s", err))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	desc := "WebhookMatcher (" + matcher.String() + ") updated"

	objType := db.EventWebhookMatcher

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		WebhookID:   &webhook_id,
		ExtractorID: &extractor.ID,
		Description: &desc,
		ObjectID:    &matcher.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
	}

	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var matcher db.WebhookMatcher
	matcher, err = helpers.Store(r).GetWebhookMatcher(extractor.ID, matcher_id)


	err = helpers.Store(r).DeleteWebhookMatcher(extractor.ID, matcher.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook Matcher failed to be deleted",
		})
	}

	user := context.Get(r, "user").(*db.User)
	project := context.Get(r, "project").(db.Project)
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
  
	desc := "Webhook Matcher (" + matcher.String() + ") deleted"
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
