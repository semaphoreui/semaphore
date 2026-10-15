package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reproduces "Failed to delete user ... constraint failed: FOREIGN KEY
// constraint failed (787)": on SQLite session.user_id references user(id)
// without an ON DELETE action, so deleting any user who has ever logged in
// (session rows are only flagged expired, never removed) fails.
func TestDeleteUser_WithSession(t *testing.T) {
	store := InitConfigCreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
	})
	require.NoError(t, err)

	now := time.Now()
	_, err = store.CreateSession(db.Session{
		UserID:     user.ID,
		Created:    now,
		LastActive: now,
		IP:         "127.0.0.1",
		UserAgent:  "test",
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteUser(user.ID))

	_, err = store.GetUser(user.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// Same failure via task.user_id: the SQLite schema declares
// task.user_id REFERENCES user(id) with no ON DELETE action, so a user who
// has ever run a task cannot be deleted.
func TestDeleteUser_WithTask(t *testing.T) {
	store := InitConfigCreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
	})
	require.NoError(t, err)

	projectID, repositoryID := newTemplateTestProject(t, store)
	template, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "tpl",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		TemplateID: template.ID,
		ProjectID:  projectID,
		Status:     "success",
		Playbook:   "site.yml",
		UserID:     &user.ID,
		Created:    time.Now(),
	}, 0)
	require.NoError(t, err)

	require.NoError(t, store.DeleteUser(user.ID))

	_, err = store.GetUser(user.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	// The task itself must survive the deletion of its author.
	_, err = store.GetTask(projectID, task.ID)
	assert.NoError(t, err)
}

func createMembershipTestProjectAndUser(t *testing.T, store db.Store) (db.Project, db.User) {
	p, err := store.CreateProject(db.Project{Name: "membership"})
	require.NoError(t, err)
	u, err := store.CreateUserWithoutPassword(db.User{Username: "a", Name: "A", Email: "a@e.st"})
	require.NoError(t, err)
	return p, u
}

func slugsOf(roles []db.Role) []string {
	slugs := []string{}
	for _, r := range roles {
		slugs = append(slugs, r.Slug)
	}
	return slugs
}

func TestMembershipRoles(t *testing.T) {
	store := InitConfigCreateTestStore()
	p, u := createMembershipTestProjectAndUser(t, store)
	owner, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)
	runner, err := store.GetGlobalRoleBySlug("task_runner")
	require.NoError(t, err)

	err = store.CreateProjectUser(p.ID, u.ID, []int{runner.ID})
	require.NoError(t, err)
	roles, err := store.GetProjectUserRoles(p.ID, u.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"task_runner"}, slugsOf(roles))

	require.NoError(t, store.SetProjectUserRoles(p.ID, u.ID, []int{owner.ID, runner.ID}))
	roles, err = store.GetProjectUserRoles(p.ID, u.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"owner", "task_runner"}, slugsOf(roles))

	require.NoError(t, store.DeleteProjectUser(p.ID, u.ID))
	_, err = store.GetProjectUser(p.ID, u.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	assert.ErrorIs(t, store.SetProjectUserRoles(p.ID, u.ID, []int{owner.ID}), db.ErrNotFound, "a removed member does not come back")
	_, err = store.GetProjectUser(p.ID, u.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestMembershipRejectsForeignProjectRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	p1, u := createMembershipTestProjectAndUser(t, store)
	p2, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	foreign, err := store.CreateRole(db.Role{Slug: "x", Name: "X", ProjectID: &p2.ID})
	require.NoError(t, err)
	err = store.CreateProjectUser(p1.ID, u.ID, []int{foreign.ID})
	assert.Error(t, err)
}

func TestMembershipNeedsARole(t *testing.T) {
	store := InitConfigCreateTestStore()
	p, u := createMembershipTestProjectAndUser(t, store)
	runner, err := store.GetGlobalRoleBySlug("task_runner")
	require.NoError(t, err)

	assert.Error(t, store.CreateProjectUser(p.ID, u.ID, []int{}))
	_, err = store.GetProjectUser(p.ID, u.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	require.NoError(t, store.CreateProjectUser(p.ID, u.ID, []int{runner.ID}))
	assert.Error(t, store.SetProjectUserRoles(p.ID, u.ID, nil))
	roles, err := store.GetProjectUserRoles(p.ID, u.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"task_runner"}, slugsOf(roles), "an empty list does not remove the member")
}

func TestDeleteRoleRefusedWhileAssigned(t *testing.T) {
	store := InitConfigCreateTestStore()
	p, u := createMembershipTestProjectAndUser(t, store)
	r, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &p.ID})
	require.NoError(t, err)
	err = store.CreateProjectUser(p.ID, u.ID, []int{r.ID})
	require.NoError(t, err)
	assert.ErrorIs(t, store.DeleteRole(r.ID), db.ErrInvalidOperation)
}

func TestDeleteProjectRemovesRolesAndMemberships(t *testing.T) {
	store := InitConfigCreateTestStore()
	p, u := createMembershipTestProjectAndUser(t, store)
	r, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &p.ID})
	require.NoError(t, err)
	err = store.CreateProjectUser(p.ID, u.ID, []int{r.ID})
	require.NoError(t, err)
	require.NoError(t, store.DeleteProject(p.ID))
	_, err = store.GetProjectRole(p.ID, "dev")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestProjectUsersSortByRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	p, u1 := createMembershipTestProjectAndUser(t, store)
	u2, err := store.CreateUserWithoutPassword(db.User{Username: "b", Name: "B", Email: "b@e.st"})
	require.NoError(t, err)
	owner, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)
	guest, err := store.GetGlobalRoleBySlug("guest")
	require.NoError(t, err)
	err = store.CreateProjectUser(p.ID, u1.ID, []int{owner.ID})
	require.NoError(t, err)
	err = store.CreateProjectUser(p.ID, u2.ID, []int{guest.ID})
	require.NoError(t, err)
	users, err := store.GetProjectUsers(p.ID, db.RetrieveQueryParams{SortBy: "role"})
	require.NoError(t, err)
	assert.Equal(t, []string{"guest", "owner"}, []string{users[0].Roles[0].Slug, users[1].Roles[0].Slug})
}

func TestCountProjectOwners(t *testing.T) {
	store, projectID, _ := newTemplatePermissionTest(t)
	owner, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)
	guest, err := store.GetGlobalRoleBySlug("guest")
	require.NoError(t, err)
	createTemplatePermissionTestUser(t, store, projectID, "a", owner.ID)
	b := createTemplatePermissionTestUser(t, store, projectID, "b", guest.ID)
	createTemplatePermissionTestUser(t, store, projectID, "c", guest.ID)

	count, err := store.CountProjectOwners(projectID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	require.NoError(t, store.SetProjectUserRoles(projectID, b.ID, []int{owner.ID, guest.ID}))
	count, err = store.CountProjectOwners(projectID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// addMappedRole gives the user a role through a new mapping, as a login sync would.
func addMappedRole(t *testing.T, store *SqlDb, projectID int, userID int, roleID int, group string) int {
	t.Helper()
	_, err := store.exec(`
		insert into group_role_mapping (provider, group_name, project_id, role_id, created)
		values ('okta', ?, ?, ?, ?)
	`, group, projectID, roleID, time.Now().UTC())
	require.NoError(t, err)
	mappingID := roleIDOf(t, store, "select id from group_role_mapping where group_name=? and project_id=? and role_id=?", group, projectID, roleID)
	_, err = store.exec("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)", projectID, userID, mappingID)
	require.NoError(t, err)
	return mappingID
}

func TestMappedRoles_Reads(t *testing.T) {
	store, projectID, templateID := newTemplatePermissionTest(t)
	guest, err := store.GetGlobalRoleBySlug("guest")
	require.NoError(t, err)
	dev, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", Permissions: db.CanRunProjectTasks, ProjectID: &projectID})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: templateID, RoleID: dev.ID, Permissions: db.CanUpdateProject})
	require.NoError(t, err)

	mapped, err := store.CreateUserWithoutPassword(db.User{Username: "m", Name: "m", Email: "m@example.com"})
	require.NoError(t, err)
	addMappedRole(t, store, projectID, mapped.ID, dev.ID, "devops")
	both := createTemplatePermissionTestUser(t, store, projectID, "b", dev.ID)
	addMappedRole(t, store, projectID, both.ID, dev.ID, "admins")
	createTemplatePermissionTestUser(t, store, projectID, "g", guest.ID)

	t.Run("effective roles hold a mapped role once", func(t *testing.T) {
		roles, err := store.GetProjectUserRoles(projectID, both.ID)
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, dev.ID, roles[0].ID)
	})
	t.Run("a mapped role gives its mask and template grants", func(t *testing.T) {
		roles, err := store.GetProjectUserRoles(projectID, mapped.ID)
		require.NoError(t, err)
		require.Len(t, roles, 1)
		grant, err := store.GetTemplatePermission(projectID, templateID, mapped.ID)
		require.NoError(t, err)
		assert.Equal(t, db.CanUpdateProject, grant)
		templates, err := store.GetTemplatesWithPermissions(projectID, mapped.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.NotNil(t, templates[0].Permissions)
		assert.Equal(t, db.CanUpdateProject, *templates[0].Permissions)
	})
	t.Run("a mapped member sees the project", func(t *testing.T) {
		projects, err := store.GetProjects(mapped.ID)
		require.NoError(t, err)
		assert.Len(t, projects, 1)
		_, err = store.GetProjectUser(projectID, mapped.ID)
		assert.NoError(t, err)
	})
	t.Run("the member list keeps manual and mapped roles apart", func(t *testing.T) {
		users, err := store.GetProjectUsers(projectID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		byName := map[string]db.UserWithProjectRoles{}
		for _, u := range users {
			byName[u.Username] = u
		}
		require.Len(t, byName, 3)
		assert.Empty(t, byName["m"].Roles)
		require.Len(t, byName["m"].MappedRoles, 1)
		assert.Equal(t, "devops", byName["m"].MappedRoles[0].Group)
		assert.Equal(t, "dev", byName["m"].MappedRoles[0].RoleSlug)
		assert.Len(t, byName["b"].Roles, 1)
		assert.Len(t, byName["b"].MappedRoles, 1)
	})
	t.Run("a member keeps the manual and the mapped roles apart", func(t *testing.T) {
		member, err := store.GetProjectMember(projectID, both.ID)
		require.NoError(t, err)
		assert.Len(t, member.Roles, 1)
		assert.Len(t, member.MappedRoles, 1)
		_, err = store.GetProjectMember(projectID, 9999)
		assert.ErrorIs(t, err, db.ErrNotFound)
	})
}

func TestMappedRoles_ManualChanges(t *testing.T) {
	store, projectID, _ := newTemplatePermissionTest(t)
	guest, err := store.GetGlobalRoleBySlug("guest")
	require.NoError(t, err)
	manager, err := store.GetGlobalRoleBySlug("manager")
	require.NoError(t, err)
	newMapped := func(name string) db.User {
		user, err := store.CreateUserWithoutPassword(db.User{Username: name, Name: name, Email: name + "@example.com"})
		require.NoError(t, err)
		addMappedRole(t, store, projectID, user.ID, guest.ID, name)
		return user
	}

	t.Run("a mapped member can get manual roles", func(t *testing.T) {
		user := newMapped("a")
		require.NoError(t, store.CreateProjectUser(projectID, user.ID, []int{manager.ID}))
		member, err := store.GetProjectMember(projectID, user.ID)
		require.NoError(t, err)
		assert.Len(t, member.Roles, 1)
	})
	t.Run("a user with manual roles is already a member", func(t *testing.T) {
		user := createTemplatePermissionTestUser(t, store, projectID, "b", guest.ID)
		assert.ErrorIs(t, store.CreateProjectUser(projectID, user.ID, []int{manager.ID}), db.ErrInvalidOperation)
	})
	t.Run("a mapped member can lose the last manual role", func(t *testing.T) {
		user := newMapped("c")
		require.NoError(t, store.CreateProjectUser(projectID, user.ID, []int{manager.ID}))
		require.NoError(t, store.SetProjectUserRoles(projectID, user.ID, []int{}))
		member, err := store.GetProjectMember(projectID, user.ID)
		require.NoError(t, err)
		assert.Empty(t, member.Roles)
		assert.Len(t, member.MappedRoles, 1)
	})
	t.Run("a mapped member can get manual roles through set", func(t *testing.T) {
		user := newMapped("d")
		require.NoError(t, store.SetProjectUserRoles(projectID, user.ID, []int{manager.ID}))
		member, err := store.GetProjectMember(projectID, user.ID)
		require.NoError(t, err)
		assert.Len(t, member.Roles, 1)
	})
	t.Run("a member without mapped roles keeps at least one role", func(t *testing.T) {
		user := createTemplatePermissionTestUser(t, store, projectID, "e", guest.ID)
		var validation *common_errors.ValidationError
		assert.ErrorAs(t, store.SetProjectUserRoles(projectID, user.ID, []int{}), &validation)
	})
	t.Run("a user who is not a member is not found", func(t *testing.T) {
		user, err := store.CreateUserWithoutPassword(db.User{Username: "f", Name: "f", Email: "f@example.com"})
		require.NoError(t, err)
		assert.ErrorIs(t, store.SetProjectUserRoles(projectID, user.ID, []int{manager.ID}), db.ErrNotFound)
	})
	t.Run("removing the manual roles keeps the mapped membership", func(t *testing.T) {
		user := newMapped("g")
		require.NoError(t, store.CreateProjectUser(projectID, user.ID, []int{manager.ID}))
		require.NoError(t, store.DeleteProjectUser(projectID, user.ID))
		roles, err := store.GetProjectUserRoles(projectID, user.ID)
		require.NoError(t, err)
		assert.Len(t, roles, 1)
	})
	t.Run("owners count only manual roles", func(t *testing.T) {
		owner, err := store.GetGlobalRoleBySlug("owner")
		require.NoError(t, err)
		user, err := store.CreateUserWithoutPassword(db.User{Username: "h", Name: "h", Email: "h@example.com"})
		require.NoError(t, err)
		addMappedRole(t, store, projectID, user.ID, owner.ID, "owners")
		count, err := store.CountProjectOwners(projectID)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}
