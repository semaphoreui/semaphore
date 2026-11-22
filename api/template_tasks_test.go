package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Through the real router a member who may not manage the project resources is stopped by the
// route's middleware with 403; one who may, but is not an admin, reaches the handler and gets 401.
func TestDeleteTemplateTasks_ThroughTheRouter(t *testing.T) {
	store := setupSessionTest(t)
	util.Config.Debugging = &util.DebuggingConfig{}
	util.Config.TmpPath = t.TempDir()
	pool := poolWithTasks()
	router := Route(store, fakeGlobalRoleStore{permissions: map[int]db.GlobalPermission{}}, nil, nil, fakeRoleMappingStore{}, nil, pool, nil, nil, nil, nil, nil, nil, nil, noLicense{}, nil, nil, nil, nil)

	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{ProjectID: project.ID, SSHKeyID: key.ID, Name: "r", GitURL: "git@example.com:r", GitBranch: "main"})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{Name: "deploy", Playbook: "site.yml", ProjectID: project.ID, RepositoryID: repo.ID})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: tpl.ID, Status: task_logger.TaskSuccessStatus}, 0)
	require.NoError(t, err)

	member := func(username string, permissions db.ProjectUserPermission) db.User {
		user := createUserOptionsTestUser(t, store, username)
		role, err := store.CreateRole(db.Role{Slug: "history-" + username, Name: username, Permissions: permissions, ProjectID: &project.ID})
		require.NoError(t, err)
		require.NoError(t, store.CreateProjectUser(project.ID, user.ID, []int{role.ID}))
		return user
	}
	reader := member("reader", db.CanReadProject|db.CanRunProjectTasks)
	manager := member("manager", db.CanReadProject|db.CanManageProjectResources)
	root := makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))

	send := func(user db.User) (int, []string) {
		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/project/%d/templates/%d/tasks", project.ID, tpl.ID), nil)
		r = addSessionCookie(t, store, r, user, db.SessionVerificationNone)
		r = helpers.SetContextValue(r, "store", store)
		r = helpers.SetContextValue(r, "task_pool", pool)
		r, rec := withAuditRecorder(r)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		var denied []string
		for _, got := range denials(rec) {
			denied = append(denied, got.Event.Metadata.(audit.DenyMetadata).Permission)
		}
		return w.Code, denied
	}

	code, denied := send(reader)
	assert.Equal(t, http.StatusForbidden, code)
	assert.Equal(t, []string{"manage_resources"}, denied)

	code, denied = send(manager)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.Equal(t, []string{"admin"}, denied)

	_, err = store.GetTask(project.ID, task.ID)
	require.NoError(t, err, "a refused request deletes nothing")

	code, denied = send(root)
	assert.Equal(t, http.StatusNoContent, code)
	assert.Empty(t, denied)
	_, err = store.GetTask(project.ID, task.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}
