package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunnerService struct {
	server.RunnerService
}

func (fakeRunnerService) CreateRunner(runner db.Runner) (db.Runner, error) {
	runner.ID = 5
	return runner, nil
}

func (fakeRunnerService) RegenerateRegistrationToken(db.Runner) (string, error) {
	return "smrs_x", nil
}

func TestGlobalRunnerEvents(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1"})
	require.NoError(t, err)
	controller := NewGlobalRunnerController(fakeRunnerService{})
	admin := db.User{ID: 1, Username: "admin", Admin: true}

	tests := []struct {
		name       string
		handler    func(w http.ResponseWriter, r *http.Request)
		body       string
		kind       audit.Kind
		targetName string
		wasActive  bool
	}{
		{"create", controller.AddRunner, `{"name":"r2"}`, audit.RunnerLifecycleCreate, "r2", false},
		{"update", controller.UpdateRunner, `{"name":"r1b"}`, audit.RunnerLifecycleUpdate, "r1b", false},
		{"enable", controller.SetRunnerActive, `{"active":true}`, audit.RunnerLifecycleEnable, "r1", false},
		{"disable", controller.SetRunnerActive, `{"active":false}`, audit.RunnerLifecycleDisable, "r1", true},
		{"rotate", controller.RegenerateRegistrationToken, "", audit.RunnerCredentialRotate, "r1", false},
		{"clear cache", controller.ClearRunnerCache, "", audit.RunnerCacheClear, "r1", false},
		{"delete", controller.DeleteRunner, "", audit.RunnerLifecycleDelete, "r1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := runner
			current.Active = tt.wasActive
			r := httptest.NewRequest(http.MethodPost, "/api/runners/1", strings.NewReader(tt.body))
			r = helpers.SetContextValue(r, "store", store)
			r = helpers.SetContextValue(r, "runner", &current)
			r, rec := withAuditRecorder(asActor(r, admin))

			tt.handler(httptest.NewRecorder(), r)

			got := onlyEvent(t, rec, tt.kind)
			assert.Equal(t, audit.TargetRunner, got.Event.Target.Type)
			assert.Equal(t, tt.targetName, got.Event.Target.Name)
			assert.Zero(t, got.Event.ProjectID, "global runners have no project")
			assert.NotContains(t, fmt.Sprintf("%+v", got.Event), "smrs_x")
		})
	}
}

func TestGlobalRunnerEnableUnchangedRecordsNothing(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Active: true})
	require.NoError(t, err)
	runner.Active = true
	controller := NewGlobalRunnerController(fakeRunnerService{})

	r := httptest.NewRequest(http.MethodPost, "/api/runners/1", strings.NewReader(`{"active":true}`))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "runner", &runner)
	r, rec := withAuditRecorder(asActor(r, db.User{ID: 1, Username: "admin", Admin: true}))
	w := httptest.NewRecorder()

	controller.SetRunnerActive(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, rec.All())
}
