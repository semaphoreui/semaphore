package projects

import (
	//	"strconv"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
)

func GetWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
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
		return
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
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Extractor ID in body and URL must be the same",
		})
		return
	}

	err := matcher.Validate()

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newMatcher, err := helpers.Store(r).CreateWebhookMatcher(matcher)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, newMatcher)
}

func UpdateWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
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

	w.WriteHeader(http.StatusNoContent)
}

func DeleteWebhookMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.WebhookExtractor)
	var matcher db.WebhookMatcher
	matcher, err = helpers.Store(r).GetWebhookMatcher(extractor.ID, matcher_id)

	err = helpers.Store(r).DeleteWebhookMatcher(extractor.ID, matcher.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook Matcher failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
