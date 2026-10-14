package projects

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// ProjectMiddleware ensures a project exists and loads it to the context
func ProjectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetFromContext(r, "user").(*db.User)

		projectID, ok := helpers.GetIntParamOrAbort("project_id", w, r)

		if !ok {
			return
		}

		// check if user in project's team
		roles, err := helpers.Store(r).GetProjectUserRoles(projectID, user.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		if len(roles) == 0 && !user.Admin {
			helpers.WriteError(w, db.ErrNotFound)
			return
		}

		project, err := helpers.Store(r).GetProject(projectID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		var permissions db.ProjectUserPermission
		for _, role := range roles {
			permissions |= role.Permissions
		}

		if helpers.HasParam("template_id", r) {
			templateID, templateOk := helpers.GetIntParamOrAbort("template_id", w, r)
			if !templateOk {
				return
			}
			grants, err := helpers.Store(r).GetTemplatePermission(project.ID, templateID, user.ID)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
			permissions |= grants
		}

		r = helpers.SetContextValue(r, "projectRoles", roles)
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
				projectID := 0
				if project, ok := helpers.GetOkFromContext(r, "project"); ok {
					projectID = project.(db.Project).ID
				}
				// Every call site passes a single permission bit.
				helpers.RecordDenied(r, strings.Join(audit.PermissionNames(permissions), ","), projectID)
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

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceProjectUpdate,
		Target:    audit.ResourceTarget(audit.TargetProject, project.ID, body.Name),
		ProjectID: project.ID,
	})

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

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceProjectDelete,
		Target:    audit.ResourceTarget(audit.TargetProject, project.ID, project.Name),
		ProjectID: project.ID,
	})

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
	var result struct {
		Role        string                   `json:"role"`
		Permissions db.ProjectUserPermission `json:"permissions"`
	}
	if roles := helpers.GetFromContext(r, "projectRoles").([]db.Role); len(roles) > 0 {
		result.Role = roles[0].Slug
	}
	result.Permissions = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
	helpers.WriteJSON(w, http.StatusOK, result)
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
