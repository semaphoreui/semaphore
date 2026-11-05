package sql

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
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
			"update task set status=?, start=?, `end`=?, commit_hash=?, commit_message=?, runner_id=? where id=?",
			task.Status,
			task.Start,
			task.End,
			task.CommitHash,
			task.CommitMessage,
			task.RunnerID,
			task.ID)
	} else {
		_, err = d.exec(
			"update task set status=?, start=?, `end`=?, runner_id=? where id=?",
			task.Status,
			task.Start,
			task.End,
			task.RunnerID,
			task.ID)
	}

	return err
}

func (d *SqlDb) UpdateTaskArtifacts(projectID int, taskID int, artifacts *string) error {
	if _, err := d.GetTask(projectID, taskID); err != nil {
		return err
	}
	_, err := d.exec("update task set artifacts=? where id=?", artifacts, taskID)
	return err
}

func (d *SqlDb) SetWaitingTasksToStopped(projectID int, templateID int) error {
	_, err := d.exec(
		"update task set status=?, `end`=? where template_id=? and project_id=? and status=?",
		task_logger.TaskStoppedStatus,
		time.Now().UTC(),
		templateID,
		projectID,
		task_logger.TaskWaitingStatus)
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

// getTasks retrieves tasks for a given project, optionally filtered by template and/or task IDs.
// The taskIDs parameter has three-way semantics: nil means no filtering by ID,
// and a non-nil non-empty slice restricts the query to only those task IDs.
func (d *SqlDb) getTasks(projectID int, templateID *int, workflowRunID *int, taskIDs []int, params db.RetrieveQueryParams, tasks *[]db.TaskWithTpl) (err error) {

	if taskIDs != nil && len(taskIDs) == 0 {
		*tasks = []db.TaskWithTpl{}
		return nil
	}

	fields := "task.*"
	fields += ", tpl.playbook as tpl_playbook" +
		", `user`.name as user_name" +
		", task.runner_id as used_runner_id" +
		", runner.name as used_runner_name" +
		", tpl.name as tpl_alias" +
		", tpl.type as tpl_type" +
		", tpl.app as tpl_app"

	q := squirrel.Select(fields).
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("`user` on task.user_id=`user`.id").
		LeftJoin("runner on task.runner_id=runner.id").
		OrderBy("id desc")

	if params.TaskFilter != nil && len(params.TaskFilter.Status) > 0 {
		q = q.Where(squirrel.Eq{"status": params.TaskFilter.Status})
	}

	if params.TaskFilter != nil && filtered(params.TaskFilter.Viewer) {
		where, args := visibleTasksWhere(projectID, params.TaskFilter.Viewer.UserID)
		q = q.Where(where, args...)
	}

	if templateID == nil {
		q = q.Where("tpl.project_id=?", projectID)
	} else {
		q = q.Where("tpl.project_id=? AND task.template_id=?", projectID, templateID)
	}

	if workflowRunID != nil {
		q = q.Where("task.workflow_run_id=?", workflowRunID)
	}

	if taskIDs != nil {
		q = q.Where(squirrel.Eq{"task.id": taskIDs})
	}

	// Keyset (cursor) pagination: return only tasks strictly older than the
	// given cursor. Combined with the `id desc` ordering this walks backwards
	// through the table using the primary key index, so it stays cheap even
	// for projects with millions of tasks (unlike OFFSET, which scans and
	// discards all skipped rows).
	if params.BeforeID > 0 {
		q = q.Where("task.id < ?", params.BeforeID)
	}

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}

	query, args, _ := q.ToSql()

	_, err = d.selectAll(tasks, query, args...)

	for i := range *tasks {
		if params.TaskFilter != nil && filtered(params.TaskFilter.Viewer) {
			visible, err := d.taskVisible(projectID, (*tasks)[i].BuildTaskID, params.TaskFilter.Viewer.UserID)
			if err != nil {
				return err
			}
			if !visible {
				continue
			}
		}
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

func (d *SqlDb) GetTaskByID(taskID int) (task db.Task, err error) {
	err = d.selectOne(&task, d.PrepareQuery("select * from task where id=?"), taskID)
	return
}

func (d *SqlDb) GetTemplateTasks(projectID int, templateID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	err = d.getTasks(projectID, &templateID, nil, nil, params, &tasks)
	return
}

func (d *SqlDb) GetProjectTasks(projectID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	tasks = make([]db.TaskWithTpl, 0)
	err = d.getTasks(projectID, nil, nil, nil, params, &tasks)
	return
}

func (d *SqlDb) GetWorkflowRunTasks(projectID int, runID int, params db.RetrieveQueryParams) (tasks []db.TaskWithTpl, err error) {
	tasks = make([]db.TaskWithTpl, 0)
	err = d.getTasks(projectID, nil, &runID, nil, params, &tasks)
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

	return requireDeletedRow(d.exec("delete from task where id=?", taskID))
}

// deleteTemplateTasksBatch is how many tasks DeleteTemplateTasks removes in one transaction: it keeps
// the IN lists under the parameter limit of every dialect and the write lock short on SQLite, where
// the store has a single connection.
const deleteTemplateTasksBatch = 500

// DeleteTemplateTasks deletes the finished tasks of a template, with their output and task-bound
// keys, and returns how many it deleted. A task is finished by its status in the database, not by
// the task pool of this node: in HA mode it may be running on another node or a remote runner.
// It keeps the tasks of a workflow run that is still going, the build chains unfinished tasks
// read their version from and, for a build template, the newest task, which the next build takes
// its version from. Each batch is one transaction, so a long history never holds the database for
// long.
func (d *SqlDb) DeleteTemplateTasks(projectID int, templateID int) (deleted int, err error) {
	tpl, err := d.GetTemplate(projectID, templateID)
	if err != nil {
		return
	}

	// The batches committed before a failed one stay deleted, so the counter is repaired either way.
	defer func() {
		_, countErr := d.exec(
			"update project__template set tasks=(select count(*) from task where template_id=?) where project_id=? and id=?",
			templateID, projectID, templateID)
		if err == nil {
			err = countErr
		}
	}()

	finished := task_logger.FinishedTaskStatuses()

	// Read before the chains in use: every batch checks the tasks created after this one.
	lastTaskID, err := d.Sql().SelectInt(d.PrepareQuery("select coalesce(max(id), 0) from task where project_id=?"), projectID)
	if err != nil {
		return
	}

	keep := make(map[int]struct{})
	if err = d.addBuildChainsInUse(d.Sql(), projectID, squirrel.NotEq{"status": finished}, keep); err != nil {
		return
	}

	// AddTaskFrom numbers a new build after the newest task of the template; without it the
	// numbering would start over from the start version.
	if tpl.Type == db.TemplateBuild {
		var newest int64
		newest, err = d.Sql().SelectInt(
			d.PrepareQuery("select coalesce(max(id), 0) from task where project_id=? and template_id=?"),
			projectID, templateID)
		if err != nil {
			return
		}
		if newest > 0 {
			if err = d.addBuildChains(d.Sql(), []int{int(newest)}, keep); err != nil {
				return
			}
		}
	}

	activeRun, activeRunArgs, err := squirrel.Select("1").
		From("project__workflow_run r").
		Where("r.id = t.workflow_run_id").
		Where(squirrel.NotEq{"r.status": []db.WorkflowRunStatus{
			db.WorkflowRunSuccess, db.WorkflowRunStopped, db.WorkflowRunFailed,
		}}).
		ToSql()
	if err != nil {
		return
	}

	lastID := 0
	for {
		var query string
		var args []any
		query, args, err = squirrel.Select("t.id").
			From("task t").
			Where("t.project_id=? and t.template_id=?", projectID, templateID).
			Where(squirrel.Eq{"t.status": finished}).
			Where("t.id > ?", lastID).
			Where("not exists ("+activeRun+")", activeRunArgs...).
			OrderBy("t.id").
			Limit(deleteTemplateTasksBatch).
			ToSql()
		if err != nil {
			return
		}

		var ids []int
		if _, err = d.selectAll(&ids, query, args...); err != nil {
			return
		}
		if len(ids) == 0 {
			break
		}

		var n int
		n, err = d.deleteTasksBatch(projectID, templateID, ids, keep, int(lastTaskID))
		if err != nil {
			return
		}
		deleted += n

		if len(ids) < deleteTemplateTasksBatch {
			break
		}
		lastID = ids[len(ids)-1]
	}

	return
}

// addBuildChainsInUse adds to inUse the tasks that the tasks of the project matching referrers read
// their version from. GetIncomingVersion follows build_task_id from a deploy through finished
// deploys up to the build, so every task of such a chain is in use, not only the direct parent.
func (d *SqlDb) addBuildChainsInUse(ex gorp.SqlExecutor, projectID int, referrers squirrel.Sqlizer, inUse map[int]struct{}) error {
	query, args, err := squirrel.Select("build_task_id").
		From("task").
		Where("project_id=?", projectID).
		Where("build_task_id is not null").
		Where(referrers).
		ToSql()
	if err != nil {
		return err
	}

	var parents []int
	if _, err = ex.Select(&parents, d.PrepareQuery(query), args...); err != nil {
		return err
	}

	return d.addBuildChains(ex, parents, inUse)
}

// addBuildChains adds the given tasks to inUse with every task they read their version from.
func (d *SqlDb) addBuildChains(ex gorp.SqlExecutor, parents []int, inUse map[int]struct{}) error {
	for len(parents) > 0 {
		var next []int
		for _, id := range parents {
			if _, ok := inUse[id]; !ok {
				inUse[id] = struct{}{}
				next = append(next, id)
			}
		}
		if len(next) == 0 {
			return nil
		}

		query, args, err := squirrel.Select("build_task_id").
			From("task").
			Where(squirrel.Eq{"id": next}).
			Where("build_task_id is not null").
			ToSql()
		if err != nil {
			return err
		}

		parents = nil
		if _, err = ex.Select(&parents, d.PrepareQuery(query), args...); err != nil {
			return err
		}
	}

	return nil
}

// deleteTasksBatch deletes the given tasks of a template and what is bound to them in one
// transaction. Stages, stage results and Ansible hosts and errors go by ON DELETE CASCADE on every
// dialect. MySQL has no foreign key on access_key.task_id (added with a column REFERENCES, which
// MySQL ignores) nor on task.build_task_id (dropped by 2.8.42), so the task-bound keys are deleted
// and the references of other tasks cleared here. Tasks in keep are left out.
//
// keep was read before the first batch. build_task_id is set when a task is created and never
// changes, so only a task newer than lastTaskID, such as a deploy started meanwhile, can add a
// reference; those are added to keep inside the transaction, whatever their status.
func (d *SqlDb) deleteTasksBatch(projectID int, templateID int, candidates []int, keep map[int]struct{}, lastTaskID int) (deleted int, err error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = d.addBuildChainsInUse(tx, projectID, squirrel.Gt{"id": lastTaskID}, keep); err != nil {
		return
	}

	ids := make([]int, 0, len(candidates))
	for _, id := range candidates {
		if _, ok := keep[id]; !ok {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		err = tx.Commit()
		return
	}

	statements := []squirrel.Sqlizer{
		squirrel.Delete("access_key").Where(squirrel.Eq{
			"project_id": projectID,
			"owner":      db.AccessKeyTaskSecret,
			"task_id":    ids,
		}),
		squirrel.Update("task").Set("build_task_id", nil).Where(squirrel.Eq{"build_task_id": ids}),
		squirrel.Delete("task__output").Where(squirrel.Eq{"task_id": ids}),
	}

	for _, statement := range statements {
		var query string
		var args []any
		if query, args, err = statement.ToSql(); err != nil {
			return
		}
		if _, err = d.execTx(tx, query, args...); err != nil {
			return
		}
	}

	query, args, err := squirrel.Delete("task").Where(squirrel.Eq{
		"project_id":  projectID,
		"template_id": templateID,
		"id":          ids,
	}).ToSql()
	if err != nil {
		return
	}

	res, err := d.execTx(tx, query, args...)
	if err != nil {
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	deleted = int(affected)
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

func (d *SqlDb) GetNodeCount() (int, error) {
	return 0, nil
}

func (d *SqlDb) GetUiCount() (int, error) {
	return 1, nil
}
