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

func GetIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(extractor.ID, matcher_id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matcher)
}

func GetIntegrationMatcherRefs(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(extractor.ID, matcher_id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	refs, err := helpers.Store(r).GetIntegrationMatcherRefs(matcher.ExtractorID, matcher.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func GetIntegrationMatchers(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	matchers, err := helpers.Store(r).GetIntegrationMatchers(extractor.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matchers)
}

func AddIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
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

	newMatcher, err := helpers.Store(r).CreateIntegrationMatcher(matcher)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, newMatcher)
}

func UpdateIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	var matcher db.IntegrationMatcher

	if !helpers.Bind(w, r, &matcher) {
		return
	}

	log.Info(fmt.Sprintf("Updating API Matcher %v for Extractor %v, matcher ID: %v", matcher_id, extractor.ID, matcher.ID))

	err = helpers.Store(r).UpdateIntegrationMatcher(matcher)
	log.Info(fmt.Sprintf("Err %s", err))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(extractor.ID, matcher_id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteIntegrationMatcher(extractor.ID, matcher.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Matcher failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
