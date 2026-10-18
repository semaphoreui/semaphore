package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectOrSharedRoleBySlug_IgnoresCase(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "Deployer", Name: "D", ProjectID: &project.ID})
	require.NoError(t, err)

	role, err := store.GetProjectOrSharedRoleBySlug(project.ID, "deployer")

	require.NoError(t, err)
	assert.Equal(t, "Deployer", role.Slug)
}

func TestGetProjectOrSharedRoleBySlug_ProjectRoleWins(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	// Two HA nodes created these at the same moment, past the slug clash check.
	_, err = store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ci', 'Global', 0, null, 0, false)")
	require.NoError(t, err)
	_, err = store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ci', 'Project', 0, ?, ?, false)", project.ID, project.ID)
	require.NoError(t, err)

	role, err := store.GetProjectOrSharedRoleBySlug(project.ID, "ci")

	require.NoError(t, err)
	assert.Equal(t, "Project", role.Name)
}

func TestGetProjectRole_IgnoresGlobalAndOtherProjects(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "global", Name: "G"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "foreign", Name: "F", ProjectID: &other.ID})
	require.NoError(t, err)

	_, err = store.GetProjectRole(project.ID, "global")
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetProjectRole(project.ID, "foreign")
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetProjectOrSharedRoleBySlug(project.ID, "foreign")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestGetGlobalRoles_IncludesBuiltins(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "auditor", Name: "Auditor"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "local", Name: "Local", ProjectID: &project.ID})
	require.NoError(t, err)

	globals, err := store.GetSharedRoles()
	require.NoError(t, err)
	projectRoles, err := store.GetProjectRoles(project.ID)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"auditor", "guest", "manager", "owner", "task_runner"}, slugsOf(globals))
	assert.Equal(t, []string{"local"}, slugsOf(projectRoles))
}

func TestCreateRole_SlugScopes(t *testing.T) {
	store := InitConfigCreateTestStore()
	p1, err := store.CreateProject(db.Project{Name: "p1"})
	require.NoError(t, err)
	p2, err := store.CreateProject(db.Project{Name: "p2"})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &p1.ID})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &p2.ID})
	require.NoError(t, err, "the same slug in two projects")

	_, err = store.CreateRole(db.Role{Slug: "DEV", Name: "Dev", ProjectID: &p1.ID})
	assert.ErrorContains(t, err, "Role slug is taken")
	_, err = store.CreateRole(db.Role{Slug: "dev", Name: "Dev"})
	assert.ErrorContains(t, err, "Role slug is taken", "a global slug used by a project")

	_, err = store.CreateRole(db.Role{Slug: "auditor", Name: "A"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "Auditor", Name: "A", ProjectID: &p1.ID})
	assert.ErrorContains(t, err, "Role slug is taken", "a project slug used by a global role")
	_, err = store.CreateRole(db.Role{Slug: "AUDITOR", Name: "A"})
	assert.ErrorContains(t, err, "Role slug is taken", "two global roles")
}

func TestUpdateRole_ChangesNameAndPermissionsOnly(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{Slug: "ops", Name: "Ops", Permissions: db.CanRunProjectTasks, ProjectID: &project.ID})
	require.NoError(t, err)

	require.NoError(t, store.UpdateRole(db.Role{ID: role.ID, Slug: "other", Name: "Changed", Permissions: db.CanUpdateProject}))

	got, err := store.GetProjectRole(project.ID, "ops")
	require.NoError(t, err)
	assert.Equal(t, "Changed", got.Name)
	assert.Equal(t, db.CanUpdateProject, got.Permissions)
}

func TestBuiltinRolesCannotChange(t *testing.T) {
	store := InitConfigCreateTestStore()
	owner, err := store.GetSharedRoleBySlug("owner")
	require.NoError(t, err)

	require.NoError(t, store.UpdateRole(db.Role{ID: owner.ID, Name: "Changed", Permissions: 0}))
	assert.ErrorIs(t, store.DeleteRole(owner.ID), db.ErrNotFound)

	got, err := store.GetSharedRoleBySlug("owner")
	require.NoError(t, err)
	assert.Equal(t, owner, got)
}

func TestDeleteRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	role, err := store.CreateRole(db.Role{Slug: "custom", Name: "Custom"})
	require.NoError(t, err)

	require.NoError(t, store.DeleteRole(role.ID))

	_, err = store.GetSharedRoleBySlug("custom")
	assert.ErrorIs(t, err, db.ErrNotFound)
	assert.ErrorIs(t, store.DeleteRole(role.ID), db.ErrNotFound)
}

func TestDeleteRoleWithAssignments(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	role, err := store.CreateRole(db.Role{Slug: "deployer", Name: "Deployer"})
	require.NoError(t, err)
	guest, err := store.GetSharedRoleBySlug("guest")
	require.NoError(t, err)
	only := createTemplatePermissionTestUser(t, store, projectID, "only", role.ID)
	both := createTemplatePermissionTestUser(t, store, projectID, "both", role.ID)
	require.NoError(t, store.SetProjectUserRoles(projectID, both.ID, []int{role.ID, guest.ID}))
	grant, err := store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: role.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)
	check := func(assigned db.RoleAssignments) {
		t.Helper()
		require.Len(t, assigned.Members, 2)
		assert.Equal(t, "only", assigned.Members[0].Username)
		assert.Equal(t, []string{"deployer"}, slugsOf(assigned.Members[0].Roles))
		assert.Equal(t, "both", assigned.Members[1].Username)
		assert.ElementsMatch(t, []string{"deployer", "guest"}, slugsOf(assigned.Members[1].Roles))
		require.Len(t, assigned.Grants, 1)
		assert.Equal(t, grant.ID, assigned.Grants[0].ID)
		assert.Equal(t, db.RoleImpact{Members: 2, Projects: 1, TemplateGrants: 1}, assigned.Impact())
	}

	assigned, err := store.GetRoleAssignments(role.ID)
	require.NoError(t, err)
	check(assigned)
	assert.ErrorIs(t, store.DeleteRole(role.ID), db.ErrInvalidOperation, "the foreign keys keep an assigned role")

	removed, err := store.DeleteRoleWithAssignments(role.ID, false)

	require.NoError(t, err)
	check(removed)
	_, err = store.GetSharedRoleBySlug("deployer")
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetProjectUser(projectID, only.ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "a member whose only role it was leaves the project")
	member, err := store.GetProjectMember(projectID, both.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"guest"}, slugsOf(member.Roles))
	grants, err := store.GetTemplateRoles(projectID, templateID)
	require.NoError(t, err)
	assert.Empty(t, grants)
}

func TestDeleteRoleWithAssignments_BuiltinRole(t *testing.T) {
	store, projectID, _ := newTemplatePermissionTest(t)
	owner, err := store.GetSharedRoleBySlug("owner")
	require.NoError(t, err)
	user := createTemplatePermissionTestUser(t, store, projectID, "u", owner.ID)

	_, err = store.DeleteRoleWithAssignments(owner.ID, false)

	assert.ErrorIs(t, err, db.ErrNotFound)
	member, err := store.GetProjectMember(projectID, user.ID)
	require.NoError(t, err)
	assert.Len(t, member.Roles, 1, "the rollback keeps the assignment")
}

func TestDeleteRoleWithAssignments_Mappings(t *testing.T) {
	store, projectID, _ := newTemplatePermissionTest(t)
	guest, err := store.GetSharedRoleBySlug("guest")
	require.NoError(t, err)
	dev, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &projectID})
	require.NoError(t, err)
	keeps := createTemplatePermissionTestUser(t, store, projectID, "keeps", dev.ID)
	addMappedRole(t, store, projectID, keeps.ID, guest.ID, "admins")
	mapped, err := store.CreateUserWithoutPassword(db.User{Username: "mapped", Name: "mapped", Email: "mapped@example.com"})
	require.NoError(t, err)
	mappingID := addMappedRole(t, store, projectID, mapped.ID, dev.ID, "devops")
	check := func(assigned db.RoleAssignments) {
		t.Helper()
		require.Len(t, assigned.Members, 1)
		assert.Equal(t, keeps.ID, assigned.Members[0].UserID)
		assert.True(t, assigned.Members[0].Mapped)
		require.Len(t, assigned.Mappings, 1)
		assert.Equal(t, mappingID, assigned.Mappings[0].ID)
		assert.Equal(t, "devops", assigned.Mappings[0].Group)
		assert.Equal(t, "dev", assigned.Mappings[0].RoleSlug)
		require.Len(t, assigned.MappedRoles, 1)
		assert.Equal(t, mapped.ID, assigned.MappedRoles[0].UserID)
		assert.Equal(t, mappingID, assigned.MappedRoles[0].MappingID)
		assert.Equal(t, db.RoleImpact{Members: 2, Projects: 1, Mappings: 1}, assigned.Impact())
	}

	assigned, err := store.GetRoleAssignments(dev.ID)
	require.NoError(t, err)
	check(assigned)
	assert.ErrorIs(t, store.DeleteRole(dev.ID), db.ErrInvalidOperation)

	_, err = store.DeleteRoleWithAssignments(dev.ID, false)

	assert.ErrorIs(t, err, db.ErrInvalidOperation, "the mapping keeps the role")
	assigned, err = store.GetRoleAssignments(dev.ID)
	require.NoError(t, err)
	check(assigned)

	removed, err := store.DeleteRoleWithAssignments(dev.ID, true)

	require.NoError(t, err)
	check(removed)
	_, err = store.GetProjectRole(projectID, "dev")
	assert.ErrorIs(t, err, db.ErrNotFound)
	assert.Equal(t, 0, roleIDOf(t, store, "select count(*) from group_role_mapping where id=?", mappingID))
	_, err = store.GetProjectUser(projectID, keeps.ID)
	assert.NoError(t, err, "the mapped guest role keeps the membership")
	_, err = store.GetProjectUser(projectID, mapped.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// A mapping of the deleted role goes with it, so it does not keep the membership.
func TestGetRoleAssignments_MappingOfTheSameRoleIsNotMapped(t *testing.T) {
	store, projectID, _ := newTemplatePermissionTest(t)
	dev, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &projectID})
	require.NoError(t, err)
	user := createTemplatePermissionTestUser(t, store, projectID, "u", dev.ID)
	addMappedRole(t, store, projectID, user.ID, dev.ID, "devops")

	assigned, err := store.GetRoleAssignments(dev.ID)

	require.NoError(t, err)
	require.Len(t, assigned.Members, 1)
	assert.False(t, assigned.Members[0].Mapped)
	assert.Equal(t, db.RoleImpact{Members: 1, Projects: 1, Mappings: 1}, assigned.Impact())
}
