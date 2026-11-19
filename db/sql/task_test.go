package sql

import (
	"fmt"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type templateTasksFixture struct {
	store     *SqlDb
	projectID int
	repoID    int
	tpl       db.Template
	other     db.Template
}

func newTemplateTasksFixture(t *testing.T, store *SqlDb) templateTasksFixture {
	t.Helper()
	projectID, repoID := newTemplateTestProject(t, store)
	f := templateTasksFixture{store: store, projectID: projectID, repoID: repoID}
	f.tpl = f.template(t, projectID, "deploy")
	f.other = f.template(t, projectID, "other")
	return f
}

func (f templateTasksFixture) template(t *testing.T, projectID int, name string) db.Template {
	t.Helper()
	tpl, err := f.store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: f.repoID, Name: name, Playbook: name + ".yml"})
	require.NoError(t, err)
	return tpl
}

// task creates a task with output, a stage with a result, an Ansible host row and a task-bound key.
func (f templateTasksFixture) task(t *testing.T, tpl db.Template, status task_logger.TaskStatus, mod func(*db.Task)) db.Task {
	t.Helper()
	task := db.Task{ProjectID: tpl.ProjectID, TemplateID: tpl.ID, Status: status, Created: time.Now()}
	if mod != nil {
		mod(&task)
	}
	task, err := f.store.CreateTask(task, 0)
	require.NoError(t, err)

	require.NoError(t, f.store.InsertTaskOutputBatch([]db.TaskOutput{
		{TaskID: task.ID, Time: time.Now(), Output: "line 1"},
		{TaskID: task.ID, Time: time.Now().Add(time.Millisecond), Output: "line 2"},
	}))
	stage, err := f.store.CreateTaskStage(db.TaskStage{TaskID: task.ID, Type: db.TaskStageInit})
	require.NoError(t, err)
	require.NoError(t, f.store.CreateTaskStageResult(task.ID, stage.ID, map[string]any{"ok": true}))
	_, err = f.store.exec(
		"insert into task__ansible_host (task_id, project_id, host, failed, ignored, changed, ok, rescued, skipped, unreachable) values (?, ?, 'h', 0, 0, 0, 1, 0, 0, 0)",
		task.ID, tpl.ProjectID)
	require.NoError(t, err)
	_, err = f.store.CreateAccessKey(db.AccessKey{
		Name:      fmt.Sprintf("task-%d", task.ID),
		Type:      db.AccessKeyString,
		ProjectID: &tpl.ProjectID,
		Owner:     db.AccessKeyTaskSecret,
		TaskID:    &task.ID,
	})
	require.NoError(t, err)
	return task
}

func (f templateTasksFixture) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	n, err := f.store.Sql().SelectInt(f.store.PrepareQuery(query), args...)
	require.NoError(t, err)
	return int(n)
}

// bound counts the rows of every table bound to the task.
func (f templateTasksFixture) bound(t *testing.T, taskID int) map[string]int {
	t.Helper()
	res := map[string]int{}
	for _, table := range []string{"task", "task__output", "task__stage", "task__stage_result", "task__ansible_host", "access_key"} {
		column := "task_id"
		if table == "task" {
			column = "id"
		}
		res[table] = f.count(t, "select count(*) from "+table+" where "+column+"=?", taskID)
	}
	return res
}

func (f templateTasksFixture) counter(t *testing.T, tpl db.Template) int {
	t.Helper()
	return f.count(t, "select tasks from project__template where id=?", tpl.ID)
}

func gone() map[string]int {
	return map[string]int{"task": 0, "task__output": 0, "task__stage": 0, "task__stage_result": 0, "task__ansible_host": 0, "access_key": 0}
}

func kept() map[string]int {
	return map[string]int{"task": 1, "task__output": 2, "task__stage": 1, "task__stage_result": 1, "task__ansible_host": 1, "access_key": 1}
}

func testDeleteTemplateTasksDeletesOnlyFinishedTasksOfTheTemplate(t *testing.T, f templateTasksFixture) {

	var finished []db.Task
	for _, status := range task_logger.FinishedTaskStatuses() {
		finished = append(finished, f.task(t, f.tpl, status, nil))
	}
	var unfinished []db.Task
	for _, status := range task_logger.UnfinishedTaskStatuses() {
		unfinished = append(unfinished, f.task(t, f.tpl, status, nil))
	}
	otherTemplate := f.task(t, f.other, task_logger.TaskSuccessStatus, nil)

	otherProjectID, _ := newTemplateTestProject(t, f.store)
	otherProjectTpl := f.template(t, otherProjectID, "foreign")
	otherProject := f.task(t, otherProjectTpl, task_logger.TaskSuccessStatus, nil)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, len(finished), deleted)

	for _, task := range finished {
		assert.Equal(t, gone(), f.bound(t, task.ID), "finished task %d (%s)", task.ID, task.Status)
	}
	for _, task := range unfinished {
		assert.Equal(t, kept(), f.bound(t, task.ID), "unfinished task %d (%s)", task.ID, task.Status)
	}
	assert.Equal(t, kept(), f.bound(t, otherTemplate.ID))
	assert.Equal(t, kept(), f.bound(t, otherProject.ID))

	assert.Equal(t, len(unfinished), f.counter(t, f.tpl))
	assert.Equal(t, 1, f.counter(t, f.other))
	assert.Equal(t, 1, f.counter(t, otherProjectTpl))

	// A second call finds nothing more to delete.
	deleted, err = f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)
	require.NoError(t, err)
	assert.Zero(t, deleted)
	assert.Equal(t, len(unfinished), f.count(t, "select count(*) from task where template_id=?", f.tpl.ID))
	assert.Equal(t, len(unfinished), f.counter(t, f.tpl))
}

func testDeleteTemplateTasksNotFound(t *testing.T, f templateTasksFixture) {
	task := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	otherProjectID, _ := newTemplateTestProject(t, f.store)

	tests := []struct {
		name       string
		projectID  int
		templateID int
	}{
		{"template of another project", otherProjectID, f.tpl.ID},
		{"unknown template", f.projectID, f.tpl.ID + 1000},
		{"unknown project", otherProjectID + 1000, f.tpl.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := f.store.DeleteTemplateTasks(tt.projectID, tt.templateID)
			assert.ErrorIs(t, err, db.ErrNotFound)
			assert.Zero(t, deleted)
			assert.Equal(t, kept(), f.bound(t, task.ID))
			assert.Equal(t, 1, f.counter(t, f.tpl))
		})
	}
}

func testDeleteTemplateTasksEmptyHistory(t *testing.T, f templateTasksFixture) {

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)

	require.NoError(t, err)
	assert.Zero(t, deleted)
	assert.Zero(t, f.counter(t, f.tpl))
}

func testDeleteTemplateTasksRepairsTheCounter(t *testing.T, f templateTasksFixture) {
	f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	f.task(t, f.tpl, task_logger.TaskRunningStatus, nil)
	_, err := f.store.exec("update project__template set tasks=42 where id=?", f.tpl.ID)
	require.NoError(t, err)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, 1, f.counter(t, f.tpl))
}

func testDeleteTemplateTasksKeepsTasksOfUnfinishedWorkflowRuns(t *testing.T, f templateTasksFixture) {
	workflowID := insertWorkflow(t, f.store, f.projectID, "release")
	run := func(status db.WorkflowRunStatus) int {
		id, err := f.store.insert("id",
			"insert into project__workflow_run (project_id, workflow_template_id, status) values (?, ?, ?)",
			f.projectID, workflowID, status)
		require.NoError(t, err)
		return id
	}
	inRun := func(runID int) func(*db.Task) {
		return func(task *db.Task) { task.WorkflowRunID = &runID }
	}

	running, approval, done := run(db.WorkflowRunRunning), run(db.WorkflowRunApproval), run(db.WorkflowRunSuccess)
	ofRunning := f.task(t, f.tpl, task_logger.TaskSuccessStatus, inRun(running))
	ofApproval := f.task(t, f.tpl, task_logger.TaskSuccessStatus, inRun(approval))
	ofDone := f.task(t, f.tpl, task_logger.TaskSuccessStatus, inRun(done))
	_, err := f.store.exec("update project__workflow_run set root_task_id=? where id=?", ofDone.ID, done)
	require.NoError(t, err)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, kept(), f.bound(t, ofRunning.ID))
	assert.Equal(t, kept(), f.bound(t, ofApproval.ID))
	assert.Equal(t, gone(), f.bound(t, ofDone.ID))
	// The finished run stays, without its root task.
	assert.Equal(t, 1, f.count(t, "select count(*) from project__workflow_run where id=? and root_task_id is null", done))
}

func testDeleteTemplateTasksBuildTasks(t *testing.T, f templateTasksFixture) {
	usedByWaiting := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	usedByFinished := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	waitingDeploy := f.task(t, f.other, task_logger.TaskWaitingStatus, func(task *db.Task) { task.BuildTaskID = &usedByWaiting.ID })
	finishedDeploy := f.task(t, f.other, task_logger.TaskSuccessStatus, func(task *db.Task) { task.BuildTaskID = &usedByFinished.ID })

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	// The waiting deployment still needs its build to read the version.
	assert.Equal(t, kept(), f.bound(t, usedByWaiting.ID))
	assert.Equal(t, gone(), f.bound(t, usedByFinished.ID))

	deploy, err := f.store.GetTask(f.projectID, waitingDeploy.ID)
	require.NoError(t, err)
	assert.Equal(t, &usedByWaiting.ID, deploy.BuildTaskID)
	deploy, err = f.store.GetTask(f.projectID, finishedDeploy.ID)
	require.NoError(t, err)
	assert.Nil(t, deploy.BuildTaskID)
}

// GetIncomingVersion follows a waiting deploy through a finished deploy up to the build, so the
// whole chain stays; once the last deploy of the chain has finished, the build goes.
func testDeleteTemplateTasksBuildChains(t *testing.T, f templateTasksFixture) {
	third := f.template(t, f.projectID, "third")
	build := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	deploy := f.task(t, f.other, task_logger.TaskSuccessStatus, func(task *db.Task) { task.BuildTaskID = &build.ID })
	waiting := f.task(t, third, task_logger.TaskWaitingStatus, func(task *db.Task) { task.BuildTaskID = &deploy.ID })

	for _, tpl := range []db.Template{f.tpl, f.other} {
		deleted, err := f.store.DeleteTemplateTasks(f.projectID, tpl.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, deleted, tpl.Name)
	}
	assert.Equal(t, kept(), f.bound(t, build.ID))
	assert.Equal(t, kept(), f.bound(t, deploy.ID))

	got, err := f.store.GetTask(f.projectID, deploy.ID)
	require.NoError(t, err)
	assert.Equal(t, &build.ID, got.BuildTaskID)

	_, err = f.store.exec("update task set status=? where id=?", task_logger.TaskSuccessStatus, waiting.ID)
	require.NoError(t, err)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, gone(), f.bound(t, build.ID))
}

// The next build takes its version from the newest task of a build template, so that task stays.
func testDeleteTemplateTasksKeepsTheNewestBuild(t *testing.T, f templateTasksFixture) {
	build := f.tpl
	build.Type = db.TemplateBuild
	build.StartVersion = new("1.0.0")
	require.NoError(t, f.store.UpdateTemplate(build))

	versioned := func(version string) func(*db.Task) {
		return func(task *db.Task) { task.Version = new(version) }
	}
	first := f.task(t, build, task_logger.TaskSuccessStatus, versioned("1.0.0"))
	second := f.task(t, build, task_logger.TaskSuccessStatus, versioned("1.0.1"))
	newest := f.task(t, build, task_logger.TaskFailStatus, versioned("1.0.2"))
	plain := f.task(t, f.other, task_logger.TaskSuccessStatus, nil)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, build.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, deleted)
	assert.Equal(t, gone(), f.bound(t, first.ID))
	assert.Equal(t, gone(), f.bound(t, second.ID))
	assert.Equal(t, kept(), f.bound(t, newest.ID))
	assert.Equal(t, 1, f.counter(t, build))

	// Read the way TaskPool.AddTaskFrom numbers the next build.
	builds, err := f.store.GetTemplateTasks(f.projectID, build.ID, db.RetrieveQueryParams{Count: 1})
	require.NoError(t, err)
	require.Len(t, builds, 1)
	assert.Equal(t, "1.0.3", db.GetNextBuildVersion(*build.StartVersion, *builds[0].Version))

	// Any other template loses its whole history.
	deleted, err = f.store.DeleteTemplateTasks(f.projectID, f.other.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, gone(), f.bound(t, plain.ID))
}

// A deploy started while the history is being deleted is not in the chains read up front; the
// batch transaction sees it and keeps its build.
func testDeleteTemplateTasksKeepsBuildsOfTasksCreatedMeanwhile(t *testing.T, f templateTasksFixture) {
	build := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	other := f.task(t, f.tpl, task_logger.TaskSuccessStatus, nil)
	lastTaskID := f.count(t, "select max(id) from task")
	deploy := f.task(t, f.other, task_logger.TaskWaitingStatus, func(task *db.Task) { task.BuildTaskID = &build.ID })

	deleted, err := f.store.deleteTasksBatch(f.projectID, f.tpl.ID, []int{build.ID, other.ID}, map[int]struct{}{}, lastTaskID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, kept(), f.bound(t, build.ID))
	assert.Equal(t, gone(), f.bound(t, other.ID))
	got, err := f.store.GetTask(f.projectID, deploy.ID)
	require.NoError(t, err)
	assert.Equal(t, &build.ID, got.BuildTaskID)
}

func testDeleteTemplateTasksManyBatches(t *testing.T, f templateTasksFixture) {
	total := 2*deleteTemplateTasksBatch + 7
	for i := 0; i < total; i++ {
		_, err := f.store.CreateTask(db.Task{ProjectID: f.projectID, TemplateID: f.tpl.ID, Status: task_logger.TaskSuccessStatus, Created: time.Now()}, 0)
		require.NoError(t, err)
	}
	running := f.task(t, f.tpl, task_logger.TaskRunningStatus, nil)

	deleted, err := f.store.DeleteTemplateTasks(f.projectID, f.tpl.ID)

	require.NoError(t, err)
	assert.Equal(t, total, deleted)
	assert.Equal(t, 1, f.count(t, "select count(*) from task where template_id=?", f.tpl.ID))
	assert.Equal(t, kept(), f.bound(t, running.ID))
	assert.Equal(t, 1, f.counter(t, f.tpl))
}

// deleteTemplateTasksCases are run on SQLite here; each case makes its own project, so they can also
// run one after another on a shared MySQL or PostgreSQL database.
func deleteTemplateTasksCases() map[string]func(*testing.T, templateTasksFixture) {
	return map[string]func(*testing.T, templateTasksFixture){
		"deletes only finished tasks of the template": testDeleteTemplateTasksDeletesOnlyFinishedTasksOfTheTemplate,
		"not found":           testDeleteTemplateTasksNotFound,
		"empty history":       testDeleteTemplateTasksEmptyHistory,
		"repairs the counter": testDeleteTemplateTasksRepairsTheCounter,
		"keeps tasks of unfinished workflow runs": testDeleteTemplateTasksKeepsTasksOfUnfinishedWorkflowRuns,
		"build tasks":            testDeleteTemplateTasksBuildTasks,
		"build chains":           testDeleteTemplateTasksBuildChains,
		"keeps the newest build": testDeleteTemplateTasksKeepsTheNewestBuild,
		"keeps builds of tasks created meanwhile": testDeleteTemplateTasksKeepsBuildsOfTasksCreatedMeanwhile,
		"many batches": testDeleteTemplateTasksManyBatches,
	}
}

func TestDeleteTemplateTasks(t *testing.T) {
	for name, test := range deleteTemplateTasksCases() {
		t.Run(name, func(t *testing.T) {
			test(t, newTemplateTasksFixture(t, InitConfigCreateTestStore()))
		})
	}
}
