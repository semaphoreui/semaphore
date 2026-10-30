package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanRunTemplate(t *testing.T) {
	run := db.CanRunProjectTasks
	manage := db.CanManageProjectResources
	tests := []struct {
		name        string
		admin       bool
		permissions db.ProjectUserPermission
		grant       *db.ProjectUserPermission
		expected    bool
	}{
		{"admin", true, 0, nil, true},
		{"project role runs", false, run, nil, true},
		{"template grant runs", false, 0, &run, true},
		{"template grant without run", false, 0, &manage, false},
		{"no permission", false, 0, nil, false},
		{"a grant does not take the project role away", false, run, &manage, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := db.TemplateWithPerms{Permissions: tt.grant}
			assert.Equal(t, tt.expected, canRunTemplate(&db.User{Admin: tt.admin}, tt.permissions, tpl))
		})
	}
}

// missingInventories stops a task at preparation, after its row is created and
// before it reaches the pool queue, which no test runs.
type missingInventories struct{}

func (missingInventories) GetInventory(int, int) (db.Inventory, error) {
	return db.Inventory{}, db.ErrNotFound
}

func TestRunViewTasks(t *testing.T) {
	f := newTaskControlFixture(t)
	view, err := f.store.CreateView(db.View{ProjectID: f.project.ID, Title: "v"})
	require.NoError(t, err)
	inView, err := f.store.CreateTemplate(db.Template{Name: "in view", Playbook: "site.yml", ProjectID: f.project.ID, RepositoryID: f.template.RepositoryID, InventoryID: f.template.InventoryID, ViewID: &view.ID})
	require.NoError(t, err)
	user, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "u", Name: "u", Email: "u@example.com"}})
	require.NoError(t, err)
	encryption := server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store)
	pool := tasks.CreateTaskPool(f.store, tasks.NewMemoryTaskStateStore(), nil, missingInventories{}, encryption, nil, nopLogWriter{}, nil, nil)

	call := func(permissions db.ProjectUserPermission) (*httptest.ResponseRecorder, *audittest.Recorder) {
		rec := &audittest.Recorder{}
		r := httptest.NewRequest(http.MethodPost, "/api/project/1/views/1/tasks", nil)
		r = helpers.SetContextValue(r, "store", f.store)
		r = helpers.SetContextValue(r, "task_pool", &pool)
		r = helpers.SetContextValue(r, "project", f.project)
		r = helpers.SetContextValue(r, "view", view)
		r = helpers.SetContextValue(r, "user", &user)
		r = helpers.SetContextValue(r, "permissions", permissions)
		r = helpers.SetContextValue(r, "audit", rec)
		w := httptest.NewRecorder()
		RunViewTasks(w, r)
		return w, rec
	}
	tasksOf := func(tpl db.Template) []db.TaskWithTpl {
		list, err := f.store.GetTemplateTasks(f.project.ID, tpl.ID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		return list
	}

	t.Run("refuses a user who may run none of the templates", func(t *testing.T) {
		w, rec := call(0)

		assert.Equal(t, http.StatusForbidden, w.Code)
		require.Len(t, rec.All(), 1)
		assert.Equal(t, audit.AuthAuthorizationDeny, rec.All()[0].Event.Kind)
		assert.Empty(t, tasksOf(inView))
	})

	t.Run("starts only the templates of the view and reports the failed ones", func(t *testing.T) {
		w, _ := call(db.CanRunProjectTasks)

		require.Equal(t, http.StatusCreated, w.Code)
		var res viewTasks
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.Empty(t, res.Tasks)
		assert.Equal(t, []int{inView.ID}, res.FailedTemplateIDs)
		assert.Len(t, tasksOf(inView), 1)
		assert.Empty(t, tasksOf(f.template))
	})
}
