package projects

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
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

func GetTemplateRefs(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(db.Template)
	refs, err := helpers.Store(r).GetTemplateRefs(tpl.ProjectID, tpl.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetTemplates returns all templates for a project in a sort order
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))

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

	var err error

	template.ProjectID = project.ID
	newTemplate, err := helpers.Store(r).CreateTemplate(template)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if _, ok := util.Config.Apps[string(newTemplate.App)]; !ok {
		helpers.WriteErrorStatus(w, "Invalid app id: "+string(newTemplate.App), http.StatusBadRequest)
		return
	}

	// Check workspace and create it if required.
	if newTemplate.App.IsTerraform() {
		var inv db.Inventory

		if newTemplate.InventoryID == nil {
			inv, err = helpers.Store(r).CreateInventory(db.Inventory{
				Name:      newTemplate.Name + " - default",
				ProjectID: project.ID,
				HolderID:  &newTemplate.ID,
				Type:      db.InventoryTerraformWorkspace,
				Inventory: "default",
			})

			if err != nil {
				helpers.WriteError(w, err)
				return
			}

			newTemplate.InventoryID = &inv.ID
			err = helpers.Store(r).UpdateTemplate(newTemplate)

		} else {
			inv, err = helpers.Store(r).GetInventory(project.ID, *newTemplate.InventoryID)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}

			inv.HolderID = &newTemplate.ID
			err = helpers.Store(r).UpdateInventory(inv)
		}

		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventSchedule,
		ObjectID:    newTemplate.ID,
		Description: fmt.Sprintf("Template ID %d created", newTemplate.ID),
	})

	helpers.WriteJSON(w, http.StatusCreated, newTemplate)
}

// UpdateTemplate writes a template to an existing key in the database
func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := context.Get(r, "template").(db.Template)

	var template db.Template
	if !helpers.Bind(w, r, &template) {
		return
	}

	if _, ok := util.Config.Apps[string(template.App)]; !ok {
		helpers.WriteErrorStatus(w, "Invalid app id: "+string(template.App), http.StatusBadRequest)
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

	if template.Type != db.TemplateDeploy {
		template.BuildTemplateID = nil
	}

	if template.Type != db.TemplateBuild {
		template.StartVersion = nil
	}

	err := helpers.Store(r).UpdateTemplate(template)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldTemplate.ProjectID,
		ObjectType:  db.EventTemplate,
		ObjectID:    oldTemplate.ID,
		Description: fmt.Sprintf("Template ID %d updated", template.ID),
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

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   tpl.ProjectID,
		ObjectType:  db.EventTemplate,
		ObjectID:    tpl.ID,
		Description: fmt.Sprintf("Template ID %d deleted", tpl.ID),
	})

	w.WriteHeader(http.StatusNoContent)
}
