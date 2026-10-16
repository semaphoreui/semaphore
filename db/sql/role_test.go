package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectOrGlobalRoleBySlug_IgnoresCase(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "Deployer", Name: "D", ProjectID: &project.ID})
	require.NoError(t, err)

	role, err := store.GetProjectOrGlobalRoleBySlug(project.ID, "deployer")

	require.NoError(t, err)
	assert.Equal(t, "Deployer", role.Slug)
}

func TestGetProjectOrGlobalRoleBySlug_ProjectRoleWins(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	// Two HA nodes created these at the same moment, past the slug clash check.
	_, err = store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ci', 'Global', 0, null, 0, false)")
	require.NoError(t, err)
	_, err = store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ci', 'Project', 0, ?, ?, false)", project.ID, project.ID)
	require.NoError(t, err)

	role, err := store.GetProjectOrGlobalRoleBySlug(project.ID, "ci")

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
	_, err = store.GetProjectOrGlobalRoleBySlug(project.ID, "foreign")
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

	globals, err := store.GetGlobalRoles()
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
	owner, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)

	require.NoError(t, store.UpdateRole(db.Role{ID: owner.ID, Name: "Changed", Permissions: 0}))
	assert.ErrorIs(t, store.DeleteRole(owner.ID), db.ErrNotFound)

	got, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)
	assert.Equal(t, owner, got)
}

func TestDeleteRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	role, err := store.CreateRole(db.Role{Slug: "custom", Name: "Custom"})
	require.NoError(t, err)

	require.NoError(t, store.DeleteRole(role.ID))

	_, err = store.GetGlobalRoleBySlug("custom")
	assert.ErrorIs(t, err, db.ErrNotFound)
	assert.ErrorIs(t, store.DeleteRole(role.ID), db.ErrNotFound)
}
