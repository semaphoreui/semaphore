package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
)

func (d *SqlDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	res, err := d.exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?",
		template.SSHKeyID,
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

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newTemplate = template
	newTemplate.ID = int(insertID)
	return
}

func (d *SqlDb) UpdateTemplate(template db.Template) error {
	_, err := d.exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?",
		template.SSHKeyID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments,
		template.ID)

	return err
}

func (d *SqlDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
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

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "alias", "playbook":
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt." + params.SortBy + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pt.ssh_key_id = ak.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("ak.name " + order)
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

	_, err = d.sql.Select(&templates, query, args...)
	return
}

func (d *SqlDb) GetTemplate(projectID int, templateID int) (db.Template, error) {
	var template db.Template

	err := d.selectOne(
		&template,
		"select * from project__template where project_id=? and id=?",
		projectID,
		templateID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return template, err
}

func (d *SqlDb) DeleteTemplate(projectID int, templateID int) error {
	res, err := d.exec(
		"delete from project__template where project_id=? and id=?",
		projectID,
		templateID)

	return validateMutationResult(res, err)
}
