package sql

import (
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
	"math/rand"
)

func (d *SqlDb) CreateTaskStage(stage db.TaskStage) (db.TaskStage, error) {
	_, err := d.exec(
		"insert into task__stage (task_id, type) VALUES (?, ?, ?, ?)",
		stage.TaskID,
		stage.Type,
		stage.Start)
	return stage, err
}

func (d *SqlDb) GetTaskStages(projectID int, taskID int) ([]db.TaskStage, error) {
	return nil, nil
}

func (d *SqlDb) clearTasks(projectID int, templateID int, maxTasks int) {
	tpl, err := d.GetTemplate(projectID, templateID)
	if err != nil {
		return
	}

	nTasks := tpl.Tasks

	if rand.Intn(10) == 0 { // randomly recalculate number of tasks for the template
		var n int64
		n, err = d.sql.SelectInt("SELECT count(*) FROM task WHERE template_id=?", templateID)
		if err != nil {
			return
		}

		if n != int64(nTasks) {
			_, err = d.exec("UPDATE `project__template` SET `tasks`=? WHERE project_id=? and id=?",
				maxTasks, projectID, templateID)
			if err != nil {
				return
			}
		}

		nTasks = int(n)
	}

	if nTasks < maxTasks+maxTasks/10 { // deadzone of 10% for clearing of old tasks
		return
	}

	var oldestTask db.Task
	err = d.selectOne(&oldestTask,
		"SELECT created FROM task WHERE template_id=? ORDER BY created DESC LIMIT 1 OFFSET ?",
		templateID, maxTasks-1)

	if err != nil {
		return
	}

	_, err = d.exec("DELETE FROM task WHERE template_id=? AND created>?", templateID, oldestTask.Created)

	if err != nil {
		return
	}

	_, _ = d.exec("UPDATE `project__template` SET `tasks`=? WHERE project_id=? and id=?",
		maxTasks, projectID, templateID)
}

func (d *SqlDb) CreateTask(task db.Task, maxTasks int) (newTask db.Task, err error) {
	err = d.sql.Insert(&task)
	newTask = task

	if err != nil {
		return
	}

	_, err = d.exec("UPDATE `project__template` SET `tasks` = `tasks` + 1 WHERE project_id=? and id=?",
		task.ProjectID, task.TemplateID)

	if err != nil {
		return
	}

	if maxTasks > 0 {
		d.clearTasks(task.ProjectID, task.TemplateID, maxTasks)
	}

	return
}

func (d *SqlDb) UpdateTask(task db.Task) error {
	err := task.PreUpdate(d.sql)
	if err != nil {
		return err
	}

	_, err = d.exec(
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
		output.Time.UTC())
	return output, err
}

func (d *SqlDb) getTasks(projectID int, templateID *int, taskIDs []int, params db.RetrieveQueryParams, tasks *[]db.TaskWithTpl) (err error) {
	fields := "task.*"
	fields += ", tpl.playbook as tpl_playbook" +
		", `user`.name as user_name" +
		", tpl.name as tpl_alias" +
		", tpl.type as tpl_type" +
		", tpl.app as tpl_app"

	q := squirrel.Select(fields).
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("`user` on task.user_id=`user`.id").
		OrderBy("id desc")

	if templateID == nil {
		q = q.Where("tpl.project_id=?", projectID)
	} else {
		q = q.Where("tpl.project_id=? AND task.template_id=?", projectID, templateID)
	}

	if len(taskIDs) > 0 {
		q = q.Where(squirrel.Eq{"task.id": taskIDs})
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

	if err != nil {
		return
	}

	return
}

func (d *SqlDb) GetTemplateTasks(projectID int, templateID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	err = d.getTasks(projectID, &templateID, nil, params, &tasks)
	return
}

func (d *SqlDb) GetProjectTasks(projectID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	tasks = make([]db.TaskWithTpl, 0)
	err = d.getTasks(projectID, nil, nil, params, &tasks)
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
