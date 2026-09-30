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
	roleID int,
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
		RoleID:    roleID,
	})
	require.NoError(t, err)

	return user
}

func Test_GetTemplatePermission_UsesBuiltinRolePermissionsFromDatabase(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	ownerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	owner := createTemplatePermissionTestUser(t, store, projectID, "owner", ownerRole.ID)

	// Patch the built-in role's permissions directly in the database so this test
	// can verify that authorization reads the persisted value instead of a Go map.
	_, err = store.exec(
		"update `role` set permissions=? where builtin_key=?",
		db.CanUpdateProject,
		db.BuiltinRoleOwner)
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanUpdateProject, permissions)
}

func Test_GetTemplatePermission_AddsTemplateRolePermissions(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	taskRunnerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleTaskRunner})
	require.NoError(t, err)
	taskRunner := createTemplatePermissionTestUser(
		t,
		store,
		projectID,
		"task_runner",
		taskRunnerRole.ID,
	)

	templateUpdatePermission := db.CanManageProjectResources
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{
		RoleID:      taskRunnerRole.ID,
		TemplateID:  templateID,
		ProjectID:   projectID,
		Permissions: templateUpdatePermission,
	})
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, taskRunner.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanRunProjectTasks|templateUpdatePermission, permissions)
}

func Test_CreateTemplateRole_ReturnsConflictForDuplicateRole(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	managerRole, err := store.GetRole(db.BuiltinRoleQuery{
		Key: db.BuiltinRoleManager,
	})
	require.NoError(t, err)
	grant := db.TemplateRolePerm{
		RoleID:      managerRole.ID,
		TemplateID:  templateID,
		ProjectID:   projectID,
		Permissions: db.CanRunProjectTasks,
	}

	_, err = store.CreateTemplateRole(grant)
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(grant)

	require.ErrorIs(t, err, db.ErrInvalidOperation)
	assert.Contains(t, err.Error(), "Role already has permissions for this template")
}

func Test_GetTemplatePermission_UsesCustomRolePermissions(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)

	customRole, err := store.CreateRole(db.Role{
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
		customRole.ID,
	)

	permissions, err := store.GetTemplatePermission(projectID, templateID, customUser.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanManageProjectResources, permissions)
}

func Test_GetTemplatePermission_ReturnsZeroForNonMember(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "non-member",
		Name:     "non-member",
		Email:    "non-member@example.com",
	})
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, user.ID)
	require.NoError(t, err)
	assert.Zero(t, permissions)
}
