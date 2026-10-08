package runners

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRunner_InvalidTokenReturnsBadRequest(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	body, err := json.Marshal(map[string]any{
		"registration_token": "not-a-valid-token",
		"name":               "test-runner",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/runners", bytes.NewReader(body))
	req = helpers.SetContextValue(req, "store", store)

	w := httptest.NewRecorder()
	RegisterRunner(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, "Invalid registration token", res["error"])
}

func newProgressRequest(t *testing.T, store db.Store, runner db.Runner, progress runners.RunnerProgress) *http.Request {
	t.Helper()

	body, err := json.Marshal(progress)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/internal/runners", bytes.NewReader(body))
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "runner", runner)
	return req
}

func decodeProgressResponse(t *testing.T, w *httptest.ResponseRecorder) runners.RunnerProgressResponse {
	t.Helper()

	var res runners.RunnerProgressResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	return res
}

func TestUpdateRunner_StoppedTaskReportedAsTerminated(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore() // also initializes util.Config

	pool := tasks.CreateTaskPool(
		store,
		tasks.NewMemoryTaskStateStore(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	ctrl := NewRunnerController(nil, &pool, nil, nil)

	runnerID := 1

	// The task was stopped on the server while the runner was offline.
	tr := tasks.NewTaskRunner(db.Task{
		ID:        5,
		ProjectID: 1,
		RunnerID:  &runnerID,
		Status:    task_logger.TaskStoppedStatus,
	}, &pool, "", nil)
	pool.StateStore().SetRunning(tr)

	req := newProgressRequest(t, store, db.Runner{ID: runnerID}, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:     5,
			Status: task_logger.TaskRunningStatus,
			LogRecords: []runners.LogRecord{
				{Time: tz.Now(), Message: "late output"},
			},
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []int{5}, decodeProgressResponse(t, w).TerminatedJobs)

	// The late report must not overwrite the terminal status.
	assert.Equal(t, task_logger.TaskStoppedStatus, tr.Task.Status)
}

func TestUpdateRunner_UnknownTaskReportedAsTerminated(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()

	pool := tasks.CreateTaskPool(
		store,
		tasks.NewMemoryTaskStateStore(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	ctrl := NewRunnerController(nil, &pool, nil, nil)

	req := newProgressRequest(t, store, db.Runner{ID: 1}, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:     999, // neither in the pool nor in the database
			Status: task_logger.TaskRunningStatus,
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []int{999}, decodeProgressResponse(t, w).TerminatedJobs)
}

func TestUpdateRunner_ReassignedTaskReportedAsTerminated(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()

	pool := tasks.CreateTaskPool(
		store,
		tasks.NewMemoryTaskStateStore(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	ctrl := NewRunnerController(nil, &pool, nil, nil)

	oldRunnerID := 1
	newRunnerID := 2

	// Task was reassigned from runner 1 to runner 2 while runner 1 still had
	// the job in its local pool (e.g. after requeueTaskRunnerOffline).
	tr := tasks.NewTaskRunner(db.Task{
		ID:        8,
		ProjectID: 1,
		RunnerID:  &newRunnerID,
		Status:    task_logger.TaskStartingStatus,
	}, &pool, "", nil)
	pool.StateStore().SetRunning(tr)

	req := newProgressRequest(t, store, db.Runner{ID: oldRunnerID}, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:     8,
			Status: task_logger.TaskRunningStatus,
			LogRecords: []runners.LogRecord{
				{Time: tz.Now(), Message: "stale runner output"},
			},
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []int{8}, decodeProgressResponse(t, w).TerminatedJobs)

	// The late report must not overwrite status or assignee.
	assert.Equal(t, task_logger.TaskStartingStatus, tr.Task.Status)
	assert.Equal(t, newRunnerID, *tr.Task.RunnerID)
}

func TestUpdateRunner_RunningTaskAcceptedWithoutTermination(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()

	pool := tasks.CreateTaskPool(
		store,
		tasks.NewMemoryTaskStateStore(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	ctrl := NewRunnerController(nil, &pool, nil, nil)

	runnerID := 1

	tr := tasks.NewTaskRunner(db.Task{
		ID:        7,
		ProjectID: 1,
		RunnerID:  &runnerID,
		Status:    task_logger.TaskRunningStatus,
	}, &pool, "", nil)
	pool.StateStore().SetRunning(tr)

	req := newProgressRequest(t, store, db.Runner{ID: runnerID}, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:     7,
			Status: task_logger.TaskRunningStatus,
			LogRecords: []runners.LogRecord{
				{Time: tz.Now(), Message: "normal output"},
			},
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, decodeProgressResponse(t, w).TerminatedJobs)
	assert.Equal(t, task_logger.TaskRunningStatus, tr.Task.Status)
}

func TestRegisterRunner_NonSmrsTokenWithoutGlobalMatchReturnsBadRequest(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{
		RunnerRegistrationToken: "global-reg-token",
	}

	store := sql.InitConfigCreateTestStore()

	body, err := json.Marshal(map[string]any{
		"registration_token": "legacy-one-time-token-without-smrs-prefix",
		"name":               "test-runner",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/runners", bytes.NewReader(body))
	req = helpers.SetContextValue(req, "store", store)

	w := httptest.NewRecorder()
	RegisterRunner(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnregisterRunner_AlreadyRemovedRunnerIsNoContent(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r"})
	require.NoError(t, err)
	require.NoError(t, store.DeleteGlobalRunner(runner.ID))

	req := httptest.NewRequest(http.MethodDelete, "/api/internal/runners", nil)
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "runner", runner)
	w := httptest.NewRecorder()
	UnregisterRunner(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// outputsCollectorFake stands in for the pro outputs collector on the server:
// it records what the handler asked it to validate and answers as told.
type outputsCollectorFake struct {
	err       error
	validated []string
}

func (c *outputsCollectorFake) Begin(string, int) (pro_interfaces.TaskOutputsCapture, error) {
	return nil, nil
}

func (c *outputsCollectorFake) Validate(document string) (*string, error) {
	c.validated = append(c.validated, document)
	if c.err != nil {
		return nil, c.err
	}
	return &document, nil
}

type nopLogWriteService struct{}

func (nopLogWriteService) WriteEventLog(pro_interfaces.EventLogRecord) error { return nil }
func (nopLogWriteService) WriteTaskLog(pro_interfaces.TaskLogRecord) error   { return nil }
func (nopLogWriteService) WriteResult(any) error                             { return nil }

// remoteWorkflowTaskFixture is a workflow task assigned to a remote runner,
// persisted so terminal statuses and outputs can be written to the store.
type remoteWorkflowTaskFixture struct {
	store     db.Store
	pool      *tasks.TaskPool
	collector *outputsCollectorFake
	runner    db.Runner
	task      *tasks.TaskRunner
}

func newRemoteWorkflowTaskFixture(t *testing.T) remoteWorkflowTaskFixture {
	t.Helper()

	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()

	// Finalizing a terminal status writes the task event through the log
	// write service, so the fixture needs one.
	pool := tasks.CreateTaskPool(
		store,
		tasks.NewMemoryTaskStateStore(),
		nil,
		nil,
		nil,
		nil,
		nopLogWriteService{},
		nil,
		nil,
	)
	collector := &outputsCollectorFake{}
	pool.SetTaskOutputsCollector(collector)

	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)
	inv, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Test",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &inv.ID,
	})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: proj.ID, TemplateID: tpl.ID}, 0)
	require.NoError(t, err)

	// UpdateTask persists runner_id, which references the runner table.
	runner, err := store.CreateRunner(db.Runner{Name: "remote", Token: "test-token", Active: true})
	require.NoError(t, err)
	runID := 11
	task.RunnerID = &runner.ID
	task.WorkflowRunID = &runID
	task.Status = task_logger.TaskRunningStatus

	tr := tasks.NewTaskRunner(task, &pool, "", nil)
	tr.Template = tpl
	pool.StateStore().SetRunning(tr)

	return remoteWorkflowTaskFixture{
		store:     store,
		pool:      &pool,
		collector: collector,
		runner:    runner,
		task:      tr,
	}
}

func TestUpdateRunner_OutputsIgnoredWhileRunning(t *testing.T) {
	fixture := newRemoteWorkflowTaskFixture(t)
	ctrl := NewRunnerController(nil, fixture.pool, nil, nil)
	document := `{"values":{"image_tag":"1.4.2"}}`

	req := newProgressRequest(t, fixture.store, fixture.runner, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:      fixture.task.Task.ID,
			Status:  task_logger.TaskRunningStatus,
			Outputs: &document,
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, fixture.collector.validated)
	assert.Nil(t, fixture.task.Task.Artifacts)
	assert.Equal(t, task_logger.TaskRunningStatus, fixture.task.Task.Status)
}

// A workflow run may be progressed by any HA node as soon as the terminal
// status is in the database, so the outputs must be stored before it.
func TestUpdateRunner_SuccessStoresOutputsBeforeStatus(t *testing.T) {
	fixture := newRemoteWorkflowTaskFixture(t)
	ctrl := NewRunnerController(nil, fixture.pool, nil, nil)
	document := `{"values":{"image_tag":"1.4.2"}}`

	var storedAtSuccess *string
	fixture.task.AddStatusListener(func(status task_logger.TaskStatus) {
		if status != task_logger.TaskSuccessStatus {
			return
		}
		persisted, err := fixture.store.GetTaskByID(fixture.task.Task.ID)
		require.NoError(t, err)
		storedAtSuccess = persisted.Artifacts
	})

	req := newProgressRequest(t, fixture.store, fixture.runner, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:      fixture.task.Task.ID,
			Status:  task_logger.TaskSuccessStatus,
			Outputs: &document,
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, decodeProgressResponse(t, w).TerminatedJobs)
	assert.Equal(t, []string{document}, fixture.collector.validated)
	assert.Equal(t, task_logger.TaskSuccessStatus, fixture.task.Task.Status)
	require.NotNil(t, storedAtSuccess)
	assert.Equal(t, document, *storedAtSuccess)
	require.NotNil(t, fixture.task.Task.Artifacts)
	assert.Equal(t, document, *fixture.task.Task.Artifacts)
}

func TestUpdateRunner_InvalidOutputsFailTheTask(t *testing.T) {
	fixture := newRemoteWorkflowTaskFixture(t)
	fixture.collector.err = errors.New("output name \"bad name\" is not valid")
	ctrl := NewRunnerController(nil, fixture.pool, nil, nil)
	document := `{"values":{"bad name":1}}`

	req := newProgressRequest(t, fixture.store, fixture.runner, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:      fixture.task.Task.ID,
			Status:  task_logger.TaskSuccessStatus,
			Outputs: &document,
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, task_logger.TaskFailStatus, fixture.task.Task.Status)
	assert.Nil(t, fixture.task.Task.Artifacts)
	persisted, err := fixture.store.GetTaskByID(fixture.task.Task.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, persisted.Status)
	assert.Nil(t, persisted.Artifacts)
}

func TestUpdateRunner_SuccessWithoutOutputs(t *testing.T) {
	fixture := newRemoteWorkflowTaskFixture(t)
	ctrl := NewRunnerController(nil, fixture.pool, nil, nil)

	req := newProgressRequest(t, fixture.store, fixture.runner, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{
			ID:     fixture.task.Task.ID,
			Status: task_logger.TaskSuccessStatus,
		}},
	})
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, fixture.collector.validated)
	assert.Equal(t, task_logger.TaskSuccessStatus, fixture.task.Task.Status)
	assert.Nil(t, fixture.task.Task.Artifacts)
}
