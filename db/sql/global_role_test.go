package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGlobalRoleTestUser(t *testing.T, store *SqlDb, username string, admin bool) db.User {
	t.Helper()
	user, err := store.CreateUserWithoutPassword(db.User{Username: username, Name: username, Email: username + "@e.st", Admin: admin})
	require.NoError(t, err)
	return user
}

// createGlobalRole writes the row the Pro store would write; this store only reads global roles.
func createGlobalRole(t *testing.T, store *SqlDb, slug string, permissions db.GlobalPermission) int {
	t.Helper()
	id, err := store.insert("id", "insert into global_role (slug, name, permissions) values (?, ?, ?)", slug, slug, permissions)
	require.NoError(t, err)
	return id
}

// assignGlobally writes the row the Pro store would write; this store only reads and removes assignments.
func assignGlobally(t *testing.T, store *SqlDb, userID int, roleID int) int {
	t.Helper()
	id, err := store.insert("id", "insert into user__global_role (user_id, role_id) values (?, ?)", userID, roleID)
	require.NoError(t, err)
	return id
}

func globalAssignmentsOf(t *testing.T, store *SqlDb, userID int) (assignments []db.GlobalRoleAssignment) {
	t.Helper()
	_, err := store.selectAll(&assignments, GlobalRoleAssignmentSelect+" where a.user_id=? order by a.id", userID)
	require.NoError(t, err)
	return
}

func TestGlobalRoleAssignmentSelect_JoinsTheGlobalRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	id := assignGlobally(t, store, user.ID, auditor)

	assert.Equal(t, []db.GlobalRoleAssignment{{
		ID: id, UserID: user.ID, RoleID: auditor, Username: "alice",
		RoleSlug: "auditor", RoleName: "auditor", GlobalPermissions: db.CanReadGlobalAudit,
	}}, globalAssignmentsOf(t, store, user.ID))
}

// A global role and a project role may share a slug: they live in different tables.
func TestGlobalRole_SlugIsIndependentOfProjectRoles(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.CreateRole(db.Role{Slug: "auditor", Name: "Project auditor"})
	require.NoError(t, err)
	createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	_, err = store.insert("id", "insert into global_role (slug, name, permissions) values (?, ?, ?)", "auditor", "again", 0)
	assert.Error(t, err, "a global role slug is unique among global roles")
}

func TestDeleteUser_RemovesGlobalAssignments(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	assignGlobally(t, store, user.ID, auditor)

	require.NoError(t, store.DeleteUser(user.ID))

	count, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from user__global_role where role_id=?"), auditor)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// The last user who can manage global roles cannot lose that ability through this store: by losing the
// admin flag or the account. The Pro store protects the role and the assignment themselves.
func TestLastGlobalAdministrator(t *testing.T) {
	store := InitConfigCreateTestStore()
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	manager := createGlobalRole(t, store, "role_manager", db.CanManageGlobalRoles|db.CanReadGlobalAudit)
	assignGlobally(t, store, alice.ID, manager)

	assert.ErrorIs(t, store.DeleteUser(alice.ID), db.ErrLastGlobalAdministrator)

	// A second administrator lifts the refusal.
	root := createGlobalRoleTestUser(t, store, "root", true)
	require.NoError(t, store.DeleteUser(alice.ID))

	// Now root is the last one: the admin flag and the account are protected the same way.
	assert.ErrorIs(t, store.UpdateUser(db.UserWithPwd{User: db.User{ID: root.ID, Username: "root", Name: "root", Email: "root@e.st", Admin: false}}), db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.DeleteUser(root.ID), db.ErrLastGlobalAdministrator)
	require.NoError(t, store.UpdateUser(db.UserWithPwd{User: db.User{ID: root.ID, Username: "root", Name: "Root", Email: "root@e.st", Admin: true}}))
	user, err := store.GetUser(root.ID)
	require.NoError(t, err)
	assert.Equal(t, "Root", user.Name)
}

// An instance without any administrator (a fresh database) is not protected: there is nothing to keep.
func TestLastGlobalAdministrator_NoAdministratorsYet(t *testing.T) {
	store := InitConfigCreateTestStore()
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	require.NoError(t, store.UpdateUser(db.UserWithPwd{User: db.User{ID: alice.ID, Username: "alice", Name: "A", Email: "a@e.st"}}))
	require.NoError(t, store.DeleteUser(alice.ID))
}

func TestRequireAdministrator_CountsAdminsAndRoleHolders(t *testing.T) {
	store := InitConfigCreateTestStore()
	root := createGlobalRoleTestUser(t, store, "root", true)
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	manager := createGlobalRole(t, store, "role_manager", db.CanManageGlobalRoles)
	assignment := assignGlobally(t, store, alice.ID, manager)
	tx, err := store.Sql().Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	count, err := store.CountAdministrators(tx, db.AdministratorExclusion{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	count, err = store.CountAdministrators(tx, db.AdministratorExclusion{AdminOfUser: root.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "root without the flag, alice through the role")
	count, err = store.CountAdministrators(tx, db.AdministratorExclusion{AdminOfUser: root.ID, AssignmentID: assignment})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.ErrorIs(t, store.RequireAdministrator(tx, db.AdministratorExclusion{UserID: root.ID, RoleID: manager}), db.ErrLastGlobalAdministrator)
	require.NoError(t, store.RequireAdministrator(tx, db.AdministratorExclusion{UserID: root.ID}))
}

// Deleting a project role with its assignments does not touch global assignments: the tables are unrelated.
func TestDeleteRoleWithAssignments_LeavesGlobalRolesAlone(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	bob := createGlobalRoleTestUser(t, store, "bob", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	assignGlobally(t, store, alice.ID, auditor)
	role, err := store.CreateRole(db.Role{Slug: "auditor", Name: "Auditor"})
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	require.NoError(t, store.CreateProjectUser(project.ID, bob.ID, []int{role.ID}))

	removed, err := store.DeleteRoleWithAssignments(role.ID)
	require.NoError(t, err)
	assert.Equal(t, db.RoleImpact{Members: 1, Projects: 1}, removed.Impact())
	assert.Len(t, globalAssignmentsOf(t, store, alice.ID), 1)
}
