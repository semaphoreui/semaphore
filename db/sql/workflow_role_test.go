package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertWorkflow writes a workflow template directly: the workflow store is a Pro feature, the table is in
// the community schema.
func insertWorkflow(t *testing.T, store *SqlDb, projectID int, name string) int {
	t.Helper()
	_, err := store.exec("insert into project__workflow_template (project_id, name) values (?, ?)", projectID, name)
	require.NoError(t, err)
	id, err := store.Sql().SelectInt(store.PrepareQuery("select id from project__workflow_template where project_id = ? and name = ?"), projectID, name)
	require.NoError(t, err)
	return int(id)
}

func newWorkflowRoleTest(t *testing.T) (*SqlDb, int, int) {
	t.Helper()
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	return store, project.ID, insertWorkflow(t, store, project.ID, "release")
}

func TestWorkflowRoles_RoundTrip(t *testing.T) {
	store, projectID, workflowID := newWorkflowRoleTest(t)
	role, err := store.CreateRole(db.Role{Slug: "ops", Name: "Ops", ProjectID: &projectID})
	require.NoError(t, err)
	grant := db.WorkflowRolePerm{ProjectID: projectID, WorkflowID: workflowID, RoleID: role.ID, Permissions: db.CanRunWorkflows}

	created, err := store.CreateWorkflowRole(grant)
	require.NoError(t, err)
	_, err = store.CreateWorkflowRole(grant)
	assert.ErrorContains(t, err, "Role already has permissions on this workflow")

	grants, err := store.GetWorkflowRoles(projectID, workflowID)
	require.NoError(t, err)
	require.Len(t, grants, 1)
	assert.Equal(t, "ops", grants[0].RoleSlug)

	created.Permissions = db.CanRunWorkflows | db.CanEditWorkflows
	require.NoError(t, store.UpdateWorkflowRole(created))
	stored, err := store.GetWorkflowRole(projectID, workflowID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanRunWorkflows|db.CanEditWorkflows, stored.Permissions)

	require.NoError(t, store.DeleteWorkflowRole(projectID, workflowID, created.ID))
	_, err = store.GetWorkflowRole(projectID, workflowID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// The grants of every role the member holds add up, per workflow.
func TestWorkflowGrants_AddUpPerWorkflow(t *testing.T) {
	store, projectID, release := newWorkflowRoleTest(t)
	other := insertWorkflow(t, store, projectID, "other")
	a, err := store.CreateRole(db.Role{Slug: "a", Name: "A", ProjectID: &projectID})
	require.NoError(t, err)
	b, err := store.CreateRole(db.Role{Slug: "b", Name: "B", ProjectID: &projectID})
	require.NoError(t, err)
	user := createTemplatePermissionTestUser(t, store, projectID, "two-roles", a.ID)
	require.NoError(t, store.SetProjectUserRoles(projectID, user.ID, []int{a.ID, b.ID}))
	for _, grant := range []db.WorkflowRolePerm{
		{WorkflowID: release, RoleID: a.ID, Permissions: db.CanRunWorkflows},
		{WorkflowID: release, RoleID: b.ID, Permissions: db.CanApproveWorkflows},
		{WorkflowID: other, RoleID: b.ID, Permissions: db.CanEditWorkflows},
	} {
		grant.ProjectID = projectID
		_, err = store.CreateWorkflowRole(grant)
		require.NoError(t, err)
	}

	perm, err := store.GetWorkflowPermission(projectID, release, user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanRunWorkflows|db.CanApproveWorkflows, perm)

	grants, err := store.GetWorkflowGrantsForUser(projectID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, map[int]db.ProjectUserPermission{release: db.CanRunWorkflows | db.CanApproveWorkflows, other: db.CanEditWorkflows}, grants)
}
