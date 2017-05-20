package projects

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func TemplatesMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	templateID, err := util.GetIntParam("template_id", w, r)
	if err != nil {
		return
	}

	var template db.Template
	if err := db.Mysql.SelectOne(&template, "select * from project__template where project_id=? and id=?", project.ID, templateID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "template", template)
}

func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var templates []db.Template

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != "asc" && order != "desc" {
		order = "asc"
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
			OrderBy("pt."+ sort + " " + order)
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

	query, args, _ := q.ToSql()

	if _, err := db.Mysql.Select(&templates, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, templates)
}

func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var template db.Template
	if err := mulekick.Bind(w, r, &template); err != nil {
		return
	}

	res, err := db.Mysql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", template.SshKeyID, project.ID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments)
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
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &template.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, template)
}

func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := context.Get(r, "template").(db.Template)

	var template db.Template
	if err := mulekick.Bind(w, r, &template); err != nil {
		return
	}

	if template.Arguments != nil && *template.Arguments == "" {
		template.Arguments = nil
	}

	if _, err := db.Mysql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?", template.SshKeyID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments, oldTemplate.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(template.ID) + " updated"
	objType := "template"
	if err := (db.Event{
		ProjectID:   &oldTemplate.ProjectID,
		Description: &desc,
		ObjectID:    &oldTemplate.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(db.Template)

	if _, err := db.Mysql.Exec("delete from project__template where id=?", tpl.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(tpl.ID) + " deleted"
	if err := (db.Event{
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
