package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// visibilityFixture: templates A and B, workflow "release" with a run whose node task uses template B, and four
// members: reader (read_project), grantee (sees A through a grant), runGrantee (sees "release" through a grant)
// and nobody.
type visibilityFixture struct {
	store      *SqlDb
	projectID  int
	a, b       db.Template
	workflowID int
	taskA      db.Task
	taskB      db.Task
	nodeTask   db.Task
	reader     db.User
	grantee    db.User
	runGrantee db.User
	nobody     db.User
}

func newVisibilityFixture(t *testing.T) visibilityFixture {
	t.Helper()
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)
	f := visibilityFixture{store: store, projectID: projectID}
	var err error
	f.a, err = store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "a", Playbook: "a.yml"})
	require.NoError(t, err)
	f.b, err = store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "b", Playbook: "b.yml"})
	require.NoError(t, err)
	f.workflowID = insertWorkflow(t, store, projectID, "release")
	_, err = store.exec("insert into project__workflow_run (project_id, workflow_template_id, status) values (?, ?, ?)", projectID, f.workflowID, "running")
	require.NoError(t, err)
	runID, err := store.Sql().SelectInt(store.PrepareQuery("select id from project__workflow_run where project_id = ?"), projectID)
	require.NoError(t, err)
	run := int(runID)
	f.taskA, err = store.CreateTask(db.Task{ProjectID: projectID, TemplateID: f.a.ID, Created: time.Now()}, 0)
	require.NoError(t, err)
	f.taskB, err = store.CreateTask(db.Task{ProjectID: projectID, TemplateID: f.b.ID, Created: time.Now()}, 0)
	require.NoError(t, err)
	f.nodeTask, err = store.CreateTask(db.Task{ProjectID: projectID, TemplateID: f.b.ID, WorkflowRunID: &run, Created: time.Now()}, 0)
	require.NoError(t, err)

	member := func(name string, mask db.ProjectUserPermission) (db.User, db.Role) {
		role, err := store.CreateRole(db.Role{Slug: name, Name: name, Permissions: mask, ProjectID: &projectID})
		require.NoError(t, err)
		user, err := store.CreateUserWithoutPassword(db.User{Username: name, Name: name, Email: name + "@example.com"})
		require.NoError(t, err)
		require.NoError(t, store.CreateProjectUser(projectID, user.ID, []int{role.ID}))
		return user, role
	}
	f.reader, _ = member("reader", db.CanReadProject)
	f.nobody, _ = member("nobody", 0)
	var role db.Role
	f.grantee, role = member("grantee", 0)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: f.a.ID, RoleID: role.ID, Permissions: db.CanReadProject})
	require.NoError(t, err)
	f.runGrantee, role = member("run-grantee", 0)
	_, err = store.CreateWorkflowRole(db.WorkflowRolePerm{ProjectID: projectID, WorkflowID: f.workflowID, RoleID: role.ID, Permissions: db.CanReadProject | db.CanRunWorkflows})
	require.NoError(t, err)
	return f
}

func templateNames(templates []db.TemplateWithPerms) []string {
	names := []string{}
	for _, template := range templates {
		names = append(names, template.Name)
	}
	return names
}

func taskIDs(tasks []db.TaskWithTpl) []int {
	ids := []int{}
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func TestVisibility_Templates(t *testing.T) {
	f := newVisibilityFixture(t)
	tests := []struct {
		name   string
		viewer *db.Viewer
		want   []string
	}{
		{"reader sees the project", &db.Viewer{UserID: f.reader.ID}, []string{"a", "b"}},
		{"grantee sees the granted template", &db.Viewer{UserID: f.grantee.ID}, []string{"a"}},
		{"a workflow grant shows no template", &db.Viewer{UserID: f.runGrantee.ID}, []string{}},
		{"nobody sees nothing", &db.Viewer{UserID: f.nobody.ID}, []string{}},
		{"admin is not filtered", &db.Viewer{UserID: f.nobody.ID, Admin: true}, []string{"a", "b"}},
		{"no viewer, no filter", nil, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := f.store.GetTemplatesWithPermissions(f.projectID, f.nobody.ID, db.TemplateFilter{Viewer: tt.viewer}, db.RetrieveQueryParams{})
			require.NoError(t, err)
			assert.Equal(t, tt.want, templateNames(templates))
		})
	}
}

func TestVisibility_Tasks(t *testing.T) {
	f := newVisibilityFixture(t)
	tests := []struct {
		name   string
		viewer *db.Viewer
		want   []int
	}{
		{"reader sees every task", &db.Viewer{UserID: f.reader.ID}, []int{f.nodeTask.ID, f.taskB.ID, f.taskA.ID}},
		{"grantee sees the tasks of the granted template", &db.Viewer{UserID: f.grantee.ID}, []int{f.taskA.ID}},
		{"a workflow grant shows the tasks of its runs", &db.Viewer{UserID: f.runGrantee.ID}, []int{f.nodeTask.ID}},
		{"nobody sees nothing", &db.Viewer{UserID: f.nobody.ID}, []int{}},
		{"admin is not filtered", &db.Viewer{UserID: f.nobody.ID, Admin: true}, []int{f.nodeTask.ID, f.taskB.ID, f.taskA.ID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := db.RetrieveQueryParams{TaskFilter: &db.TaskFilter{Viewer: tt.viewer}}
			tasks, err := f.store.GetProjectTasks(f.projectID, params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, taskIDs(tasks))

			stats, err := f.store.GetTaskStats(f.projectID, nil, db.TaskStatUnitDay, db.TaskFilter{Viewer: tt.viewer})
			require.NoError(t, err)
			count := 0
			for _, stat := range stats {
				for _, n := range stat.CountByStatus {
					count += n
				}
			}
			assert.Equal(t, len(tt.want), count, "stats count only the visible tasks")
		})
	}
}
