package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
)

// publicAlias is an endpoint alias as the API shows it: the id and the full
// public URL webhooks are sent to. The URL path stays /api/integrations/{alias}
// so URLs configured before the redesign keep working.
type publicAlias struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

func getPublicAlias(alias db.EndpointAlias) publicAlias {
	return publicAlias{
		ID:  alias.ID,
		URL: util.GetPublicAliasURL("integrations", alias.Alias),
	}
}

func getPublicAliases(aliases []db.EndpointAlias) (res []publicAlias) {
	res = make([]publicAlias, 0)
	for _, alias := range aliases {
		res = append(res, getPublicAlias(alias))
	}

	return
}

func GetEndpointAliases(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)

	aliases, err := helpers.Store(r).GetEndpointAliases(project.ID, endpoint.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, getPublicAliases(aliases))
}

func AddEndpointAlias(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)

	alias, err := helpers.Store(r).CreateEndpointAlias(db.EndpointAlias{
		Alias:      random.String(16),
		ProjectID:  project.ID,
		EndpointID: endpoint.ID,
	})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The alias value works as a bearer secret, so only its record ID is recorded.
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEndpointAliasCreate,
		Target:    audit.ResourceTarget(audit.TargetEndpointAlias, alias.ID, ""),
		ProjectID: project.ID,
		Metadata:  audit.EndpointPartMetadata{EndpointID: endpoint.ID},
	})

	helpers.WriteJSON(w, http.StatusOK, getPublicAlias(alias))
}

func RemoveEndpointAlias(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	endpoint := helpers.GetFromContext(r, "endpoint").(db.Endpoint)
	aliasID, ok := helpers.GetIntParamOrAbort("alias_id", w, r)
	if !ok {
		return
	}

	// The store deletes by endpoint, so an alias of another endpoint is not found.
	err := helpers.Store(r).DeleteEndpointAlias(project.ID, endpoint.ID, aliasID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEndpointAliasDelete,
		Target:    audit.ResourceTarget(audit.TargetEndpointAlias, aliasID, ""),
		ProjectID: project.ID,
		Metadata:  audit.EndpointPartMetadata{EndpointID: endpoint.ID},
	})

	w.WriteHeader(http.StatusNoContent)
}
