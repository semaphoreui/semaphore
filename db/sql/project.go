package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
	"time"
)

func (d *SqlDb) CreateProject(project db.Project) (newProject db.Project, err error) {
	project.Created = time.Now().UTC()

	insertId, err := d.insert(
		"id",
		"insert into project(name, created, type) values (?, ?, ?)",
		project.Name, project.Created, project.Type)

	if err != nil {
		return
	}

	newProject = project
	newProject.ID = insertId
	return
}

func (d *SqlDb) GetAllProjects() (projects []db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		OrderBy("p.name").
		Limit(200).
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&projects, query, args...)

	return
}

func (d *SqlDb) GetProjects(userID int) (projects []db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("pu.user_id=?", userID).
		OrderBy("p.name").
		Limit(200).
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&projects, query, args...)

	return
}

func (d *SqlDb) GetProject(projectID int) (project db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Where("p.id=?", projectID).
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&project, query, args...)

	return
}

func (d *SqlDb) DeleteProject(projectID int) error {

	//tpls, err := d.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	//
	//if err != nil {
	//	return err
	//}
	// TODO: sort projects

	tx, err := d.sql.Begin()

	if err != nil {
		return err
	}

	statements := []string{
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err = tx.Exec(d.PrepareQuery(statement), projectID)

		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (d *SqlDb) UpdateProject(project db.Project) error {
	_, err := d.exec(
		"update project set name=?, alert=?, alert_chat=?, max_parallel_tasks=? where id=?",
		project.Name,
		project.Alert,
		project.AlertChat,
		project.MaxParallelTasks,
		project.ID)
	return err
}
