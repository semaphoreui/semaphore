package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/neo1908/semaphore/api/helpers"
	"github.com/neo1908/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
	"strconv"
)

// TemplatesMiddleware ensures a template exists and loads it to the context
func TemplatesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		templateID, err := helpers.GetIntParam("template_id", w, r)
		if err != nil {
			return
		}

		template, err := helpers.Store(r).GetTemplate(project.ID, templateID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "template", template)
		next.ServeHTTP(w, r)
	})
}

// GetTemplate returns single template by ID
func GetTemplate(w http.ResponseWriter, r *http.Request) {
	template := context.Get(r, "template").(db.Template)
	helpers.WriteJSON(w, http.StatusOK, template)
}

// GetTemplates returns all templates for a project in a sort order
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	templates, err := helpers.Store(r).GetTemplates(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, templates)
}

// AddTemplate adds a template to the database
func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var template db.Template
	if !helpers.Bind(w, r, &template) {
		return
	}

	template.ProjectID = project.ID
	template, err := helpers.Store(r).CreateTemplate(template)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	objType := db.EventTemplate
	desc := "Template ID " + strconv.Itoa(template.ID) + " created"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &template.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, template)
}

// UpdateTemplate writes a template to an existing key in the database
func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := context.Get(r, "template").(db.Template)

	var template db.Template
	if !helpers.Bind(w, r, &template) {
		return
	}

	// project ID and template ID in the body and the path must be the same

	if template.ID != oldTemplate.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "template id in URL and in body must be the same",
		})
		return
	}

	if template.ProjectID != oldTemplate.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "You can not move template to other project",
		})
		return
	}

	if template.Arguments != nil && *template.Arguments == "" {
		template.Arguments = nil
	}

	err := helpers.Store(r).UpdateTemplate(template)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "Template ID " + strconv.Itoa(template.ID) + " updated"
	objType := db.EventTemplate

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &template.ProjectID,
		Description: &desc,
		ObjectID:    &template.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTemplate deletes a template from the database
func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(db.Template)

	err := helpers.Store(r).DeleteTemplate(tpl.ProjectID, tpl.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	desc := "Template ID " + strconv.Itoa(tpl.ID) + " deleted"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
