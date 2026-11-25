package projects

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
)

// workflowIntegrationsProOnly is the 403 body when an integration would start a workflow
// without the workflow_integrations_management feature; the web client shows it as is.
const workflowIntegrationsProOnly = "Integrations that run a workflow are only available for PRO users."

// IntegrationController serves the handlers that create or change an integration. They are the
// only ones that need the license: an integration that starts a workflow is a Pro feature.
type IntegrationController struct {
	pro proFeatureGate
}

func NewIntegrationController(subscription pro_interfaces.SubscriptionService) *IntegrationController {
	return &IntegrationController{pro: newProFeatureGate(subscription)}
}

// workflowTargetAllowed lets a template target through, and a workflow target only with the
// workflow_integrations_management feature. The one exception is an update that keeps the
// workflow the integration already has: an expired license must not freeze the name or the
// auth settings of an existing integration, it only stops new workflow targets. The answer
// is written on refusal (403 without the feature, 500 when the plan cannot be read); like the
// other plan refusals this records no audit deny event.
func (c *IntegrationController) workflowTargetAllowed(w http.ResponseWriter, r *http.Request, integration db.Integration, old *db.Integration) bool {
	if !integration.RunsWorkflow() {
		return true
	}
	if old != nil && old.RunsWorkflow() && *old.WorkflowID == *integration.WorkflowID {
		return true
	}

	features, err := c.pro.features(helpers.UserFromContext(r))
	if err != nil {
		log.WithFields(log.Fields{
			"context":    "integrations",
			"project_id": integration.ProjectID,
		}).WithError(err).Error("Failed to read the subscription plan")
		helpers.WriteErrorStatus(w, "Failed to read the subscription plan", http.StatusInternalServerError)
		return false
	}
	if !features.WorkflowIntegrationsManagement {
		helpers.WriteErrorStatus(w, workflowIntegrationsProOnly, http.StatusForbidden)
		return false
	}
	return true
}

// integrationMetadata names what the integration starts. An integration has
// exactly one target (Integration.Validate and the store enforce it), so one
// of the two ids is recorded, never both.
func integrationMetadata(integration db.Integration) audit.IntegrationMetadata {
	metadata := audit.IntegrationMetadata{AuthMethod: string(integration.AuthMethod)}
	if integration.RunsWorkflow() {
		metadata.WorkflowID = *integration.WorkflowID
	} else if integration.TemplateID != nil {
		metadata.TemplateID = *integration.TemplateID
	}
	return metadata
}

// authSecretFits checks the credential an integration points at against its
// auth method, and writes the 400 when it does not fit. Checked when the
// integration is stored: a webhook would otherwise fail at delivery time,
// far from the edit, and only in the server log.
func authSecretFits(w http.ResponseWriter, r *http.Request, integration db.Integration) bool {
	if integration.AuthSecretID == nil {
		return true
	}

	key, err := helpers.Store(r).GetAccessKey(integration.ProjectID, *integration.AuthSecretID)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Auth secret does not exist in this project",
		})
		return false
	}

	// Keys owned by a variable group, a secret storage or a task are
	// internal and never offered to integrations.
	if key.Owner != db.AccessKeyShared || !integration.AuthMethod.AcceptsKeyType(key.Type) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Auth secret type does not fit the auth method",
		})
		return false
	}

	return true
}

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

func (c *IntegrationController) AddIntegration(w http.ResponseWriter, r *http.Request) {
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

	if !c.workflowTargetAllowed(w, r, integration, nil) {
		return
	}

	if !authSecretFits(w, r, integration) {
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
		Metadata:  integrationMetadata(newIntegration),
	})

	helpers.WriteJSON(w, http.StatusCreated, newIntegration)
}

func (c *IntegrationController) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
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

	if !c.workflowTargetAllowed(w, r, integration, &oldIntegration) {
		return
	}

	if !authSecretFits(w, r, integration) {
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
		Metadata:  integrationMetadata(integration),
	})

	w.WriteHeader(http.StatusNoContent)
}

func DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	integration := helpers.GetFromContext(r, "integration").(db.Integration)
	project := helpers.GetFromContext(r, "project").(db.Project)

	err := helpers.Store(r).DeleteIntegration(project.ID, integration.ID)
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
			Target:    audit.ResourceTarget(audit.TargetIntegration, integration.ID, integration.Name),
			ProjectID: project.ID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
