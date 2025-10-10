package projects

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// ProjectMiddleware ensures a project exists and loads it to the context
func ProjectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetFromContext(r, "user").(*db.User)

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

		roleSlug := projectUser.Role

		permissions := roleSlug.GetPermissions()

		role, err := helpers.Store(r).GetProjectOrGlobalRoleBySlug(projectID, string(projectUser.Role))

		if err == nil {
			roleSlug = db.ProjectUserRole(role.Slug)
			permissions = role.Permissions
		} else if !errors.Is(err, db.ErrNotFound) {
			helpers.WriteError(w, err)
			return
		}

		if helpers.HasParam("template_id", r) {
			var templateID int
			templateID, err = helpers.GetIntParam("template_id", w, r)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
			var perm db.ProjectUserPermission
			perm, err = helpers.Store(r).GetTemplatePermission(project.ID, templateID, user.ID)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}

			permissions |= perm
		}

		r = helpers.SetContextValue(r, "projectUserRole", roleSlug)
		r = helpers.SetContextValue(r, "permissions", permissions)
		r = helpers.SetContextValue(r, "project", project)
		next.ServeHTTP(w, r)
	})
}

// GetMustCanMiddleware ensures that the user has administrator rights
func GetMustCanMiddleware(permissions db.ProjectUserPermission) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			me := helpers.GetFromContext(r, "user").(*db.User)

			userPerms := helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)

			can := (userPerms & permissions) == permissions

			if !me.Admin && r.Method != "GET" && r.Method != "HEAD" && !can {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type ProjectController struct {
	ProjectService server.ProjectService
}

// SendTestNotification triggers sending a test notification to enabled messengers for this project.
func (c *ProjectController) SendTestNotification(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Respect project.Alert flag: if disabled, still return 204 without sending
	if !project.Alert {
		w.WriteHeader(http.StatusConflict)
		return
	}

	err := tasks.SendProjectTestAlerts(project, helpers.Store(r))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProjectController) UpdateProject(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
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

	err := c.ProjectService.UpdateProject(body)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteProject removes a project from the database
func (c *ProjectController) DeleteProject(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	err := c.ProjectService.DeleteProject(project.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = util.Config.ClearProjectTmpDir(project.ID)
	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProject returns a project details
func GetProject(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, helpers.GetFromContext(r, "project"))
}

func GetUserRole(w http.ResponseWriter, r *http.Request) {
	var permissions struct {
		Role        db.ProjectUserRole       `json:"role"`
		Permissions db.ProjectUserPermission `json:"permissions"`
	}
	permissions.Role = helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)
	permissions.Permissions = permissions.Role.GetPermissions()
	helpers.WriteJSON(w, http.StatusOK, permissions)
}

func ClearCache(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	err := util.Config.ClearProjectTmpDir(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
