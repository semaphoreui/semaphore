package runners

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRunner_InvalidTokenReturnsBadRequest(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{
		RunnerRegistrationToken: "global-reg-token",
	}

	store := sql.CreateTestStore()

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

	store := sql.CreateTestStore() // also initializes util.Config

	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil)
	ctrl := NewRunnerController(nil, &pool, nil)

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

	store := sql.CreateTestStore()

	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil)
	ctrl := NewRunnerController(nil, &pool, nil)

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

	store := sql.CreateTestStore()

	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil)
	ctrl := NewRunnerController(nil, &pool, nil)

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

	store := sql.CreateTestStore()

	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil)
	ctrl := NewRunnerController(nil, &pool, nil)

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

	store := sql.CreateTestStore()

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
