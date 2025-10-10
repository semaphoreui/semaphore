package sql

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateTaskStage(stage db.TaskStage) (res db.TaskStage, err error) {
	insertID, err := d.insert(
		"id",
		"insert into task__stage "+
			"(task_id, `start`, `end`, `type`) VALUES "+
			"(?, ?, ?, ?)",
		stage.TaskID,
		stage.Start,
		stage.End,
		stage.Type)

	if err != nil {
		return
	}

	res = stage
	res.ID = insertID
	return
}

func (d *SqlDb) EndTaskStage(taskID int, stageID int, end time.Time) (err error) {
	_, err = d.exec(
		"update task__stage set `end`=? where task_id=? and id=?",
		end,
		taskID,
		stageID)

	return
}

func (d *SqlDb) CreateTaskStageResult(taskID int, stageID int, result map[string]any) (err error) {
	jsn, err := json.Marshal(result)
	if err != nil {
		return
	}

	_, err = d.insert(
		"id",
		"insert into task__stage_result "+
			"(task_id, stage_id, `json`) VALUES "+
			"(?, ?, ?)",
		taskID,
		stageID,
		string(jsn))

	return
}

func (d *SqlDb) getTaskStage(taskID int, stageID int) (res db.TaskStage, err error) {
	err = d.selectOne(
		&res,
		"select * from task__stage where task_id=? and id=?",
		taskID,
		stageID)

	return
}

func (d *SqlDb) validateTask(projectID int, taskID int) error {
	_, err := d.GetTask(projectID, taskID)

	return err
}

func (d *SqlDb) GetTaskStageResult(projectID int, taskID int, stageID int) (res db.TaskStageResult, err error) {

	if err = d.validateTask(projectID, taskID); err != nil {
		return
	}

	err = d.selectOne(
		&res,
		"select * from task__stage_result where task_id=? and stage_id=?",
		taskID,
		stageID)

	return
}

func (d *SqlDb) getTaskStages(projectID int, taskID int, stageType *db.TaskStageType) (res []db.TaskStageWithResult, err error) {
	if err = d.validateTask(projectID, taskID); err != nil {
		return
	}

	q := squirrel.Select("p.*, pu.json").
		From("task__stage as p").
		Join("task__stage_result as pu on pu.stage_id=p.id").
		Where("pu.task_id=?", taskID)

	if stageType != nil {
		q = q.Where(squirrel.Eq{"type": *stageType})
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&res, query, args...)

	return
}

func (d *SqlDb) GetTaskStages(projectID int, taskID int) ([]db.TaskStageWithResult, error) {
	return d.getTaskStages(projectID, taskID, nil)
}

func (d *SqlDb) clearTasks(projectID int, templateID int, maxTasks int) {
	tpl, err := d.GetTemplate(projectID, templateID)
	if err != nil {
		return
	}

	nTasks := tpl.Tasks

	if rand.Intn(10) == 0 { // randomly recalculate number of tasks for the template
		var n int64
		n, err = d.Sql().SelectInt("SELECT count(*) FROM task WHERE template_id=?", templateID)
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

	_, err = d.exec("DELETE FROM task WHERE template_id=? AND created<?", templateID, oldestTask.Created)

	if err != nil {
		return
	}

	_, _ = d.exec("UPDATE `project__template` SET `tasks`=? WHERE project_id=? and id=?",
		maxTasks, projectID, templateID)
}

func (d *SqlDb) CreateTask(task db.Task, maxTasks int) (newTask db.Task, err error) {
	err = d.Sql().Insert(&task)
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
	err := task.PreUpdate(d.Sql())
	if err != nil {
		return err
	}

	if task.CommitHash != nil {
		_, err = d.exec(
			"update task set status=?, start=?, `end`=?, commit_hash=?, commit_message=? where id=?",
			task.Status,
			task.Start,
			task.End,
			task.CommitHash,
			task.CommitMessage,
			task.ID)
	} else {
		_, err = d.exec(
			"update task set status=?, start=?, `end`=? where id=?",
			task.Status,
			task.Start,
			task.End,
			task.ID)
	}

	return err
}

func (d *SqlDb) CreateTaskOutput(output db.TaskOutput) (db.TaskOutput, error) {
	insertID, err := d.insert(
		"id",
		"insert into task__output (task_id, output, time) VALUES (?, ?, ?)",
		output.TaskID,
		output.Output,
		output.Time.UTC())

	output.ID = insertID
	return output, err
}

func (d *SqlDb) InsertTaskOutputBatch(output []db.TaskOutput) error {

	if len(output) == 0 {
		return nil
	}

	q := squirrel.Insert("task__output").
		Columns("task_id", "output", "time", "stage_id")

	for _, item := range output {
		q = q.Values(item.TaskID, item.Output, item.Time.UTC(), item.StageID)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}

	_, err = d.exec(query, args...)
	return err
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

	if params.TaskFilter != nil && len(params.TaskFilter.Status) > 0 {
		q = q.Where(squirrel.Eq{"status": params.TaskFilter.Status})
	}

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

func (d *SqlDb) GetTaskOutputs(projectID int, taskID int, params db.RetrieveQueryParams) (output []db.TaskOutput, err error) {

	if err = d.validateTask(projectID, taskID); err != nil {
		return
	}

	q := squirrel.Select("task_id", "time", "output").
		From("task__output").
		Where("task_id=?", taskID).
		OrderBy("time, id")

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count)).Offset(uint64(params.Offset))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&output, query, args...)
	return
}

func (d *SqlDb) GetTaskStageOutputs(projectID int, taskID int, stageID int) (output []db.TaskOutput, err error) {

	if err = d.validateTask(projectID, taskID); err != nil {
		return
	}

	q := squirrel.Select("id", "task_id", "time", "output").
		From("task__output").
		Where("task_id=?", taskID).
		Where("stage_id=?", stageID)

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&output, query, args...)
	return
}
