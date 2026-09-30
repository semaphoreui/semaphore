package projects

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
)

func IntegrationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		integrationId, ok := helpers.GetIntParamOrAbort("integration_id", w, r)
		if !ok {
			return
		}

		projectId, ok := helpers.GetIntParamOrAbort("project_id", w, r)
		if !ok {
			return
		}

		integration, err := helpers.Store(r).GetIntegration(projectId, integrationId)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "integration", integration)
		next.ServeHTTP(w, r)
	})
}

func GetIntegration(w http.ResponseWriter, r *http.Request) {
	integration := helpers.GetFromContext(r, "integration").(db.Integration)
	helpers.WriteJSON(w, http.StatusOK, integration)
}

func GetIntegrations(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integrations, err := helpers.Store(r).GetIntegrations(project.ID, helpers.QueryParams(r.URL), false)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, integrations)
}

func GetIntegrationRefs(w http.ResponseWriter, r *http.Request) {
	integration_id, ok := helpers.GetIntParamOrAbort("integration_id", w, r)

	if !ok {
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	refs, err := helpers.Store(r).GetIntegrationRefs(project.ID, integration_id)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func AddIntegration(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
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

	// auth_method=none lets anyone start the template, which is what SIEM rules look for.
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationCreate,
		Target:    audit.ResourceTarget(audit.TargetIntegration, newIntegration.ID, newIntegration.Name),
		ProjectID: project.ID,
		Metadata:  audit.IntegrationMetadata{TemplateID: newIntegration.TemplateID, AuthMethod: string(newIntegration.AuthMethod)},
	})

	helpers.WriteJSON(w, http.StatusCreated, newIntegration)
}

func UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	oldIntegration := helpers.GetFromContext(r, "integration").(db.Integration)
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

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationUpdate,
		Target:    audit.ResourceTarget(audit.TargetIntegration, oldIntegration.ID, integration.Name),
		ProjectID: oldIntegration.ProjectID,
		Metadata:  audit.IntegrationMetadata{TemplateID: integration.TemplateID, AuthMethod: string(integration.AuthMethod)},
	})

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	integration_id, ok := helpers.GetIntParamOrAbort("integration_id", w, r)
	if !ok {
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	err := helpers.Store(r).DeleteIntegration(project.ID, integration_id)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Integration failed to be deleted",
		})
		return
	}

	// The handler answers 204 on other store errors too, so only a real delete is recorded.
	if err == nil {
		helpers.Audit(r).Record(r.Context(), audit.Event{
			Kind:      audit.ResourceIntegrationDelete,
			Target:    audit.ResourceTarget(audit.TargetIntegration, integration_id, ""),
			ProjectID: project.ID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
