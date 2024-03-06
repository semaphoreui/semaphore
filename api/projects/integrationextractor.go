package projects

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"

	"github.com/gorilla/context"
)

func IntegrationExtractorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		extractorId, err := helpers.GetIntParam("extractor_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid extractor ID",
			})
			return
		}

		integration := context.Get(r, "integration").(db.Integration)
		var extractor db.IntegrationExtractor
		extractor, err = helpers.Store(r).GetIntegrationExtractor(project.ID, extractorId, integration.ID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "extractor", extractor)
		next.ServeHTTP(w, r)
	})
}

func GetIntegrationExtractor(w http.ResponseWriter, r *http.Request) {
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	helpers.WriteJSON(w, http.StatusOK, extractor)
}

func GetIntegrationExtractors(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	integration := context.Get(r, "integration").(db.Integration)
	extractors, err := helpers.Store(r).GetIntegrationExtractors(project.ID, helpers.QueryParams(r.URL), integration.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, extractors)
}

func AddIntegrationExtractor(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	integration := context.Get(r, "integration").(db.Integration)

	var extractor db.IntegrationExtractor

	if !helpers.Bind(w, r, &extractor) {
		return
	}

	if extractor.IntegrationID != integration.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Integration ID in body and URL must be the same",
		})
		return
	}

	if err := extractor.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newIntegrationExtractor, err := helpers.Store(r).CreateIntegrationExtractor(project.ID, extractor)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newIntegrationExtractor)

}

func UpdateIntegrationExtractor(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	oldExtractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	var extractor db.IntegrationExtractor

	if !helpers.Bind(w, r, &extractor) {
		return
	}

	if extractor.ID != oldExtractor.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "IntegrationExtractor ID in body and URL must be the same",
		})
		return
	}

	if extractor.IntegrationID != oldExtractor.IntegrationID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Integration ID in body and URL must be the same",
		})
		return
	}

	err := helpers.Store(r).UpdateIntegrationExtractor(project.ID, extractor)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetIntegrationExtractorRefs(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)

	log.Info(fmt.Sprintf("Extractor ID: %v", extractor.ID))

	refs, err := helpers.Store(r).GetIntegrationExtractorRefs(project.ID, extractor.ID, extractor.IntegrationID)
	log.Info(fmt.Sprintf("References found: %v", refs))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func DeleteIntegrationExtractor(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	extractor := context.Get(r, "extractor").(db.IntegrationExtractor)
	integration := context.Get(r, "integration").(db.Integration)

	log.Info(fmt.Sprintf("Delete requested for: %v", extractor.ID))

	err := helpers.Store(r).DeleteIntegrationExtractor(project.ID, extractor.ID, integration.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration Extractor failed to be deleted",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
