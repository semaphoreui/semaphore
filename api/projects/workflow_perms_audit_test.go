package projects

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	sqlstore "github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
)

type fakeWorkflowRoles struct {
	db.WorkflowRoleManager
	existing db.WorkflowRolePerm
	getErr   error
}

func (f *fakeWorkflowRoles) CreateWorkflowRole(perm db.WorkflowRolePerm) (db.WorkflowRolePerm, error) {
	perm.ID = 5
	return perm, nil
}
func (f *fakeWorkflowRoles) UpdateWorkflowRole(db.WorkflowRolePerm) error { return nil }
func (f *fakeWorkflowRoles) GetWorkflowRole(int, int, int) (db.WorkflowRolePerm, error) {
	return f.existing, f.getErr
}
func (f *fakeWorkflowRoles) DeleteWorkflowRole(int, int, int) error { return nil }

func workflowPermRequest(t *testing.T, store db.Store, method string, body string, permID string) (*http.Request, *audittest.Recorder) {
	t.Helper()
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(method, "/api/project/12/workflows/3/perms", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "workflow", db.WorkflowTemplate{ID: 3, ProjectID: 12})
	r = helpers.SetContextValue(r, "audit", rec)
	r = helpers.SetContextValue(r, "store", store)
	if permID != "" {
		r = mux.SetURLVars(r, map[string]string{"perm_id": permID})
	}
	return r, rec
}

func TestWorkflowPermissions_AreRecorded(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	runner := builtinRole(t, store, db.ProjectTaskRunner)
	ref := audit.NewRoleRef(runner)
	controller := NewWorkflowPermController(&fakeWorkflowRoles{existing: db.WorkflowRolePerm{ID: 5, RoleID: runner.ID, RoleSlug: runner.Slug}})

	r, rec := workflowPermRequest(t, store, http.MethodPost, `{"role_slug":"Task_Runner","permissions":16}`, "")
	w := httptest.NewRecorder()
	controller.AddWorkflowPerm(w, r)
	got := only(t, rec, audit.IAMWorkflowPermissionCreate)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, &audit.Target{Type: audit.TargetWorkflowPermission, ID: "5"}, got.Event.Target)
	assert.Equal(t, 12, got.Event.ProjectID)
	assert.Equal(t, audit.WorkflowPermissionMetadata{WorkflowID: 3, Role: ref, Permissions: []string{"run_workflows"}}, got.Event.Metadata)
	assert.Contains(t, w.Body.String(), `"role_slug":"task_runner"`)
	assert.NotContains(t, w.Body.String(), "role_id")

	r, rec = workflowPermRequest(t, store, http.MethodPut, `{"role_slug":"task_runner","permissions":96}`, "5")
	controller.UpdateWorkflowPerm(httptest.NewRecorder(), r)
	assert.Equal(t, audit.WorkflowPermissionMetadata{WorkflowID: 3, Role: ref, Permissions: []string{"edit_workflows", "approve_workflows"}},
		only(t, rec, audit.IAMWorkflowPermissionUpdate).Event.Metadata)

	r, rec = workflowPermRequest(t, store, http.MethodDelete, "", "5")
	controller.DeleteWorkflowPerm(httptest.NewRecorder(), r)
	deleted := only(t, rec, audit.IAMWorkflowPermissionDelete)
	assert.Equal(t, &audit.Target{Type: audit.TargetWorkflowPermission, ID: "5"}, deleted.Event.Target)
	assert.Equal(t, audit.WorkflowPermissionMetadata{WorkflowID: 3}, deleted.Event.Metadata, "a delete carries identity only")
}

func TestUpdateWorkflowPerm_UnknownGrantIsNotFound(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	controller := NewWorkflowPermController(&fakeWorkflowRoles{getErr: db.ErrNotFound})
	r, rec := workflowPermRequest(t, store, http.MethodPut, `{"role_slug":"guest","permissions":16}`, "5")
	w := httptest.NewRecorder()

	controller.UpdateWorkflowPerm(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, rec.All())
}

func TestWorkflowPermissions_RejectUnknownBits(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	controller := NewWorkflowPermController(&fakeWorkflowRoles{existing: db.WorkflowRolePerm{ID: 5, RoleID: 7}})
	tests := []struct {
		name   string
		method string
		permID string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{"add", http.MethodPost, "", controller.AddWorkflowPerm},
		{"update", http.MethodPut, "5", controller.UpdateWorkflowPerm},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, body := range []string{`{"role_slug":"guest","permissions":1024}`, `{"role_slug":"guest","permissions":1}`} {
				r, rec := workflowPermRequest(t, store, tt.method, body, tt.permID)
				w := httptest.NewRecorder()
				tt.handle(w, r)
				assert.Equal(t, http.StatusBadRequest, w.Code, body)
				assert.Empty(t, rec.All())
			}
		})
	}
}
