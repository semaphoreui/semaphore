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

func createGlobalRole(t *testing.T, store *SqlDb, slug string, permissions db.GlobalPermission) db.Role {
	t.Helper()
	role, err := store.CreateRole(db.Role{Slug: slug, Name: slug, GlobalPermissions: permissions})
	require.NoError(t, err)
	return role
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

func TestCreateRole_StoresGlobalPermissions(t *testing.T) {
	store := InitConfigCreateTestStore()
	created := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit|db.CanManageGlobalUsers)

	role, err := store.GetGlobalRoleBySlug("auditor")
	require.NoError(t, err)
	assert.Equal(t, db.CanReadGlobalAudit|db.CanManageGlobalUsers, role.GlobalPermissions)
	assert.Equal(t, created.ID, role.ID)

	require.NoError(t, store.UpdateRole(db.Role{ID: role.ID, Name: "Auditor", GlobalPermissions: db.CanReadGlobalAudit}))
	role, err = store.GetGlobalRoleBySlug("auditor")
	require.NoError(t, err)
	assert.Equal(t, db.CanReadGlobalAudit, role.GlobalPermissions)
	assert.Equal(t, "Auditor", role.Name)
}

func TestGlobalRoleAssignmentSelect_JoinsTheRole(t *testing.T) {
	store := InitConfigCreateTestStore()
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	id := assignGlobally(t, store, user.ID, auditor.ID)

	assert.Equal(t, []db.GlobalRoleAssignment{{
		ID: id, UserID: user.ID, RoleID: auditor.ID, Username: "alice",
		RoleSlug: "auditor", RoleName: "auditor", GlobalPermissions: db.CanReadGlobalAudit,
	}}, globalAssignmentsOf(t, store, user.ID))
}

func TestDeleteUser_RemovesGlobalAssignments(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	assignGlobally(t, store, user.ID, auditor.ID)

	require.NoError(t, store.DeleteUser(user.ID))

	assigned, err := store.GetRoleAssignments(auditor.ID)
	require.NoError(t, err)
	assert.Empty(t, assigned.GlobalMembers)
}

// The last user who can manage global roles cannot lose that ability through this store: by losing the
// role, the role's permission, the admin flag or the account. The Pro store protects the assignment itself.
func TestLastGlobalAdministrator(t *testing.T) {
	store := InitConfigCreateTestStore()
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	manager := createGlobalRole(t, store, "role_manager", db.CanManageGlobalRoles|db.CanReadGlobalAudit)
	assignGlobally(t, store, alice.ID, manager.ID)

	assert.ErrorIs(t, store.DeleteUser(alice.ID), db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "x", GlobalPermissions: db.CanReadGlobalAudit}), db.ErrLastGlobalAdministrator)
	_, err := store.DeleteRoleWithAssignments(manager.ID)
	assert.ErrorIs(t, err, db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.DeleteRole(manager.ID), db.ErrInvalidOperation, "a plain delete is refused by the foreign key")
	// Keeping the permission and the name change is fine.
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanManageGlobalRoles}))

	// A second administrator lifts every refusal.
	root := createGlobalRoleTestUser(t, store, "root", true)
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanReadGlobalAudit}))
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanManageGlobalRoles}))
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
	assignment := assignGlobally(t, store, alice.ID, manager.ID)
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
	assert.ErrorIs(t, store.RequireAdministrator(tx, db.AdministratorExclusion{UserID: root.ID, RoleID: manager.ID}), db.ErrLastGlobalAdministrator)
	require.NoError(t, store.RequireAdministrator(tx, db.AdministratorExclusion{UserID: root.ID}))
}

func TestGetRoleAssignments_IncludesGlobalMembers(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	bob := createGlobalRoleTestUser(t, store, "bob", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	require.NoError(t, store.CreateProjectUser(project.ID, bob.ID, []int{auditor.ID}))
	assignGlobally(t, store, alice.ID, auditor.ID)

	assigned, err := store.GetRoleAssignments(auditor.ID)
	require.NoError(t, err)
	require.Len(t, assigned.GlobalMembers, 1)
	assert.Equal(t, alice.ID, assigned.GlobalMembers[0].UserID)
	assert.Equal(t, db.RoleImpact{Members: 2, Projects: 1}, assigned.Impact())

	removed, err := store.DeleteRoleWithAssignments(auditor.ID)
	require.NoError(t, err)
	assert.Equal(t, assigned.Impact(), removed.Impact())
	assert.Empty(t, globalAssignmentsOf(t, store, alice.ID))
	_, err = store.GetGlobalRoleBySlug("auditor")
	assert.ErrorIs(t, err, db.ErrNotFound)
}
