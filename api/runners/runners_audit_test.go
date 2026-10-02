package runners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerRequest(t *testing.T, store db.Store, token string) (*http.Request, *audittest.Recorder) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"registration_token": token, "name": "r1"})
	require.NoError(t, err)
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(http.MethodPost, "/api/internal/runners", bytes.NewReader(body))
	r = helpers.SetContextValue(r, "store", store)
	return helpers.SetContextValue(r, "audit", rec), rec
}

func TestRegisterRunner_Events(t *testing.T) {
	prev := util.Config
	t.Cleanup(func() { util.Config = prev })
	store := sql.InitConfigCreateTestStore()
	if util.Config.Runners == nil {
		util.Config.Runners = &util.RunnersConfig{}
	}
	util.Config.Runners.RegistrationToken = "global-secret"

	r, rec := registerRequest(t, store, "wrong")
	w := httptest.NewRecorder()
	RegisterRunner(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	failed, err := rec.Only(audit.RunnerLifecycleRegister)
	require.NoError(t, err)
	assert.Equal(t, audit.ReasonInvalidRegistrationToken, failed.Event.Reason)
	assert.Equal(t, audit.AnonymousActor(), failed.Actor)
	assert.Equal(t, audit.RunnerRegisterMetadata{Token: audit.RunnerTokenGlobal}, failed.Event.Metadata)
	assert.NotContains(t, fmt.Sprintf("%+v", failed.Event), "wrong")

	r, rec = registerRequest(t, store, "smrs_wrong")
	RegisterRunner(httptest.NewRecorder(), r)
	failedOneTime, err := rec.Only(audit.RunnerLifecycleRegister)
	require.NoError(t, err)
	assert.Equal(t, audit.ReasonInvalidRegistrationToken, failedOneTime.Event.Reason)
	assert.Equal(t, audit.RunnerRegisterMetadata{Token: audit.RunnerTokenOneTime}, failedOneTime.Event.Metadata)
	assert.NotContains(t, fmt.Sprintf("%+v", failedOneTime.Event), "smrs_wrong")

	r, rec = registerRequest(t, store, "global-secret")
	w = httptest.NewRecorder()
	RegisterRunner(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	registered, err := rec.Only(audit.RunnerLifecycleRegister)
	require.NoError(t, err)
	assert.Equal(t, audit.ActorRunner, registered.Actor.Type)
	assert.Equal(t, audit.RunnerRegisterMetadata{Token: audit.RunnerTokenGlobal}, registered.Event.Metadata)
	assert.NotContains(t, fmt.Sprintf("%+v", registered), "global-secret")

	r, rec = registerRequest(t, store, "")
	RegisterRunner(httptest.NewRecorder(), r)
	assert.Empty(t, rec.All(), "a request without a token is malformed, not an attempt")
}

func TestRunnerMiddleware_PutsTheRunnerIntoTheAuditContext(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPut, "/api/internal/runners", nil)
	r.Header.Set("X-Runner-Token", "runner-token")
	r = helpers.SetContextValue(r, "store", store)

	var actor audit.Actor
	RunnerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor = audit.ActorFrom(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, audit.RunnerActor(runner.ID, "r1"), actor)
}

func TestUpdateRunner_InvalidStatusIsRecorded(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil, nil, nil)
	ctrl := NewRunnerController(nil, &pool, nil, nil)
	runnerID := 1
	tr := tasks.NewTaskRunner(db.Task{ID: 7, ProjectID: 1, RunnerID: &runnerID, Status: task_logger.TaskRunningStatus}, &pool, "", nil)
	pool.StateStore().SetRunning(tr)
	r := newProgressRequest(t, store, db.Runner{ID: runnerID}, runners.RunnerProgress{
		Jobs: []runners.JobProgress{{ID: 7, Status: task_logger.TaskStatus("hacked")}},
	})
	rec := &audittest.Recorder{}
	r = helpers.SetContextValue(r, "audit", rec)
	w := httptest.NewRecorder()

	ctrl.UpdateRunner(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	got, err := rec.Only(audit.RunnerProgressReject)
	require.NoError(t, err)
	assert.Equal(t, audit.ReasonInvalidStatus, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetTask, ID: "7"}, got.Event.Target)
	assert.NotContains(t, fmt.Sprintf("%+v", got.Event), "hacked")
}

func unregisterRequest(store db.Store, runner db.Runner) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(http.MethodDelete, "/api/internal/runners", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "runner", runner)
	r = helpers.SetContextValue(r, "audit", rec)
	r = r.WithContext(audit.WithActor(r.Context(), audit.RunnerActor(runner.ID, runner.Name)))
	return r, rec
}

func TestUnregisterRunner_Events(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)

	r, rec := unregisterRequest(store, runner)
	w := httptest.NewRecorder()
	UnregisterRunner(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	got, err := rec.Only(audit.RunnerLifecycleUnregister)
	require.NoError(t, err)
	assert.Equal(t, audit.RunnerActor(runner.ID, "r1"), got.Actor)
	assert.Equal(t, audit.ResourceTarget(audit.TargetRunner, runner.ID, "r1"), got.Event.Target)

	r, rec = unregisterRequest(store, runner)
	w = httptest.NewRecorder()
	UnregisterRunner(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, rec.All(), "an already removed runner is not an event")
}

func TestUnregisterRunner_EventCarriesProject(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	runner, err := store.CreateRunner(db.Runner{Name: "pr1", Token: "project-runner-token", ProjectID: &project.ID})
	require.NoError(t, err)

	r, rec := unregisterRequest(store, runner)
	w := httptest.NewRecorder()
	UnregisterRunner(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	got, err := rec.Only(audit.RunnerLifecycleUnregister)
	require.NoError(t, err)
	assert.Equal(t, project.ID, got.Event.ProjectID)
}
