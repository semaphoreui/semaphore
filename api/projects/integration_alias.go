package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
)

type publicAlias struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

func getPublicAlias(alias db.IntegrationAlias) publicAlias {

	return publicAlias{
		ID:  alias.ID,
		URL: util.GetPublicAliasURL("integrations", alias.Alias),
	}
}

func getPublicAliases(aliases []db.IntegrationAlias) (res []publicAlias) {

	res = make([]publicAlias, 0)
	for _, alias := range aliases {
		res = append(res, getPublicAlias(alias))
	}

	return
}

func GetIntegrationAlias(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration, ok := helpers.GetFromContext(r, "integration").(db.Integration)

	var integrationId *int
	if ok {
		integrationId = &integration.ID
	}

	aliases, err := helpers.Store(r).GetIntegrationAliases(project.ID, integrationId)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, getPublicAliases(aliases))
}

func AddIntegrationAlias(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	integration, ok := helpers.GetFromContext(r, "integration").(db.Integration)

	var integrationId *int
	if ok {
		integrationId = &integration.ID
	}

	alias, err := helpers.Store(r).CreateIntegrationAlias(db.IntegrationAlias{
		Alias:         random.String(16),
		ProjectID:     project.ID,
		IntegrationID: integrationId,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The alias value works as a bearer secret, so only its record ID is recorded.
	var part audit.IntegrationPartMetadata
	if integrationId != nil {
		part.IntegrationID = *integrationId
	}
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationAliasCreate,
		Target:    audit.ResourceTarget(audit.TargetIntegrationAlias, alias.ID, ""),
		ProjectID: project.ID,
		Metadata:  part,
	})

	helpers.WriteJSON(w, http.StatusOK, getPublicAlias(alias))
}

func RemoveIntegrationAlias(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	aliasID, ok := helpers.GetIntParamOrAbort("alias_id", w, r)

	if !ok {
		return
	}

	err := helpers.Store(r).DeleteIntegrationAlias(project.ID, aliasID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var part audit.IntegrationPartMetadata
	if integration, ok := helpers.GetFromContext(r, "integration").(db.Integration); ok {
		part.IntegrationID = integration.ID
	}
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceIntegrationAliasDelete,
		Target:    audit.ResourceTarget(audit.TargetIntegrationAlias, aliasID, ""),
		ProjectID: project.ID,
		Metadata:  part,
	})

	w.WriteHeader(http.StatusNoContent)
}
