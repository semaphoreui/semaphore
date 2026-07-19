package projects

import (
	"errors"
	//	"strconv"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
)

func getIntergrationMatcherFromRequest(r *http.Request) (*db.Project, *db.IntegrationMatcher, error) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	matcherID, err := helpers.GetIntParamR("matcher_id", r)
	if err != nil {
		return nil, nil, err
	}

	integration := helpers.GetFromContext(r, "integration").(db.Integration)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(project.ID, matcherID, integration.ID)
	if err != nil {
		return nil, nil, err
	}

	return &project, &matcher, nil
}

func GetIntegrationMatcher(w http.ResponseWriter, r *http.Request) {

	_, matcher, err := getIntergrationMatcherFromRequest(r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matcher)
}

func GetIntegrationMatcherRefs(w http.ResponseWriter, r *http.Request) {

	project, matcher, err := getIntergrationMatcherFromRequest(r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	refs, err := helpers.Store(r).GetIntegrationMatcherRefs(project.ID, matcher.ID, matcher.IntegrationID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func GetIntegrationMatchers(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration := helpers.GetFromContext(r, "integration").(db.Integration)

	matchers, err := helpers.Store(r).GetIntegrationMatchers(project.ID, helpers.QueryParams(r.URL), integration.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, matchers)
}

func AddIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration := helpers.GetFromContext(r, "integration").(db.Integration)

	var matcher db.IntegrationMatcher
	if !helpers.Bind(w, r, &matcher) {
		return
	}

	if matcher.IntegrationID != integration.ID {
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
	project := helpers.GetFromContext(r, "project").(db.Project)
	matcherId, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}
	integration := helpers.GetFromContext(r, "integration").(db.Integration)

	var matcher db.IntegrationMatcher

	if !helpers.Bind(w, r, &matcher) {
		return
	}

	matcher.ID = matcherId
	matcher.IntegrationID = integration.ID

	log.Info(fmt.Sprintf("Updating API Matcher %v for Extractor %v, matcher ID: %v", matcherId, integration.ID, matcher.ID))

	err = helpers.Store(r).UpdateIntegrationMatcher(project.ID, matcher)
	log.Info(fmt.Sprintf("Err %s", err))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegrationMatcher(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	matcherId, err := helpers.GetIntParam("matcher_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Matcher ID",
		})
		return
	}

	integration := helpers.GetFromContext(r, "integration").(db.Integration)
	var matcher db.IntegrationMatcher
	matcher, err = helpers.Store(r).GetIntegrationMatcher(project.ID, matcherId, integration.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteIntegrationMatcher(project.ID, matcher.ID, integration.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Integration Matcher failed to be deleted",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
