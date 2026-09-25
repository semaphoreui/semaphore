package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTemplatePermissionTest(t *testing.T) (*SqlDb, int, int) {
	t.Helper()

	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	template, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "template",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	return store, projectID, template.ID
}

func createTemplatePermissionTestUser(
	t *testing.T,
	store *SqlDb,
	projectID int,
	username string,
	role db.ProjectUserRole,
) db.User {
	t.Helper()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: username,
		Name:     username,
		Email:    username + "@example.com",
	})
	require.NoError(t, err)

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: projectID,
		UserID:    user.ID,
		Role:      role,
	})
	require.NoError(t, err)

	return user
}

func Test_GetTemplatePermission_UsesBuiltinRolePermissionsFromDatabase(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	owner := createTemplatePermissionTestUser(t, store, projectID, "owner", db.ProjectOwner)

	// Patch the built-in role's permissions directly in the database so this test
	// can verify that authorization reads the persisted value instead of the Go map.
	_, err := store.exec(
		"update `role` set permissions=? where slug=?",
		db.CanUpdateProject,
		db.ProjectOwner)
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanUpdateProject, permissions)
}

func Test_GetTemplatePermission_AddsTemplateRolePermissions(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	taskRunner := createTemplatePermissionTestUser(
		t,
		store,
		projectID,
		"task_runner",
		db.ProjectTaskRunner,
	)

	templateUpdatePermission := db.CanManageProjectResources
	_, err := store.CreateTemplateRole(db.TemplateRolePerm{
		RoleSlug:    string(db.ProjectTaskRunner),
		TemplateID:  templateID,
		ProjectID:   projectID,
		Permissions: templateUpdatePermission,
	})
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, taskRunner.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanRunProjectTasks|templateUpdatePermission, permissions)
}

func Test_GetTemplatePermission_UsesCustomRolePermissions(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)

	customRole, err := store.CreateRole(db.Role{
		Slug:        "custom",
		Name:        "Custom",
		Permissions: db.CanManageProjectResources,
		ProjectID:   &projectID,
	})
	require.NoError(t, err)

	customUser := createTemplatePermissionTestUser(
		t,
		store,
		projectID,
		"custom",
		db.ProjectUserRole(customRole.Slug),
	)

	permissions, err := store.GetTemplatePermission(projectID, templateID, customUser.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanManageProjectResources, permissions)
}

func Test_GetTemplatePermission_ReturnsZeroForMissingRole(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	user := createTemplatePermissionTestUser(t, store, projectID, "missing", "missing")

	permissions, err := store.GetTemplatePermission(projectID, templateID, user.ID)
	require.NoError(t, err)
	assert.Zero(t, permissions)
}
