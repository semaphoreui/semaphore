package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
)

func (d *SqlDb) CreateTask(task db.Task) (db.Task, error) {
	err := d.sql.Insert(&task)
	return task, err
}

func (d *SqlDb) UpdateTask(task db.Task) error {
	_, err := d.exec(
		"update task set status=?, start=?, `end`=? where id=?",
		task.Status,
		task.Start,
		task.End,
		task.ID)

	return err
}

func (d *SqlDb) CreateTaskOutput(output db.TaskOutput) (db.TaskOutput, error) {
	_, err := d.exec(
		"insert into task__output (task_id, task, output, time) VALUES (?, '', ?, ?)",
		output.TaskID,
		output.Output,
		output.Time)
	return output, err
}


func (d *SqlDb) getTasks(projectID int, templateID* int, params db.RetrieveQueryParams, tasks *[]db.TaskWithTpl) (err error) {
	fields := "task.*"
	fields += ", tpl.playbook as tpl_playbook" +
		", `user`.name as user_name" +
		", tpl.alias as tpl_alias" +
		", tpl.type as tpl_type"

	q := squirrel.Select(fields).
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("`user` on task.user_id=`user`.id").
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

	_, err = d.selectAll(tasks, query, args...)

	for i := range *tasks {
		err = (*tasks)[i].Fill(d)
		if err != nil {
			return
		}
	}

	return
}


func (d *SqlDb) GetTask(projectID int, taskID int) (task db.Task, err error) {
	q := squirrel.Select("task.*").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		Where("tpl.project_id=? AND task.id=?", projectID, taskID)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&task, query, args...)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
		return
	}

	err = task.Fill(d)
	if err != nil {
		return
	}

	return
}

func (d *SqlDb) GetTemplateTasks(template db.Template, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	err = d.getTasks(template.ProjectID, &template.ID, params, &tasks)
	return
}

func (d *SqlDb) GetProjectTasks(projectID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	err = d.getTasks(projectID, nil, params, &tasks)
	return
}

func (d *SqlDb) DeleteTaskWithOutputs(projectID int, taskID int) (err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	_, err = d.exec("delete from task__output where task_id=?", taskID)

	if err != nil {
		return
	}

	_, err = d.exec("delete from task where id=?", taskID)
	return
}

func (d *SqlDb) GetTaskOutputs(projectID int, taskID int) (output []db.TaskOutput, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	_, err = d.selectAll(&output,
		"select task_id, task, time, output from task__output where task_id=? order by time asc",
		taskID)
	return
}
