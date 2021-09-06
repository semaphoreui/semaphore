package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
)

func (d *SqlDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__template (project_id, inventory_id, repository_id, environment_id, alias, playbook, arguments, override_args)" +
			"values (?, ?, ?, ?, ?, ?, ?, ?)",
		template.ProjectID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments)

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
	_, err := d.exec("update project__template set inventory_id=?, repository_id=?, environment_id=?, alias=?, " +
		"playbook=?, arguments=?, override_args=? where removed = false and id=? and project_id=?",
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments,
		template.ID,
		template.ProjectID)
	
	//if err != nil {
	//	return err
	//}
	//
	//if template.CronFormat == "" {
	//	_, err = d.exec(
	//		"delete from project__template_schedule where project_id =? and template_id=?",
	//		template.ProjectID,
	//		template.ID)
	//} else {
	//	_, err = d.GetTemplateSchedules(template.ProjectID, template.ID)
	//	if err == nil {
	//		_, err = d.exec(
	//			"update project__template_schedule set cron_format=? where project_id =? and template_id=?",
	//			template.CronFormat,
	//			template.ProjectID,
	//			template.ID)
	//	} else if err == db.ErrNotFound {
	//		_, err = d.exec(
	//			"insert into project__template_schedule (template_id, cron_format) values (?, ?)",
	//			template.ID,
	//			template.CronFormat)
	//	}
	//}
	return err
}

func (d *SqlDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	q := squirrel.Select("pt.id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.alias",
		"pt.playbook",
		"pt.arguments",
		"pt.override_args").
		From("project__template pt").
		Where("pt.removed = false")

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "alias", "playbook":
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
			OrderBy("pt.alias " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&templates, query, args...)
	return
}

func (d *SqlDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.selectOne(
		&template,
		"select * from project__template where project_id=? and id=? and removed = false",
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
	_, err := d.exec("update project__template set removed=true where project_id=? and id=?", projectID, templateID)

	return err
}
