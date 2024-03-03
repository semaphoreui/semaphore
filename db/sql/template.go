package sql

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/Masterminds/squirrel"
)

func (d *SqlDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	err = template.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__template (project_id, inventory_id, repository_id, environment_id, "+
			"name, playbook, arguments, allow_override_args_in_task, description, vault_key_id, `type`, start_version,"+
			"build_template_id, view_id, autorun, survey_vars, suppress_success_alerts, app)"+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		template.ProjectID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Name,
		template.Playbook,
		template.Arguments,
		template.AllowOverrideArgsInTask,
		template.Description,
		template.VaultKeyID,
		template.Type,
		template.StartVersion,
		template.BuildTemplateID,
		template.ViewID,
		template.Autorun,
		db.ObjectToJSON(template.SurveyVars),
		template.SuppressSuccessAlerts,
		template.App)

	if err != nil {
		return
	}

	err = db.FillTemplate(d, &newTemplate)

	if err != nil {
		return
	}

	newTemplate = template
	newTemplate.ID = insertID

	return
}

func (d *SqlDb) UpdateTemplate(template db.Template) error {
	err := template.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec("update project__template set "+
		"inventory_id=?, "+
		"repository_id=?, "+
		"environment_id=?, "+
		"name=?, "+
		"playbook=?, "+
		"arguments=?, "+
		"allow_override_args_in_task=?, "+
		"description=?, "+
		"vault_key_id=?, "+
		"`type`=?, "+
		"start_version=?,"+
		"build_template_id=?, "+
		"view_id=?, "+
		"autorun=?, "+
		"survey_vars=?, "+
		"suppress_success_alerts=?, "+
		"app=? "+
		"where id=? and project_id=?",
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Name,
		template.Playbook,
		template.Arguments,
		template.AllowOverrideArgsInTask,
		template.Description,
		template.VaultKeyID,
		template.Type,
		template.StartVersion,
		template.BuildTemplateID,
		template.ViewID,
		template.Autorun,
		db.ObjectToJSON(template.SurveyVars),
		template.SuppressSuccessAlerts,
		template.App,
		template.ID,
		template.ProjectID,
	)
	return err
}

func (d *SqlDb) GetTemplates(projectID int, filter db.TemplateFilter, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	q := squirrel.Select("pt.id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.name",
		"pt.playbook",
		"pt.arguments",
		"pt.allow_override_args_in_task",
		"pt.vault_key_id",
		"pt.build_template_id",
		"pt.start_version",
		"pt.view_id",
		"pt.`app`",
		"pt.survey_vars",
		"pt.start_version",
		"pt.`type`").
		From("project__template pt")

	if filter.ViewID != nil {
		q = q.Where("pt.view_id=?", *filter.ViewID)
	}

	if filter.BuildTemplateID != nil {
		q = q.Where("pt.build_template_id=?", *filter.BuildTemplateID)
		if filter.AutorunOnly {
			q = q.Where("pt.autorun=true")
		}
	}

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "name", "playbook":
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt." + params.SortBy + " " + order)
	case "inventory":
		q = q.LeftJoin("project__inventory pi ON (pt.inventory_id = pi.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pi.name " + order)
	case "environment":
		q = q.LeftJoin("project__environment pe ON (pt.environment_id = pe.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pe.name " + order)
	case "repository":
		q = q.LeftJoin("project__repository pr ON (pt.repository_id = pr.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pr.name " + order)
	default:
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt.name " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&templates, query, args...)

	if err != nil {
		return
	}

	err = db.FillTemplates(d, templates)

	return
}

func (d *SqlDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.selectOne(
		&template,
		"select * from project__template where project_id=? and id=?",
		projectID,
		templateID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	if err != nil {
		return
	}

	err = db.FillTemplate(d, &template)
	return
}

func (d *SqlDb) DeleteTemplate(projectID int, templateID int) error {
	_, err := d.exec("delete from project__template where project_id=? and id=?", projectID, templateID)
	return err
}

func (d *SqlDb) GetTemplateRefs(projectID int, templateID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.TemplateProps, templateID)
}
