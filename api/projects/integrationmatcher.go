package projects

import (
	//	"strconv"
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
)

func GetIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	matcher_id, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(project.ID, matcher_id, extractor.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matcher)
}

func GetIntegrationMatcherRefs(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	matcherId, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(project.ID, matcherId, extractor.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	refs, err := helpers.Store(r).GetIntegrationMatcherRefs(project.ID, matcher.ID, matcher.ExtractorID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func GetIntegrationMatchers(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	matchers, err := helpers.Store(r).GetIntegrationMatchers(project.ID, helpers.QueryParams(r.URL), extractor.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matchers)
}

func AddIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
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

	newMatcher, err := helpers.Store(r).CreateIntegrationMatcher(project.ID, matcher)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, newMatcher)
}

func UpdateIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	matcherId, err := helpers.GetIntParam("matcher_id", w, r)

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

	log.Info(fmt.Sprintf("Updating API Matcher %v for Extractor %v, matcher ID: %v", matcherId, extractor.ID, matcher.ID))

	err = helpers.Store(r).UpdateIntegrationMatcher(project.ID, matcher)
	log.Info(fmt.Sprintf("Err %s", err))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	matcherId, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(project.ID, matcherId, extractor.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteIntegrationMatcher(project.ID, matcher.ID, extractor.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Matcher failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
