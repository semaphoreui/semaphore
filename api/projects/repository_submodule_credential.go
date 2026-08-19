package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// SubmoduleCredentialMiddleware ensures a submodule credential exists on the
// repository in context and loads it into the request context.
func SubmoduleCredentialMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repository := helpers.GetFromContext(r, "repository").(db.Repository)

		credentialID, err := helpers.GetIntParam("submodule_credential_id", w, r)
		if err != nil {
			return
		}

		credential, err := helpers.Store(r).GetRepositorySubmoduleCredential(repository.ProjectID, repository.ID, credentialID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "submoduleCredential", credential)
		next.ServeHTTP(w, r)
	})
}

// GetRepositorySubmoduleCredentials returns the submodule host/access-key
// mappings configured for the repository in context.
func GetRepositorySubmoduleCredentials(w http.ResponseWriter, r *http.Request) {
	repository := helpers.GetFromContext(r, "repository").(db.Repository)

	credentials, err := helpers.Store(r).GetRepositorySubmoduleCredentials(repository.ProjectID, repository.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, credentials)
}

// AddRepositorySubmoduleCredential creates a new submodule credential mapping
// for the repository in context.
func AddRepositorySubmoduleCredential(w http.ResponseWriter, r *http.Request) {
	repository := helpers.GetFromContext(r, "repository").(db.Repository)

	var credential db.RepositorySubmoduleCredential

	if !helpers.Bind(w, r, &credential) {
		return
	}

	credential.ProjectID = repository.ProjectID
	credential.RepositoryID = repository.ID

	newCredential, err := helpers.Store(r).CreateRepositorySubmoduleCredential(credential)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   repository.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    repository.ID,
		Description: fmt.Sprintf("Submodule credential for host %s added to repository %s", credential.Host, repository.GitURL),
	})

	helpers.WriteJSON(w, http.StatusCreated, newCredential)
}

// UpdateRepositorySubmoduleCredential updates an existing submodule credential
// mapping.
func UpdateRepositorySubmoduleCredential(w http.ResponseWriter, r *http.Request) {
	repository := helpers.GetFromContext(r, "repository").(db.Repository)
	oldCredential := helpers.GetFromContext(r, "submoduleCredential").(db.RepositorySubmoduleCredential)

	var credential db.RepositorySubmoduleCredential

	if !helpers.Bind(w, r, &credential) {
		return
	}

	if credential.ID != 0 && credential.ID != oldCredential.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Submodule credential ID in body and URL must be the same",
		})
		return
	}

	credential.ID = oldCredential.ID
	credential.ProjectID = repository.ProjectID
	credential.RepositoryID = repository.ID

	if err := helpers.Store(r).UpdateRepositorySubmoduleCredential(credential); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   repository.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    repository.ID,
		Description: fmt.Sprintf("Submodule credential for host %s updated on repository %s", credential.Host, repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRepositorySubmoduleCredential deletes a submodule credential mapping.
func RemoveRepositorySubmoduleCredential(w http.ResponseWriter, r *http.Request) {
	repository := helpers.GetFromContext(r, "repository").(db.Repository)
	credential := helpers.GetFromContext(r, "submoduleCredential").(db.RepositorySubmoduleCredential)

	if err := helpers.Store(r).DeleteRepositorySubmoduleCredential(repository.ProjectID, repository.ID, credential.ID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   repository.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    repository.ID,
		Description: fmt.Sprintf("Submodule credential for host %s removed from repository %s", credential.Host, repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}
