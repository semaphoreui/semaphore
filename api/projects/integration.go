package projects

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
)

func IntegrationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		integrationId, err := helpers.GetIntParam("integration_id", w, r)
		projectId, err := helpers.GetIntParam("project_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid integration ID",
			})
			return
		}

		integration, err := helpers.Store(r).GetIntegration(projectId, integrationId)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "integration", integration)
		next.ServeHTTP(w, r)
	})
}

func GetIntegration(w http.ResponseWriter, r *http.Request) {
	integration := context.Get(r, "integration").(db.Integration)
	helpers.WriteJSON(w, http.StatusOK, integration)
}

func GetIntegrations(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	integrations, err := helpers.Store(r).GetIntegrations(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, integrations)
}

func GetIntegrationRefs(w http.ResponseWriter, r *http.Request) {
	integration_id, err := helpers.GetIntParam("integration_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Integration ID",
		})
		return
	}

	project := context.Get(r, "project").(db.Project)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	refs, err := helpers.Store(r).GetIntegrationRefs(project.ID, integration_id)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func AddIntegration(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var integration db.Integration
	log.Info(fmt.Sprintf("Found Project: %v", project.ID))

	if !helpers.Bind(w, r, &integration) {
		log.Info("Failed to bind for integration uploads")

		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if integration.ProjectID != project.ID {
		log.Error(fmt.Sprintf("Project ID in body and URL must be the same: %v vs. %v", integration.ProjectID, project.ID))

		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}
	err := integration.Validate()
	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newIntegration, errIntegration := helpers.Store(r).CreateIntegration(integration)

	if errIntegration != nil {
		log.Error(errIntegration)
		helpers.WriteError(w, errIntegration)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newIntegration)
}

func UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	oldIntegration := context.Get(r, "integration").(db.Integration)
	var integration db.Integration

	if !helpers.Bind(w, r, &integration) {
		return
	}

	if integration.ID != oldIntegration.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Integration ID in body and URL must be the same",
		})
		return
	}

	if integration.ProjectID != oldIntegration.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	err := helpers.Store(r).UpdateIntegration(integration)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	integration_id, err := helpers.GetIntParam("integration_id", w, r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	project := context.Get(r, "project").(db.Project)

	err = helpers.Store(r).DeleteIntegration(project.ID, integration_id)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Integration failed to be deleted",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
