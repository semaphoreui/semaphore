package projects

import (
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/random"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

type publicAlias struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

func getPublicAlias(alias db.IntegrationAlias) publicAlias {

	aliasURL := util.Config.WebHost

	if !strings.HasSuffix(aliasURL, "/") {
		aliasURL += "/"
	}

	aliasURL += "api/integrations/" + alias.Alias

	return publicAlias{
		ID:  alias.ID,
		URL: aliasURL,
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
	project := context.Get(r, "project").(db.Project)
	integration, ok := context.Get(r, "integration").(db.Integration)

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
	project := context.Get(r, "project").(db.Project)
	integration, ok := context.Get(r, "integration").(db.Integration)

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

	helpers.WriteJSON(w, http.StatusOK, getPublicAlias(alias))
}

func RemoveIntegrationAlias(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	aliasID, err := helpers.GetIntParam("alias_id", w, r)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteIntegrationAlias(project.ID, aliasID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
