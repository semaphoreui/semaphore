package projects

import (
	"github.com/neo1908/semaphore/api/helpers"
	"github.com/neo1908/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

// ProjectMiddleware ensures a project exists and loads it to the context
func ProjectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.Get(r, "user").(*db.User)

		projectID, err := helpers.GetIntParam("project_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid project ID",
			})
			return
		}

		// check if user it project's team
		_, err = helpers.Store(r).GetProjectUser(projectID, user.ID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		project, err := helpers.Store(r).GetProject(projectID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "project", project)
		next.ServeHTTP(w, r)
	})
}

// MustBeAdmin ensures that the user has administrator rights
func MustBeAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		user := context.Get(r, "user").(*db.User)

		projectUser, err := helpers.Store(r).GetProjectUser(project.ID, user.ID)

		if err == db.ErrNotFound {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		if !projectUser.Admin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

//GetProject returns a project details
func GetProject(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, context.Get(r, "project"))
}

// UpdateProject saves updated project details to the database
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var body db.Project

	if !helpers.Bind(w, r, &body) {
		return
	}

	if body.ID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	err := helpers.Store(r).UpdateProject(body)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteProject removes a project from the database
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	err := helpers.Store(r).DeleteProject(project.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
