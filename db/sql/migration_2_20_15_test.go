package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Migration 2.20.14 added global permissions as a column of the role table; 2.20.15 moves them to their own
// table together with the assignments, so a database that lived through both ends up like a fresh one.
func TestMigration_2_20_15_MovesGlobalPermissionsToGlobalRoles(t *testing.T) {
	before := "2.20.13"
	store := InitConfigCreateTestStoreAt(&before)
	target := "2.20.14"
	require.NoError(t, db.Migrate(store, &target))

	insertRole := func(slug string, permissions int, global int) int {
		id, err := store.insert("id", "insert into `role` (slug, name, permissions, global_permissions, project_id, scope, builtin) values (?, ?, ?, ?, null, 0, false)", slug, slug, permissions, global)
		require.NoError(t, err)
		return id
	}
	auditor := insertRole("auditor", 0, int(db.CanReadGlobalAudit))                     // only global: leaves the role table
	ops := insertRole("ops", int(db.CanRunProjectTasks), int(db.CanManageGlobalSystem)) // both: stays as a project role
	plain := insertRole("plain", int(db.CanRunProjectTasks), 0)                         // untouched
	used := insertRole("used", 0, int(db.CanManageGlobalUsers))                         // held in a project: stays
	alice, err := store.CreateUserWithoutPassword(db.User{Username: "alice", Name: "A", Email: "a@e.st"})
	require.NoError(t, err)
	bob, err := store.CreateUserWithoutPassword(db.User{Username: "bob", Name: "B", Email: "b@e.st"})
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	require.NoError(t, store.CreateProjectUser(project.ID, bob.ID, []int{used}))
	for _, a := range [][2]int{{alice.ID, auditor}, {alice.ID, ops}, {bob.ID, ops}} {
		_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", a[0], a[1])
		require.NoError(t, err)
	}

	target = "2.20.15"
	require.NoError(t, db.Migrate(store, &target))

	var globalRoles []db.GlobalRole
	_, err = store.selectAll(&globalRoles, "select * from global_role order by slug")
	require.NoError(t, err)
	require.Len(t, globalRoles, 3)
	bySlug := map[string]db.GlobalRole{}
	for _, role := range globalRoles {
		bySlug[role.Slug] = role
	}
	assert.Equal(t, db.CanReadGlobalAudit, bySlug["auditor"].Permissions)
	assert.Equal(t, db.CanManageGlobalSystem, bySlug["ops"].Permissions)
	assert.Equal(t, db.CanManageGlobalUsers, bySlug["used"].Permissions)

	slugsOf := func(assignments []db.GlobalRoleAssignment) (slugs []string) {
		for _, a := range assignments {
			assert.Equal(t, bySlug[a.RoleSlug].ID, a.RoleID, "the assignment points at the global role")
			slugs = append(slugs, a.RoleSlug)
		}
		return
	}
	assert.ElementsMatch(t, []string{"auditor", "ops"}, slugsOf(globalAssignmentsOf(t, store, alice.ID)))
	assert.ElementsMatch(t, []string{"ops"}, slugsOf(globalAssignmentsOf(t, store, bob.ID)))

	_, err = store.GetSharedRoleBySlug("auditor")
	assert.ErrorIs(t, err, db.ErrNotFound, "a role that only carried global permissions is gone")
	for _, slug := range []string{"ops", "plain", "used"} {
		_, err = store.GetSharedRoleBySlug(slug)
		assert.NoError(t, err, slug)
	}
	role, err := store.GetSharedRoleBySlug("ops")
	require.NoError(t, err)
	assert.Equal(t, db.CanRunProjectTasks, role.Permissions)
	assert.Equal(t, plain, plain)
	_, err = store.Sql().SelectInt(store.PrepareQuery("select global_permissions from `role` limit 1"))
	assert.Error(t, err, "the column is gone")

	_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", alice.ID, bySlug["ops"].ID)
	assert.Error(t, err, "a user holds a role once")
	_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", alice.ID, 99999)
	assert.Error(t, err, "an assignment needs a global role")
	revision, err := store.Sql().SelectInt(store.PrepareQuery("select revision from global_role_state where id=1"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), revision)
}
