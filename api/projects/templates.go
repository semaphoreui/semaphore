package projects

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// TemplatesMiddleware ensures a template exists and loads it to the context
func TemplatesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(models.Project)
		templateID, err := util.GetIntParam("template_id", w, r)
		if err != nil {
			return
		}

		var template models.Template
		if err := context.Get(r, "store").(db.Store).Sql().SelectOne(&template, "select * from project__template where project_id=? and id=?", project.ID, templateID); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "template", template)
		next.ServeHTTP(w, r)
	})
}

// GetTemplate returns single template by ID
func GetTemplate(w http.ResponseWriter, r *http.Request) {
	template := context.Get(r, "template").(models.Template)
	util.WriteJSON(w, http.StatusOK, template)
}

// GetTemplates returns all templates for a project in a sort order
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var templates []models.Template

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != asc && order != desc {
		order = asc
	}

	q := squirrel.Select("pt.id",
		"pt.ssh_key_id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.alias",
		"pt.playbook",
		"pt.arguments",
		"pt.override_args").
		From("project__template pt")

	switch sort {
	case "alias", "playbook":
		q = q.Where("pt.project_id=?", project.ID).
			OrderBy("pt." + sort + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pt.ssh_key_id = ak.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("ak.name " + order)
	case "inventory":
		q = q.LeftJoin("project__inventory pi ON (pt.inventory_id = pi.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pi.name " + order)
	case "environment":
		q = q.LeftJoin("project__environment pe ON (pt.environment_id = pe.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pe.name " + order)
	case "repository":
		q = q.LeftJoin("project__repository pr ON (pt.repository_id = pr.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pr.name " + order)
	default:
		q = q.Where("pt.project_id=?", project.ID).
			OrderBy("pt.alias " + order)
	}

	query, args, err := q.ToSql()
	util.LogWarning(err)

	if _, err := context.Get(r, "store").(db.Store).Sql().Select(&templates, query, args...); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, templates)
}

// AddTemplate adds a template to the database
func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)

	var template models.Template
	if err := util.Bind(w, r, &template); err != nil {
		return
	}

	res, err := context.Get(r, "store").(db.Store).Sql().Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", template.SSHKeyID, project.ID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	template.ID = int(insertID)

	objType := "template"
	desc := "Template ID " + strconv.Itoa(template.ID) + " created"

	_, err = context.Get(r, "store").(db.Store).CreateEvent(models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &template.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusCreated, template)
}

// UpdateTemplate writes a template to an existing key in the database
func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := context.Get(r, "template").(models.Template)

	var template models.Template
	if err := util.Bind(w, r, &template); err != nil {
		return
	}

	if template.Arguments != nil && *template.Arguments == "" {
		template.Arguments = nil
	}

	if _, err := context.Get(r, "store").(db.Store).Sql().Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?", template.SSHKeyID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments, oldTemplate.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(template.ID) + " updated"
	objType := "template"

	_, err := context.Get(r, "store").(db.Store).CreateEvent(models.Event{
		ProjectID:   &oldTemplate.ProjectID,
		Description: &desc,
		ObjectID:    &oldTemplate.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTemplate deletes a template from the database
func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(models.Template)

	if _, err := context.Get(r, "store").(db.Store).Sql().Exec("delete from project__template where id=?", tpl.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(tpl.ID) + " deleted"

	_, err := context.Get(r, "store").(db.Store).CreateEvent(models.Event{
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
