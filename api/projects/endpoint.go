package projects

import (
	"errors"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/audit"
)

// authSecretFits checks the credential an endpoint or an integration group
// verifies webhooks with against its auth method, and writes the 400 when it
// does not fit. Checked when the object is stored: a webhook would otherwise
// fail at delivery time, far from the edit, and only in the server log.
func authSecretFits(w http.ResponseWriter, r *http.Request, projectID int, auth db.IntegrationAuth) bool {
	if auth.AuthSecretID == nil {
		return true
	}

	key, err := helpers.Store(r).GetAccessKey(projectID, *auth.AuthSecretID)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Auth secret does not exist in this project",
		})
		return false
	}

	// Keys owned by a variable group, a secret storage or a task are
	// internal and never offered to webhooks.
	if key.Owner != db.AccessKeyShared || !auth.AuthMethod.AcceptsKeyType(key.Type) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Auth secret type does not fit the auth method",
		})
		return false
	}

	return true
}

func endpointMetadata(endpoint db.Endpoint) audit.EndpointMetadata {
	return audit.EndpointMetadata{AuthMethod: string(endpoint.AuthMethod)}
}

func EndpointMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpointID, ok := helpers.GetIntParamOrAbort("endpoint_id", w, r)
		if !ok {
			return
		}

		projectID, ok := helpers.GetIntParamOrAbort("project_id", w, r)
		if !ok {
			return
		}

		endpoint, err := helpers.Store(r).GetEndpoint(projectID, endpointID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "endpoint", endpoint)
		next.ServeHTTP(w, r)
	})
}

func GetEndpoint(w http.ResponseWriter, r *http.Request) {
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)
	helpers.WriteJSON(w, http.StatusOK, endpoint)
}

func GetEndpoints(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	endpoints, err := helpers.Store(r).GetEndpoints(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, endpoints)
}

func GetEndpointRefs(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)

	refs, err := helpers.Store(r).GetEndpointRefs(project.ID, endpoint.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

func AddEndpoint(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var endpoint db.Endpoint
	if !helpers.Bind(w, r, &endpoint) {
		return
	}

	if endpoint.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := endpoint.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if !authSecretFits(w, r, project.ID, endpoint.IntegrationAuth) {
		return
	}

	newEndpoint, err := helpers.Store(r).CreateEndpoint(endpoint)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// auth_method=none lets anyone with a URL start the endpoint's integrations, which is what SIEM rules look for.
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEndpointCreate,
		Target:    audit.ResourceTarget(audit.TargetEndpoint, newEndpoint.ID, newEndpoint.Name),
		ProjectID: project.ID,
		Metadata:  endpointMetadata(newEndpoint),
	})

	helpers.WriteJSON(w, http.StatusCreated, newEndpoint)
}

func UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	oldEndpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)

	var endpoint db.Endpoint
	if !helpers.Bind(w, r, &endpoint) {
		return
	}

	if endpoint.ID != oldEndpoint.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Endpoint ID in body and URL must be the same",
		})
		return
	}

	if endpoint.ProjectID != oldEndpoint.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if !authSecretFits(w, r, oldEndpoint.ProjectID, endpoint.IntegrationAuth) {
		return
	}

	if err := helpers.Store(r).UpdateEndpoint(endpoint); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEndpointUpdate,
		Target:    audit.ResourceTarget(audit.TargetEndpoint, oldEndpoint.ID, endpoint.Name),
		ProjectID: oldEndpoint.ProjectID,
		Metadata:  endpointMetadata(endpoint),
	})

	w.WriteHeader(http.StatusNoContent)
}

// DeleteEndpoint refuses with 400 while integration groups use the endpoint;
// the groups are listed by GET .../refs.
func DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)
	project := helpers.GetFromContext(r, "project").(db.Project)

	err := helpers.Store(r).DeleteEndpoint(project.ID, endpoint.ID)
	var validationErr *common_errors.ValidationError
	if errors.As(err, &validationErr) || errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Endpoint is used by integration groups",
		})
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEndpointDelete,
		Target:    audit.ResourceTarget(audit.TargetEndpoint, endpoint.ID, endpoint.Name),
		ProjectID: project.ID,
	})

	w.WriteHeader(http.StatusNoContent)
}
