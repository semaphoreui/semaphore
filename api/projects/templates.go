package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// TemplatesMiddleware ensures a template exists and loads it to the context
func TemplatesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		templateID, err := helpers.GetIntParam("template_id", w, r)
		if err != nil {
			return
		}

		template, err := helpers.Store(r).GetTemplate(project.ID, templateID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "template", template)
		next.ServeHTTP(w, r)
	})
}

type TemplateController struct {
	templateRepo db.TemplateManager
	roleRepo     db.RoleRepository
}

func NewTemplateController(
	templateRepo db.TemplateManager,
	roleRepo db.RoleRepository,
) *TemplateController {
	return &TemplateController{
		templateRepo: templateRepo,
		roleRepo:     roleRepo,
	}
}

// GetTemplate returns single template by ID
func GetTemplate(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)
	permissions := helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
	res := db.TemplateWithPerms{
		Template:    template,
		Permissions: &permissions,
	}
	helpers.WriteJSON(w, http.StatusOK, res)
}

func GetTemplateRefs(w http.ResponseWriter, r *http.Request) {
	tpl := helpers.GetFromContext(r, "template").(db.Template)
	refs, err := helpers.Store(r).GetTemplateRefs(tpl.ProjectID, tpl.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetTemplates returns all templates for a project in a sort order
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	filter := db.TemplateFilter{}
	if r.URL.Query().Get("app") != "" {
		app := db.TemplateApp(r.URL.Query().Get("app"))
		filter.App = &app
	}
	templates, err := helpers.Store(r).GetTemplates(project.ID, filter, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, templates)
}

// AddTemplate adds a template to the database
func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

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
			var inventoryType db.InventoryType

			if invTypes := newTemplate.App.InventoryTypes(); len(invTypes) > 0 {
				inventoryType = invTypes[0]
			} else {
				helpers.WriteErrorStatus(w, "Inventory type is not supported for this template", http.StatusBadRequest)
				return
			}

			inv, err = helpers.Store(r).CreateInventory(db.Inventory{
				Name:       "default",
				ProjectID:  project.ID,
				TemplateID: &newTemplate.ID,
				Type:       inventoryType,
				Inventory:  "default",
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

			inv.TemplateID = &newTemplate.ID
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

func UpdateTemplateDescription(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)

	var tpl struct {
		Description string `json:"description"`
	}

	if !helpers.Bind(w, r, &tpl) {
		return
	}

	err := helpers.Store(r).SetTemplateDescription(template.ProjectID, template.ID, tpl.Description)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   template.ProjectID,
		ObjectType:  db.EventTemplate,
		ObjectID:    template.ID,
		Description: fmt.Sprintf("Template ID %d description updated", template.ID),
	})

	w.WriteHeader(http.StatusNoContent)
}

// UpdateTemplate writes a template to an existing key in the database
func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := helpers.GetFromContext(r, "template").(db.Template)

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

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTemplate deletes a template from the database
func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := helpers.GetFromContext(r, "template").(db.Template)

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

func SetTemplateInventory(w http.ResponseWriter, r *http.Request) {
	tpl := helpers.GetFromContext(r, "template").(db.Template)
	inv := helpers.GetFromContext(r, "inventory").(db.Inventory)

	if !tpl.App.HasInventoryType(inv.Type) {
		helpers.WriteErrorStatus(w, "Inventory type is not supported for this template", http.StatusBadRequest)
		return
	}

	if tpl.App.IsTerraform() && (inv.TemplateID == nil || *inv.TemplateID != tpl.ID) {
		helpers.WriteErrorStatus(w, "Inventory is not attached to this template", http.StatusBadRequest)
		return
	}

	tpl.InventoryID = &inv.ID
	err := helpers.Store(r).UpdateTemplate(tpl)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AttachInventory(w http.ResponseWriter, r *http.Request) {
	tpl := helpers.GetFromContext(r, "template").(db.Template)
	inv := helpers.GetFromContext(r, "inventory").(db.Inventory)

	if inv.TemplateID != nil {
		helpers.WriteErrorStatus(w, "Inventory is already attached to another template", http.StatusBadRequest)
		return
	}

	if !tpl.App.HasInventoryType(inv.Type) {
		helpers.WriteErrorStatus(w, "Inventory type is not supported for this template", http.StatusBadRequest)
		return
	}

	inv.TemplateID = &tpl.ID
	err := helpers.Store(r).UpdateInventory(inv)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DetachInventory(w http.ResponseWriter, r *http.Request) {
	tpl := helpers.GetFromContext(r, "template").(db.Template)
	inv := helpers.GetFromContext(r, "inventory").(db.Inventory)

	if inv.TemplateID == nil || *inv.TemplateID != tpl.ID {
		helpers.WriteErrorStatus(w, "Inventory is not attached to this template", http.StatusBadRequest)
		return
	}

	inv.TemplateID = nil
	err := helpers.Store(r).UpdateInventory(inv)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *TemplateController) GetTemplatePerms(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	tpl := helpers.GetFromContext(r, "template").(db.Template)

	perms, err := helpers.Store(r).GetTemplateRoles(project.ID, tpl.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, perms)
}

func (c *TemplateController) AddTemplatePerm(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)

	var perm db.TemplateRolePerm
	if !helpers.Bind(w, r, &perm) {
		return
	}

	perm.ProjectID = template.ProjectID
	perm.TemplateID = template.ID

	newPerm, err := c.templateRepo.CreateTemplateRole(perm)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newPerm)
}

func (c *TemplateController) UpdateTemplatePerm(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)
	permID, err := helpers.GetIntParam("perm_id", w, r)
	if err != nil {
		return
	}

	var perm db.TemplateRolePerm
	if !helpers.Bind(w, r, &perm) {
		return
	}

	perm.ID = permID
	perm.ProjectID = template.ProjectID
	perm.TemplateID = template.ID

	err = c.templateRepo.UpdateTemplateRole(perm)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *TemplateController) DeleteTemplatePerm(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)
	permID, err := helpers.GetIntParam("perm_id", w, r)
	if err != nil {
		return
	}

	err = c.templateRepo.DeleteTemplateRole(template.ProjectID, template.ID, permID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *TemplateController) GetTemplatePerm(w http.ResponseWriter, r *http.Request) {
	template := helpers.GetFromContext(r, "template").(db.Template)
	permID, err := helpers.GetIntParam("perm_id", w, r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	perm, err := c.templateRepo.GetTemplateRole(template.ProjectID, template.ID, permID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, perm)
}
