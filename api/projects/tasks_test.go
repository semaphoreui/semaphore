package projects

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTasksPageParams(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		expectedPageSize int
		expectedCount    int // params.Count == pageSize + 1
		expectedBeforeID int
	}{
		{"defaults", "", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"count and before", "count=20&before=100", 20, 21, 100},
		{"legacy limit", "limit=50", 50, 51, 0},
		{"count overrides limit", "count=10&limit=50", 10, 11, 0},
		{"page size capped at max", "count=10000", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"negative count ignored", "count=-5", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"zero count ignored", "count=0", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"invalid count ignored", "count=abc", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"negative before ignored", "count=20&before=-1", 20, 21, 0},
		{"invalid before ignored", "count=20&before=xyz", 20, 21, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := url.ParseQuery(tt.query)
			assert.NoError(t, err)

			params, pageSize := parseTasksPageParams(query, db.RetrieveQueryParams{})

			assert.Equal(t, tt.expectedPageSize, pageSize)
			assert.Equal(t, tt.expectedCount, params.Count)
			assert.Equal(t, tt.expectedBeforeID, params.BeforeID)
		})
	}
}

func TestParseTasksPageParams_PreservesBase(t *testing.T) {
	base := db.RetrieveQueryParams{SortBy: "id", SortInverted: true}

	params, pageSize := parseTasksPageParams(url.Values{}, base)

	assert.Equal(t, "id", params.SortBy)
	assert.True(t, params.SortInverted)
	assert.Equal(t, maxTasksPageSize, pageSize)
	assert.Equal(t, maxTasksPageSize+1, params.Count)
}

// createTaskTestTemplate creates a template usable by the task tests. Templates
// are not unique by name, so the name is a parameter to cover ambiguity.
func createTaskTestTemplate(t *testing.T, store db.Store, projectID int, repositoryID int, name string) db.Template {
	t.Helper()

	tpl, err := store.CreateTemplate(db.Template{
		Name:         name,
		Playbook:     "test.yml",
		ProjectID:    projectID,
		RepositoryID: repositoryID,
	})
	require.NoError(t, err)

	return tpl
}

func TestResolveTaskTemplate(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "task template resolution"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "none",
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		SSHKeyID:  key.ID,
		Name:      "repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	build := createTaskTestTemplate(t, store, project.ID, repo.ID, "Build website")

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	c := &TaskController{store: store}

	t.Run("resolves by id", func(t *testing.T) {
		task := db.Task{TemplateID: build.ID}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
	})

	t.Run("resolves by name and writes the id back", func(t *testing.T) {
		task := db.Task{TemplateName: "Build website"}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
		assert.Equal(t, build.ID, task.TemplateID, "the resolved id must be written back to the task")
		assert.Empty(t, task.TemplateName, "the name must not survive into the stored task or the response")
	})

	t.Run("id wins when both are given", func(t *testing.T) {
		task := db.Task{TemplateID: build.ID, TemplateName: "does not exist"}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
		assert.Empty(t, task.TemplateName)
	})

	t.Run("neither id nor name is rejected", func(t *testing.T) {
		task := db.Task{}

		_, err := c.resolveTaskTemplate(project.ID, &task)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "template_id or template_name is required")
	})

	t.Run("unknown name is not found", func(t *testing.T) {
		task := db.Task{TemplateName: "no such template"}

		_, err := c.resolveTaskTemplate(project.ID, &task)

		assert.ErrorIs(t, err, db.ErrNotFound)
	})

	t.Run("a template of another project is not found", func(t *testing.T) {
		task := db.Task{TemplateName: "Build website"}

		_, err := c.resolveTaskTemplate(otherProject.ID, &task)

		assert.ErrorIs(t, err, db.ErrNotFound)
	})

	t.Run("an ambiguous name is rejected", func(t *testing.T) {
		// Both the store and the unique index reject a duplicate name, so the
		// collision has to be made behind their backs. This is a database which
		// lost the index, for example one restored from a schema-less dump: the
		// task must still refuse to guess rather than run the wrong template.
		createTaskTestTemplate(t, store, project.ID, repo.ID, "Duplicate")
		legacy := createTaskTestTemplate(t, store, project.ID, repo.ID, "Duplicate (2)")

		_, err = store.Sql().Exec("drop index project__template__project_id_name")
		require.NoError(t, err)

		_, err = store.Sql().Exec(
			"update project__template set name=? where id=?", "Duplicate", legacy.ID)
		require.NoError(t, err)

		task := db.Task{TemplateName: "Duplicate"}

		_, err = c.resolveTaskTemplate(project.ID, &task)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "more than one template")
	})
}

func runTaskMiddlewares(f taskControlFixture, taskID int, user *db.User, permissions db.ProjectUserPermission) (*httptest.ResponseRecorder, db.ProjectUserPermission) {
	r := httptest.NewRequest(http.MethodGet, "/api/project/1/tasks/"+strconv.Itoa(taskID), nil)
	r = mux.SetURLVars(r, map[string]string{"task_id": strconv.Itoa(taskID)})
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", user)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "permissions", permissions)
	w := httptest.NewRecorder()
	var effective db.ProjectUserPermission
	controller := NewTaskController(f.store, nil)
	controller.GetTaskMiddleware(controller.GetTaskPermissionsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		effective = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
		w.WriteHeader(http.StatusNoContent)
	}))).ServeHTTP(w, r)
	return w, effective
}

// A task is readable through a grant on its template or on the workflow of its run; otherwise it is not there.
func TestTaskMiddlewares_ReadThroughGrants(t *testing.T) {
	f := newTaskControlFixture(t)
	role, err := f.store.CreateRole(db.Role{Slug: "nobody", Name: "Nobody", ProjectID: &f.project.ID})
	require.NoError(t, err)
	bob, err := f.store.CreateUserWithoutPassword(db.User{Username: "bob", Name: "Bob", Email: "bob@example.com"})
	require.NoError(t, err)
	require.NoError(t, f.store.CreateProjectUser(f.project.ID, bob.ID, []int{role.ID}))
	task := f.task(t, task_logger.TaskWaitingStatus, false)

	missing, _ := runTaskMiddlewares(f, 9999, &bob, db.CanReadProject)
	assert.Equal(t, http.StatusNotFound, missing.Code)
	hidden, _ := runTaskMiddlewares(f, task.ID, &bob, 0)
	assert.Equal(t, http.StatusNotFound, hidden.Code)
	assert.Equal(t, missing.Body.String(), hidden.Body.String())

	_, err = f.store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: f.project.ID, TemplateID: f.template.ID, RoleID: role.ID, Permissions: db.CanReadProject | db.CanRunProjectTasks})
	require.NoError(t, err)
	w, effective := runTaskMiddlewares(f, task.ID, &bob, 0)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, db.CanReadProject|db.CanRunProjectTasks, effective)

	_, err = f.store.Sql().Exec(f.store.PrepareQuery("insert into project__workflow_template (project_id, name) values (?, ?)"), f.project.ID, "release")
	require.NoError(t, err)
	workflowID, err := f.store.Sql().SelectInt(f.store.PrepareQuery("select id from project__workflow_template where project_id = ?"), f.project.ID)
	require.NoError(t, err)
	_, err = f.store.Sql().Exec(f.store.PrepareQuery("insert into project__workflow_run (project_id, workflow_template_id, status) values (?, ?, ?)"), f.project.ID, workflowID, "running")
	require.NoError(t, err)
	runID, err := f.store.Sql().SelectInt(f.store.PrepareQuery("select id from project__workflow_run where project_id = ?"), f.project.ID)
	require.NoError(t, err)
	other, err := f.store.CreateTemplate(db.Template{Name: "node", Playbook: "node.yml", ProjectID: f.project.ID, RepositoryID: f.template.RepositoryID, InventoryID: f.template.InventoryID})
	require.NoError(t, err)
	id := int(runID)
	nodeTask, err := f.store.CreateTask(db.Task{ProjectID: f.project.ID, TemplateID: other.ID, WorkflowRunID: &id}, 0)
	require.NoError(t, err)
	hidden, _ = runTaskMiddlewares(f, nodeTask.ID, &bob, 0)
	assert.Equal(t, http.StatusNotFound, hidden.Code)
	_, err = f.store.CreateWorkflowRole(db.WorkflowRolePerm{ProjectID: f.project.ID, WorkflowID: int(workflowID), RoleID: role.ID, Permissions: db.CanReadProject | db.CanApproveWorkflows})
	require.NoError(t, err)
	w, effective = runTaskMiddlewares(f, nodeTask.ID, &bob, 0)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, db.CanReadProject|db.CanApproveWorkflows, effective)
}
