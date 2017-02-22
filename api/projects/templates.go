package projects

import (
	"database/sql"
	"strconv"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func TemplatesMiddleware(w http.ResponseWriter, r *http.Request) {
	project := c.MustGet("project").(models.Project)
	templateID, err := util.GetIntParam("template_id", c)
	if err != nil {
		return
	}

	var template models.Template
	if err := database.Mysql.SelectOne(&template, "select * from project__template where project_id=? and id=?", project.ID, templateID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	c.Set("template", template)
	c.Next()
}

func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := c.MustGet("project").(models.Project)
	var templates []models.Template

	q := squirrel.Select("*").
		From("project__template").
		Where("project_id=?", project.ID)

	query, args, _ := q.ToSql()

	if _, err := database.Mysql.Select(&templates, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, templates)
}

func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := c.MustGet("project").(models.Project)

	var template models.Template
	if err := c.Bind(&template); err != nil {
		return
	}

	res, err := database.Mysql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", template.SshKeyID, project.ID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments)
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
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &template.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.JSON(201, template)
}

func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := c.MustGet("template").(models.Template)

	var template models.Template
	if err := c.Bind(&template); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?", template.SshKeyID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments, oldTemplate.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(template.ID) + " updated"
	objType := "template"
	if err := (models.Event{
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
	tpl := c.MustGet("template").(models.Template)

	if _, err := database.Mysql.Exec("delete from project__template where id=?", tpl.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(tpl.ID) + " deleted"
	if err := (models.Event{
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
