package projects

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/mux"
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

		// check if user in project's team
		projectUser, err := helpers.Store(r).GetProjectUser(projectID, user.ID)

		if !user.Admin && err != nil {
			helpers.WriteError(w, err)
			return
		}

		project, err := helpers.Store(r).GetProject(projectID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "projectUserRole", projectUser.Role)
		context.Set(r, "project", project)
		next.ServeHTTP(w, r)
	})
}

// GetMustCanMiddleware ensures that the user has administrator rights
func GetMustCanMiddleware(permissions db.ProjectUserPermission) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			me := context.Get(r, "user").(*db.User)
			myRole := context.Get(r, "projectUserRole").(db.ProjectUserRole)

			if !me.Admin && r.Method != "GET" && r.Method != "HEAD" && !myRole.Can(permissions) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetProject returns a project details
func GetProject(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, context.Get(r, "project"))
}

func GetUserRole(w http.ResponseWriter, r *http.Request) {
	var permissions struct {
		Role        db.ProjectUserRole       `json:"role"`
		Permissions db.ProjectUserPermission `json:"permissions"`
	}
	permissions.Role = context.Get(r, "projectUserRole").(db.ProjectUserRole)
	permissions.Permissions = permissions.Role.GetPermissions()
	helpers.WriteJSON(w, http.StatusOK, permissions)
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
