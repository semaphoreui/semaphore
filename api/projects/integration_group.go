package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
)

func integrationGroupMetadata(group db.IntegrationGroup) audit.IntegrationGroupMetadata {
	metadata := audit.IntegrationGroupMetadata{EndpointID: group.EndpointID, AuthOverride: group.AuthOverride}
	if group.AuthOverride {
		metadata.AuthMethod = string(group.AuthMethod)
	}
	return metadata
}

func IntegrationGroupMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, ok := helpers.GetIntParamOrAbort("group_id", w, r)
		if !ok {
			return
		}

		projectID, ok := helpers.GetIntParamOrAbort("project_id", w, r)
		if !ok {
			return
		}

		group, err := helpers.Store(r).GetIntegrationGroup(projectID, groupID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "integration_group", group)
		next.ServeHTTP(w, r)
	})
}

func GetIntegrationGroup(w http.ResponseWriter, r *http.Request) {
	group := helpers.GetFromContext(r, "integration_group").(db.IntegrationGroup)
	helpers.WriteJSON(w, http.StatusOK, group)
}

func GetIntegrationGroups(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	groups, err := helpers.Store(r).GetIntegrationGroups(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, groups)
}

func GetIntegrationGroupRefs(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	group := helpers.GetFromContext(r, "integration_group").(db.IntegrationGroup)

	refs, err := helpers.Store(r).GetIntegrationGroupRefs(project.ID, group.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetIntegrationGroupIntegrations lists the integrations of one group, the
// way the integrations page shows them.
func GetIntegrationGroupIntegrations(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	group := helpers.GetFromContext(r, "integration_group").(db.IntegrationGroup)

	integrations, err := helpers.Store(r).GetIntegrationsByGroup(project.ID, group.ID, false)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, integrations)
}

// groupAuthFits checks the group's own key only when the group overrides the
// endpoint's auth; otherwise the store clears the auth anyway.
func groupAuthFits(w http.ResponseWriter, r *http.Request, projectID int, group db.IntegrationGroup) bool {
	if !group.AuthOverride {
		return true
	}
	return authSecretFits(w, r, projectID, group.IntegrationAuth)
}

func AddIntegrationGroup(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var group db.IntegrationGroup
	if !helpers.Bind(w, r, &group) {
		return
	}

	if group.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := group.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if !groupAuthFits(w, r, project.ID, group) {
		return
	}

	newGroup, err := helpers.Store(r).CreateIntegrationGroup(group)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationGroupCreate,
		Target:    audit.ResourceTarget(audit.TargetIntegrationGroup, newGroup.ID, newGroup.Name),
		ProjectID: project.ID,
		Metadata:  integrationGroupMetadata(newGroup),
	})

	helpers.WriteJSON(w, http.StatusCreated, newGroup)
}

func UpdateIntegrationGroup(w http.ResponseWriter, r *http.Request) {
	oldGroup := helpers.GetFromContext(r, "integration_group").(db.IntegrationGroup)

	var group db.IntegrationGroup
	if !helpers.Bind(w, r, &group) {
		return
	}

	if group.ID != oldGroup.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Integration group ID in body and URL must be the same",
		})
		return
	}

	if group.ProjectID != oldGroup.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if !groupAuthFits(w, r, oldGroup.ProjectID, group) {
		return
	}

	if err := helpers.Store(r).UpdateIntegrationGroup(group); err != nil {
		helpers.WriteError(w, err)
		return
	}

	group.NormalizeAuth()
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationGroupUpdate,
		Target:    audit.ResourceTarget(audit.TargetIntegrationGroup, oldGroup.ID, group.Name),
		ProjectID: oldGroup.ProjectID,
		Metadata:  integrationGroupMetadata(group),
	})

	w.WriteHeader(http.StatusNoContent)
}

// DeleteIntegrationGroup deletes the group with its integrations.
func DeleteIntegrationGroup(w http.ResponseWriter, r *http.Request) {
	group := helpers.GetFromContext(r, "integration_group").(db.IntegrationGroup)
	project := helpers.GetFromContext(r, "project").(db.Project)

	err := helpers.Store(r).DeleteIntegrationGroup(project.ID, group.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationGroupDelete,
		Target:    audit.ResourceTarget(audit.TargetIntegrationGroup, group.ID, group.Name),
		ProjectID: project.ID,
	})

	w.WriteHeader(http.StatusNoContent)
}
