package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type taskControlFixture struct {
	store    *sql.SqlDb
	pool     *tasks.TaskPool
	project  db.Project
	template db.Template
}

func newTaskControlFixture(t *testing.T) taskControlFixture {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{ProjectID: project.ID, SSHKeyID: key.ID, Name: "r", GitURL: "git@example.com:r", GitBranch: "main"})
	require.NoError(t, err)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: project.ID})
	require.NoError(t, err)
	template, err := store.CreateTemplate(db.Template{Name: "deploy", Playbook: "site.yml", ProjectID: project.ID, RepositoryID: repo.ID, InventoryID: &inventory.ID})
	require.NoError(t, err)
	encryption := server.NewAccessKeyEncryptionService(store, store, store, store)
	inventories := server.NewInventoryService(store, store, store, encryption)
	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, inventories, encryption, nil, nopLogWriter{}, nil, nil)
	return taskControlFixture{store: store, pool: &pool, project: project, template: template}
}

// task creates a task row and, when active, puts it into the pool like a queued or running task.
func (f taskControlFixture) task(t *testing.T, status task_logger.TaskStatus, active bool) db.Task {
	t.Helper()
	task, err := f.store.CreateTask(db.Task{ProjectID: f.project.ID, TemplateID: f.template.ID, Status: status}, 0)
	require.NoError(t, err)
	if active {
		f.pool.StateStore().SetRunning(tasks.NewTaskRunner(task, f.pool, "", nil))
	}
	return task
}

func (f taskControlFixture) request(body string, user db.User, task db.Task) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(http.MethodPost, "/api/project/1/tasks/1/x", strings.NewReader(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "task_pool", f.pool)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "template", f.template)
	r = helpers.SetContextValue(r, "task", task)
	r = helpers.SetContextValue(r, "user", &user)
	r = helpers.SetContextValue(r, "audit", rec)
	return r.WithContext(audit.WithActor(r.Context(), audit.UserActor(user.ID, user.Username, audit.AuthSession, ""))), rec
}

func TestTaskControlEvents(t *testing.T) {
	admin := db.User{ID: 1, Username: "admin", Admin: true}
	tests := []struct {
		name   string
		status task_logger.TaskStatus
		call   func(c *TaskController, w http.ResponseWriter, r *http.Request)
		body   string
		kind   audit.Kind
	}{
		{"approve", task_logger.TaskWaitingConfirmation, (*TaskController).ConfirmTask, "", audit.TaskApprovalApprove},
		{"reject", task_logger.TaskWaitingConfirmation, (*TaskController).RejectTask, "", audit.TaskApprovalReject},
		{"stop", task_logger.TaskStartingStatus, (*TaskController).StopTask, `{"force":false}`, audit.TaskControlStop},
		{"force stop", task_logger.TaskStartingStatus, (*TaskController).StopTask, `{"force":true}`, audit.TaskControlForceStop},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTaskControlFixture(t)
			task := f.task(t, tt.status, true)
			r, rec := f.request(tt.body, admin, task)

			tt.call(NewTaskController(f.store, nil), httptest.NewRecorder(), r)

			got, err := rec.Only(tt.kind)
			require.NoError(t, err)
			assert.Equal(t, &audit.Target{Type: audit.TargetTask, ID: strconv.Itoa(task.ID)}, got.Event.Target)
			assert.Equal(t, audit.TaskMetadata{TemplateID: f.template.ID}, got.Event.Metadata)
			assert.Equal(t, f.project.ID, got.Event.ProjectID)
		})
	}
}

func TestTaskControl_UnchangedStatusIsNotRecorded(t *testing.T) {
	admin := db.User{ID: 1, Username: "admin", Admin: true}
	tests := []struct {
		name   string
		status task_logger.TaskStatus
		active bool
		call   func(c *TaskController, w http.ResponseWriter, r *http.Request)
		body   string
	}{
		{"approve a task that does not wait", task_logger.TaskStartingStatus, true, (*TaskController).ConfirmTask, ""},
		{"reject a task that does not wait", task_logger.TaskStartingStatus, true, (*TaskController).RejectTask, ""},
		{"stop a finished task", task_logger.TaskSuccessStatus, false, (*TaskController).StopTask, `{"force":false}`},
		// In HA every existing task looks active.
		{"stop a finished task that looks active", task_logger.TaskSuccessStatus, true, (*TaskController).StopTask, `{"force":false}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTaskControlFixture(t)
			task := f.task(t, tt.status, tt.active)
			r, rec := f.request(tt.body, admin, task)
			w := httptest.NewRecorder()

			tt.call(NewTaskController(f.store, nil), w, r)

			assert.Equal(t, http.StatusNoContent, w.Code, "the API answer is unchanged")
			assert.Empty(t, rec.All())
		})
	}
}

func TestStopTask_InactiveUnfinishedTaskCompletes(t *testing.T) {
	f := newTaskControlFixture(t)
	task := f.task(t, task_logger.TaskWaitingStatus, false)
	poolRec := &audittest.Recorder{}
	f.pool.SetAuditRecorder(poolRec)
	r, rec := f.request(`{"force":false}`, db.User{ID: 1, Username: "admin", Admin: true}, task)

	NewTaskController(f.store, nil).StopTask(httptest.NewRecorder(), r)

	_, err := rec.Only(audit.TaskControlStop)
	require.NoError(t, err)
	complete, err := poolRec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err)
	assert.Equal(t, "stopped", complete.Event.Metadata.(audit.TaskCompleteMetadata).Result)
}

func TestConfirmTask_InactiveTaskIsNotRecorded(t *testing.T) {
	f := newTaskControlFixture(t)
	task := f.task(t, task_logger.TaskWaitingConfirmation, false)
	r, rec := f.request("", db.User{ID: 1, Username: "admin", Admin: true}, task)

	NewTaskController(f.store, nil).ConfirmTask(httptest.NewRecorder(), r)

	assert.Empty(t, rec.All())
}

func TestStopAllTasks_IsOneEventForTheTemplate(t *testing.T) {
	f := newTaskControlFixture(t)
	r, rec := f.request(`{}`, db.User{ID: 1, Username: "admin", Admin: true}, db.Task{})

	NewTaskController(f.store, nil).StopAllTasks(httptest.NewRecorder(), r)

	got, err := rec.Only(audit.TaskControlStopAll)
	require.NoError(t, err)
	assert.Equal(t, &audit.Target{Type: audit.TargetTemplate, ID: strconv.Itoa(f.template.ID), Name: "deploy"}, got.Event.Target)
}

func TestRemoveTask_Events(t *testing.T) {
	f := newTaskControlFixture(t)
	task := f.task(t, task_logger.TaskSuccessStatus, false)

	r, rec := f.request("", db.User{ID: 2, Username: "bob"}, task)
	w := httptest.NewRecorder()
	NewTaskController(f.store, nil).RemoveTask(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	denied, err := rec.Only(audit.AuthAuthorizationDeny)
	require.NoError(t, err)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: "admin"}, denied.Event.Metadata)

	r, rec = f.request("", db.User{ID: 1, Username: "admin", Admin: true}, task)
	NewTaskController(f.store, nil).RemoveTask(httptest.NewRecorder(), r)
	deleted, err := rec.Only(audit.TaskHistoryDelete)
	require.NoError(t, err)
	assert.Equal(t, strconv.Itoa(task.ID), deleted.Event.Target.ID)

	// A second delete finds no task.
	r, rec = f.request("", db.User{ID: 1, Username: "admin", Admin: true}, task)
	w = httptest.NewRecorder()
	NewTaskController(f.store, nil).RemoveTask(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, rec.All())
}
