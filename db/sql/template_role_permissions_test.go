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

	template, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "template", Playbook: "site.yml"})
	require.NoError(t, err)

	return store, projectID, template.ID
}

func createTemplatePermissionTestUser(t *testing.T, store *SqlDb, projectID int, username string, roleID int) db.User {
	t.Helper()
	user, err := store.CreateUserWithoutPassword(db.User{Username: username, Name: username, Email: username + "@example.com"})
	require.NoError(t, err)
	err = store.CreateProjectUser(projectID, user.ID, []int{roleID})
	require.NoError(t, err)
	return user
}

func Test_GetTemplatePermission_ReturnsZeroForNonMember(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	user, err := store.CreateUserWithoutPassword(db.User{Username: "non-member", Name: "non-member", Email: "non-member@example.com"})
	require.NoError(t, err)

	permissions, err := store.GetTemplatePermission(projectID, templateID, user.ID)
	require.NoError(t, err)
	assert.Zero(t, permissions)
}

func TestTemplateListWithTwoRolesHasNoDuplicates(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	guest, err := store.GetSharedRoleBySlug("guest")
	require.NoError(t, err)
	user := createTemplatePermissionTestUser(t, store, projectID, "two-roles", guest.ID)
	a, err := store.CreateRole(db.Role{Slug: "a", Name: "A", ProjectID: &projectID})
	require.NoError(t, err)
	b, err := store.CreateRole(db.Role{Slug: "b", Name: "B", ProjectID: &projectID})
	require.NoError(t, err)
	require.NoError(t, store.SetProjectUserRoles(projectID, user.ID, []int{a.ID, b.ID}))
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: a.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: b.ID, Permissions: db.CanUpdateProject})
	require.NoError(t, err)

	templates, err := store.GetTemplatesWithPermissions(projectID, user.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.NotNil(t, templates[0].Permissions)
	assert.Equal(t, db.CanRunProjectTasks|db.CanUpdateProject, *templates[0].Permissions)

	perm, err := store.GetTemplatePermission(projectID, templateID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanRunProjectTasks|db.CanUpdateProject, perm)
}

func TestCreateTemplateRole_RejectsRoleOfAnotherProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	template, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "t", Playbook: "p.yml"})
	require.NoError(t, err)
	foreign, err := store.CreateRole(db.Role{Slug: "foreign", Name: "F", ProjectID: &other.ID})
	require.NoError(t, err)

	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: template.ID, RoleID: foreign.ID})

	assert.ErrorContains(t, err, "Role is not available in this project")
}

func TestCreateTemplateRole_RefusesASecondGrantForTheRole(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	role, err := store.CreateRole(db.Role{Slug: "ops", Name: "Ops", ProjectID: &projectID})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: role.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)

	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: role.ID, Permissions: db.CanUpdateProject})

	assert.ErrorContains(t, err, "Role already has permissions on this template")
}
