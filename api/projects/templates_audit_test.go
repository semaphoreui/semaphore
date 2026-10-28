package projects

import (
	"bytes"
	"errors"
	"fmt"
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
	"github.com/stretchr/testify/require"
)

type fakeTemplateRoles struct {
	db.TemplateManager
	existing  db.TemplateRolePerm
	getErr    error
	deleteErr error
}

func (f *fakeTemplateRoles) CreateTemplateRole(perm db.TemplateRolePerm) (db.TemplateRolePerm, error) {
	perm.ID = 5
	return perm, nil
}
func (f *fakeTemplateRoles) UpdateTemplateRole(db.TemplateRolePerm) error { return nil }
func (f *fakeTemplateRoles) GetTemplateRole(int, int, int) (db.TemplateRolePerm, error) {
	return f.existing, f.getErr
}
func (f *fakeTemplateRoles) DeleteTemplateRole(int, int, int) error { return f.deleteErr }

func templatePermRequest(t *testing.T, store db.Store, method string, body string, permID string) (*http.Request, *audittest.Recorder) {
	t.Helper()
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(method, "/api/project/12/templates/3/perms", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "template", db.Template{ID: 3, ProjectID: 12})
	r = helpers.SetContextValue(r, "audit", rec)
	r = helpers.SetContextValue(r, "store", store)
	if permID != "" {
		r = mux.SetURLVars(r, map[string]string{"perm_id": permID})
	}
	return r, rec
}

func builtinRole(t *testing.T, store db.Store, slug db.ProjectUserRole) db.Role {
	t.Helper()
	role, err := store.GetSharedRoleBySlug(string(slug))
	require.NoError(t, err)
	return role
}

func TestTemplatePermissions_AreRecorded(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	runner := builtinRole(t, store, db.ProjectTaskRunner)
	controller := NewTemplateController(&fakeTemplateRoles{existing: db.TemplateRolePerm{ID: 5, RoleID: runner.ID, RoleSlug: runner.Slug, Permissions: db.CanRunProjectTasks}}, nil)

	r, rec := templatePermRequest(t, store, http.MethodPost, `{"role_slug":"Task_Runner","permissions":1}`, "")
	w := httptest.NewRecorder()
	controller.AddTemplatePerm(w, r)
	got := only(t, rec, audit.IAMTemplatePermissionCreate)
	assert.Equal(t, &audit.Target{Type: audit.TargetTemplatePermission, ID: "5"}, got.Event.Target)
	assert.Equal(t, 12, got.Event.ProjectID)
	ref := audit.NewRoleRef(runner)
	assert.Equal(t, audit.TemplatePermissionMetadata{TemplateID: 3, Role: ref, Permissions: []string{"run_tasks"}}, got.Event.Metadata)
	assert.Contains(t, w.Body.String(), `"role_slug":"task_runner"`)
	assert.NotContains(t, w.Body.String(), "role_id")

	r, rec = templatePermRequest(t, store, http.MethodPut, `{"role_slug":"task_runner","permissions":5}`, "5")
	controller.UpdateTemplatePerm(httptest.NewRecorder(), r)
	assert.Equal(t, audit.TemplatePermissionMetadata{TemplateID: 3, Role: ref, Permissions: []string{"run_tasks", "manage_resources"}},
		only(t, rec, audit.IAMTemplatePermissionUpdate).Event.Metadata)

	r, rec = templatePermRequest(t, store, http.MethodDelete, "", "5")
	controller.DeleteTemplatePerm(httptest.NewRecorder(), r)
	deleted := only(t, rec, audit.IAMTemplatePermissionDelete)
	assert.Equal(t, &audit.Target{Type: audit.TargetTemplatePermission, ID: "5"}, deleted.Event.Target)
	assert.Equal(t, audit.TemplatePermissionMetadata{TemplateID: 3}, deleted.Event.Metadata, "a delete carries identity only")
}

func TestAddTemplatePerm_RejectsUnknownSlug(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	controller := NewTemplateController(&fakeTemplateRoles{}, nil)
	r, rec := templatePermRequest(t, store, http.MethodPost, `{"role_slug":"nope","permissions":1}`, "")
	w := httptest.NewRecorder()

	controller.AddTemplatePerm(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, rec.All())
}

func TestUpdateTemplatePerm_RecordsStoredRole(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	runner := builtinRole(t, store, db.ProjectTaskRunner)
	ref := audit.NewRoleRef(runner)
	tests := []struct {
		name   string
		getErr error
		want   audit.RoleRef
	}{
		{"stored role, not the request body", nil, ref},
		{"failed audit read records no role", errors.New("db down"), audit.RoleRef{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewTemplateController(&fakeTemplateRoles{existing: db.TemplateRolePerm{ID: 5, RoleID: runner.ID, RoleSlug: runner.Slug}, getErr: tt.getErr}, nil)
			r, rec := templatePermRequest(t, store, http.MethodPut, `{"role_slug":"guest","permissions":1}`, "5")
			w := httptest.NewRecorder()

			controller.UpdateTemplatePerm(w, r)

			assert.Equal(t, http.StatusNoContent, w.Code)
			meta := only(t, rec, audit.IAMTemplatePermissionUpdate).Event.Metadata.(audit.TemplatePermissionMetadata)
			assert.Equal(t, tt.want, meta.Role)
		})
	}
}

func TestTemplatePermissions_RejectUnknownBits(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	controller := NewTemplateController(&fakeTemplateRoles{existing: db.TemplateRolePerm{ID: 5, RoleID: 7}}, nil)
	tests := []struct {
		name   string
		method string
		permID string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{"add", http.MethodPost, "", controller.AddTemplatePerm},
		{"update", http.MethodPut, "5", controller.UpdateTemplatePerm},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, body := range []string{`{"role_slug":"guest","permissions":1024}`, `{"role_slug":"guest","permissions":16}`} {
				r, rec := templatePermRequest(t, store, tt.method, body, tt.permID)
				w := httptest.NewRecorder()
				tt.handle(w, r)
				assert.Equal(t, http.StatusBadRequest, w.Code, body)
				assert.Empty(t, rec.All())
			}
		})
	}
}

func TestAddTemplatePerm_IgnoresRoleIDInBody(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	runner := builtinRole(t, store, db.ProjectTaskRunner)
	guest := builtinRole(t, store, db.ProjectGuest)
	controller := NewTemplateController(&fakeTemplateRoles{}, nil)
	r, rec := templatePermRequest(t, store, http.MethodPost, fmt.Sprintf(`{"role_id":%d,"role_slug":"task_runner","permissions":1}`, guest.ID), "")

	controller.AddTemplatePerm(httptest.NewRecorder(), r)

	ref := audit.NewRoleRef(runner)
	meta := only(t, rec, audit.IAMTemplatePermissionCreate).Event.Metadata.(audit.TemplatePermissionMetadata)
	assert.Equal(t, ref, meta.Role)
}
