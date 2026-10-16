package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
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
