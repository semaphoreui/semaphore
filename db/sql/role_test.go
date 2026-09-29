package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RoleQueries(t *testing.T) {
	store := InitConfigCreateTestStore()

	project1, err := store.CreateProject(db.Project{
		Name: "project1",
	})
	require.NoError(t, err)
	project2, err := store.CreateProject(db.Project{
		Name: "project2",
	})
	require.NoError(t, err)

	globalRole, err := store.CreateRole(db.Role{
		Name: "Global custom",
	})
	require.NoError(t, err)
	projectRole, err := store.CreateRole(db.Role{
		Name:      "Project custom",
		ProjectID: &project1.ID,
	})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{
		Name:      "Other project custom",
		ProjectID: &project2.ID,
	})
	require.NoError(t, err)

	t.Run("global custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.GlobalRolesQuery{
			Kinds: db.RoleKindCustom,
		})
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, globalRole.ID, roles[0].ID)
	})

	t.Run("global built-in", func(t *testing.T) {
		roles, err := store.GetRoles(db.GlobalRolesQuery{
			Kinds: db.RoleKindBuiltin,
		})
		require.NoError(t, err)
		require.Len(t, roles, 4)
		for _, role := range roles {
			assert.True(t, role.IsBuiltin())
		}
	})

	t.Run("project custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.ProjectRolesQuery{
			ProjectID: project1.ID,
		})
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, projectRole.ID, roles[0].ID)
	})

	t.Run("available custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.AvailableRolesQuery{
			ProjectID: project1.ID,
			Kinds:     db.RoleKindCustom,
		})
		require.NoError(t, err)

		ids := make([]int, 0, len(roles))
		for _, role := range roles {
			ids = append(ids, role.ID)
		}
		assert.ElementsMatch(t, []int{globalRole.ID, projectRole.ID}, ids)
	})

	t.Run("available all", func(t *testing.T) {
		roles, err := store.GetRoles(db.AvailableRolesQuery{
			ProjectID: project1.ID,
			Kinds:     db.RoleKindAll,
		})
		require.NoError(t, err)
		assert.Len(t, roles, 6)
	})
}

func Test_GetRole_ByID(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	globalRole, err := store.CreateRole(db.Role{
		Name: "Global custom",
	})
	require.NoError(t, err)
	projectRole, err := store.CreateRole(db.Role{
		Name:      "Project custom",
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	role, err := store.GetRole(db.RoleByIDQuery{ID: globalRole.ID})
	require.NoError(t, err)
	assert.Equal(t, globalRole.ID, role.ID)

	role, err = store.GetRole(db.RoleByIDQuery{ID: projectRole.ID})
	require.NoError(t, err)
	assert.Equal(t, projectRole.ID, role.ID)

	_, err = store.GetRole(db.RoleByIDQuery{})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_GetRole_ByBuiltinKey(t *testing.T) {
	store := InitConfigCreateTestStore()

	role, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	require.NotNil(t, role.BuiltinKey)
	assert.Equal(t, db.BuiltinRoleOwner, *role.BuiltinKey)
	assert.Equal(t, db.ProjectUserPermission(15), role.Permissions)

	_, err = store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleKey("unknown")})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_RoleMutations_ProtectBuiltinRoles(t *testing.T) {
	store := InitConfigCreateTestStore()
	original, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)

	updated := original
	updated.Name = "Modified owner"
	updated.Permissions = 0
	require.ErrorIs(t, store.UpdateRole(updated), db.ErrInvalidOperation)
	require.ErrorIs(t, store.DeleteRole(original.ID), db.ErrInvalidOperation)

	after, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	assert.Equal(t, original, after)
}

func Test_CreateRole_RejectsBuiltinKey(t *testing.T) {
	store := InitConfigCreateTestStore()
	key := db.BuiltinRoleOwner

	_, err := store.CreateRole(db.Role{
		Name:       "Custom built-in",
		BuiltinKey: &key,
	})
	require.Error(t, err)
}

func Test_CreateRole_PersistsGeneratedID(t *testing.T) {
	store := InitConfigCreateTestStore()

	created, err := store.CreateRole(db.Role{
		Name: "Custom",
	})
	require.NoError(t, err)
	require.Positive(t, created.ID)
	assert.Nil(t, created.BuiltinKey)

	persisted, err := store.GetRole(db.RoleByIDQuery{ID: created.ID})
	require.NoError(t, err)
	assert.Equal(t, created.ID, persisted.ID)
	assert.Equal(t, created.Name, persisted.Name)
	assert.Nil(t, persisted.BuiltinKey)
}

func Test_CreateRole_AllowsDuplicateNames(t *testing.T) {
	store := InitConfigCreateTestStore()

	first, err := store.CreateRole(db.Role{
		Name: "Duplicate",
	})
	require.NoError(t, err)
	second, err := store.CreateRole(db.Role{
		Name: "Duplicate",
	})
	require.NoError(t, err)
	assert.NotEqual(t, first.ID, second.ID)
}

func Test_UpdateRole_RejectsScopeChange(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name: "Custom",
	})
	require.NoError(t, err)

	role.ProjectID = &project.ID
	require.Error(t, store.UpdateRole(role))

	persisted, err := store.GetRole(db.RoleByIDQuery{ID: role.ID})
	require.NoError(t, err)
	assert.Nil(t, persisted.ProjectID)
}

func Test_Role_IsAvailableToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{
		Name: "other project",
	})
	require.NoError(t, err)

	ownerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	globalRole, err := store.CreateRole(db.Role{
		Name: "Global custom",
	})
	require.NoError(t, err)
	projectRole, err := store.CreateRole(db.Role{
		Name:      "Project custom",
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	assert.True(t, ownerRole.IsAvailableToProject(project.ID))
	assert.True(t, globalRole.IsAvailableToProject(project.ID))
	assert.True(t, projectRole.IsAvailableToProject(project.ID))
	assert.False(t, projectRole.IsAvailableToProject(otherProject.ID))
}

func Test_CreateProjectUser_RejectsRoleFromAnotherProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{
		Name: "other project",
	})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "user",
		Name:     "User",
		Email:    "user@example.com",
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name:      "Other project custom",
		ProjectID: &otherProject.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    user.ID,
		RoleID:    role.ID,
	})
	require.Error(t, err)
}

func Test_UpdateProjectUser_RejectsRoleFromAnotherProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{
		Name: "other project",
	})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "user",
		Name:     "User",
		Email:    "user@example.com",
	})
	require.NoError(t, err)
	owner, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    user.ID,
		RoleID:    owner.ID,
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name:      "Other project custom",
		ProjectID: &otherProject.ID,
	})
	require.NoError(t, err)

	err = store.UpdateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    user.ID,
		RoleID:    role.ID,
	})
	require.Error(t, err)
}

func Test_CreateTemplateRole_RejectsRoleFromAnotherProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{
		Name: "other project",
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name:      "Other project custom",
		ProjectID: &otherProject.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateRole(db.TemplateRolePerm{
		ProjectID:  project.ID,
		TemplateID: 1,
		RoleID:     role.ID,
	})
	require.Error(t, err)
}

func Test_CreateProjectInvite_RequiresAvailableBuiltinRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	customRole, err := store.CreateRole(db.Role{
		Name: "Global custom",
	})
	require.NoError(t, err)

	_, err = store.CreateProjectInvite(db.ProjectInvite{
		ProjectID: project.ID,
		RoleID:    customRole.ID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "built-in role")
}

func Test_DeleteRole_ReturnsConflictWhenRoleIsReferenced(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "user",
		Name:     "User",
		Email:    "user@example.com",
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name: "Global custom",
	})
	require.NoError(t, err)

	_, err = store.exec(
		"insert into project__user (project_id, user_id, role_id) values (?, ?, ?)",
		project.ID,
		user.ID,
		role.ID)
	require.NoError(t, err)

	err = store.DeleteRole(role.ID)
	require.ErrorIs(t, err, db.ErrInvalidOperation)
	assert.Contains(t, err.Error(), "assigned or granted")
}

func Test_DeleteProject_RemovesRoleRelationshipsBeforeProjectRoles(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "user",
		Name:     "User",
		Email:    "user@example.com",
	})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{
		Name:      "Project custom",
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	_, err = store.exec(
		"insert into project__user (project_id, user_id, role_id) values (?, ?, ?)",
		project.ID,
		user.ID,
		role.ID)
	require.NoError(t, err)

	require.NoError(t, store.DeleteProject(project.ID))
	_, err = store.GetRole(db.RoleByIDQuery{ID: role.ID})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_DeleteRole_DeletesUnreferencedCustomRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	role, err := store.CreateRole(db.Role{
		Name: "Custom",
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteRole(role.ID))
	_, err = store.GetRole(db.RoleByIDQuery{ID: role.ID})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_RoleQueries_RejectInvalidFilters(t *testing.T) {
	store := InitConfigCreateTestStore()

	_, err := store.GetRoles(nil)
	assert.Error(t, err)

	// Kinds is required; its zero value is intentionally invalid so callers
	// cannot accidentally include or exclude built-in roles.
	_, err = store.GetRoles(db.GlobalRolesQuery{})
	assert.Error(t, err)

	_, err = store.GetRoles(db.AvailableRolesQuery{Kinds: db.RoleKind(8)})
	assert.Error(t, err)
}
