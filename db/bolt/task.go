package bolt

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
)

func (d *BoltDb) CreateTask(task db.Task) (db.Task, error) {
	err := d.sql.Insert(&task)
	return task, err
}

func (d *BoltDb) UpdateTask(task db.Task) error {
	_, err := d.sql.Exec(
		"update task set status=?, start=?, end=? where id=?",
		task.Status,
		task.Start,
		task.End,
		task.ID)

	return err
}

func (d *BoltDb) CreateTaskOutput(output db.TaskOutput) (db.TaskOutput, error) {
	_, err := d.sql.Exec(
		"insert into task__output (task_id, task, output, time) VALUES (?, '', ?, ?)",
		output.TaskID,
		output.Output,
		output.Time)
	return output, err
}

func (d *BoltDb) getTasks(projectID int, templateID* int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	q := squirrel.Select("task.*, tpl.playbook as tpl_playbook, user.name as user_name, tpl.alias as tpl_alias").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("user on task.user_id=user.id").
		OrderBy("task.created desc, id desc")

	if templateID == nil {
		q = q.Where("tpl.project_id=?", projectID)
	} else {
		q = q.Where("tpl.project_id=? AND task.template_id=?", projectID, templateID)
	}

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}

	query, args, _ := q.ToSql()

	_, err = d.sql.Select(&tasks, query, args...)

	return
}

func (d *BoltDb) GetTask(projectID int, taskID int) (task db.Task, err error) {
	q := squirrel.Select("task.*").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		Where("tpl.project_id=? AND task.id=?", projectID, taskID)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	err = d.sql.SelectOne(&task, query, args...)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *BoltDb) GetTemplateTasks(projectID int, templateID int, params db.RetrieveQueryParams) ([]db.TaskWithTpl, error) {
	return d.getTasks(projectID, &templateID, params)
}

func (d *BoltDb) GetProjectTasks(projectID int, params db.RetrieveQueryParams) ([]db.TaskWithTpl, error) {
	return d.getTasks(projectID, nil, params)
}

func (d *BoltDb) DeleteTaskWithOutputs(projectID int, taskID int) (err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	_, err = d.sql.Exec("delete from task__output where task_id=?", taskID)

	if err != nil {
		return
	}

	_, err = d.sql.Exec("delete from task where id=?", taskID)
	return
}

func (d *BoltDb) GetTaskOutputs(projectID int, taskID int) (output []db.TaskOutput, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	_, err = d.sql.Select(&output,
		"select task_id, task, time, output from task__output where task_id=? order by time asc",
		taskID)
	return
}
