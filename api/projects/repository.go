package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

// RepositoryMiddleware ensures a repository exists and loads it to the context
func RepositoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		repositoryID, err := helpers.GetIntParam("repository_id", w, r)
		if err != nil {
			return
		}

		repository, err := helpers.Store(r).GetRepository(project.ID, repositoryID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "repository", repository)
		next.ServeHTTP(w, r)
	})
}

func GetRepositoryRefs(w http.ResponseWriter, r *http.Request) {
	repo := context.Get(r, "repository").(db.Repository)
	refs, err := helpers.Store(r).GetRepositoryRefs(repo.ProjectID, repo.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetRepositories returns all repositories in a project sorted by type
func GetRepositories(w http.ResponseWriter, r *http.Request) {
	if repo := context.Get(r, "repository"); repo != nil {
		helpers.WriteJSON(w, http.StatusOK, repo.(db.Repository))
		return
	}

	project := context.Get(r, "project").(db.Project)

	repos, err := helpers.Store(r).GetRepositories(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, repos)
}

// AddRepository creates a new repository in the database
func AddRepository(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var repository db.Repository

	if !helpers.Bind(w, r, &repository) {
		return
	}

	if repository.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
	}

	if err := db.ValidateRepository(helpers.Store(r), &repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	newRepo, err := helpers.Store(r).CreateRepository(repository)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newRepo.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    newRepo.ID,
		Description: fmt.Sprintf("Repository %s created", repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}

// UpdateRepository updates the values of a repository in the database
func UpdateRepository(w http.ResponseWriter, r *http.Request) {
	oldRepo := context.Get(r, "repository").(db.Repository)
	var repository db.Repository

	if !helpers.Bind(w, r, &repository) {
		return
	}

	if repository.ID != oldRepo.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Repository ID in body and URL must be the same",
		})
		return
	}

	if repository.ProjectID != oldRepo.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := db.ValidateRepository(helpers.Store(r), &repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateRepository(repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if oldRepo.GitURL != repository.GitURL {
		util.LogWarning(oldRepo.ClearCache())
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldRepo.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    oldRepo.ID,
		Description: fmt.Sprintf("Repository %s updated", repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRepository deletes a repository from a project in the database
func RemoveRepository(w http.ResponseWriter, r *http.Request) {
	repository := context.Get(r, "repository").(db.Repository)

	var err error

	err = helpers.Store(r).DeleteRepository(repository.ProjectID, repository.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Repository is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	util.LogWarning(repository.ClearCache())

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   repository.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    repository.ID,
		Description: fmt.Sprintf("Repository %s deleted", repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}
