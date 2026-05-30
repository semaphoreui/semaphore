package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	taskServices "github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClusterTestConfig() {
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
	util.Config.HA = nil
}

// poolWithTasks builds a TaskPool backed by an in-memory store pre-populated
// with one queued and one running task.
func poolWithTasks() *taskServices.TaskPool {
	state := taskServices.NewMemoryTaskStateStore()
	state.Enqueue(&taskServices.TaskRunner{
		Task: db.Task{ID: 1, ProjectID: 10, TemplateID: 100},
	})
	state.SetRunning(&taskServices.TaskRunner{
		Task: db.Task{ID: 2, ProjectID: 10, TemplateID: 100},
	})
	pool := taskServices.CreateTaskPool(nil, state, nil, nil, nil, nil, nil)
	return &pool
}

func TestGetClusterStatus_HADisabled(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/cluster", nil)
	w := httptest.NewRecorder()

	getClusterStatus(w, req)

	var body map[string]any
	var bodyBytes = w.Body.Bytes()
	require.NoError(t, json.Unmarshal(bodyBytes, &body))
	assert.Equal(t, false, body["ha_enabled"])
	// No inspector -> no node / redis sections, but never a 500.
	assert.NotContains(t, body, "nodes")
	assert.NotContains(t, body, "redis")
}

func TestGetClusterTasks_Snapshot(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/cluster/tasks", nil)
	req = helpers.SetContextValue(req, "task_pool", poolWithTasks())
	w := httptest.NewRecorder()

	getClusterTasks(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var snap taskServices.TaskStateSnapshot
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &snap))
	require.Len(t, snap.Queue, 1)
	assert.Equal(t, 1, snap.Queue[0].TaskID)
	require.Len(t, snap.Running, 1)
	assert.Equal(t, 2, snap.Running[0].TaskID)
}

func TestGetClusterTasks_NoPool(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/cluster/tasks", nil)
	w := httptest.NewRecorder()

	getClusterTasks(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var snap taskServices.TaskStateSnapshot
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &snap))
	assert.Empty(t, snap.Queue)
	assert.Empty(t, snap.Running)
}

func TestClearClusterTasks_HonorsScope(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodDelete, "/api/cluster/tasks",
		strings.NewReader(`{"scope":{"queue":true}}`))
	req = helpers.SetContextValue(req, "task_pool", poolWithTasks())
	req = helpers.SetContextValue(req, "user", &db.User{Username: "admin"})
	w := httptest.NewRecorder()

	clearClusterTasks(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var res taskServices.ClearResult
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, 1, res.DeletedKeys)
	assert.Equal(t, 1, res.PerGroup["queue"])
	// Running was not in scope.
	_, hasRunning := res.PerGroup["running"]
	assert.False(t, hasRunning)
}

func TestClearClusterTasks_RejectsEmptyScope(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodDelete, "/api/cluster/tasks",
		strings.NewReader(`{"scope":{}}`))
	req = helpers.SetContextValue(req, "task_pool", poolWithTasks())
	req = helpers.SetContextValue(req, "user", &db.User{Username: "admin"})
	w := httptest.NewRecorder()

	clearClusterTasks(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClearClusterTasks_RejectsMissingBody(t *testing.T) {
	setupClusterTestConfig()

	req := httptest.NewRequest(http.MethodDelete, "/api/cluster/tasks", nil)
	req = helpers.SetContextValue(req, "task_pool", poolWithTasks())
	req = helpers.SetContextValue(req, "user", &db.User{Username: "admin"})
	w := httptest.NewRecorder()

	clearClusterTasks(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
