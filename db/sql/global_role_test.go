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

func TestGlobalRoleAssignments_Lifecycle(t *testing.T) {
	store := InitConfigCreateTestStore()
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	operator := createGlobalRole(t, store, "operator", db.CanManageGlobalSystem)

	permissions, err := store.GetEffectiveGlobalPermissions(user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.GlobalPermission(0), permissions)

	first, err := store.CreateGlobalRoleAssignment(user.ID, auditor.ID)
	require.NoError(t, err)
	assert.Equal(t, db.GlobalRoleAssignment{
		ID: first.ID, UserID: user.ID, RoleID: auditor.ID,
		RoleSlug: "auditor", RoleName: "auditor", GlobalPermissions: db.CanReadGlobalAudit,
	}, first)
	_, err = store.CreateGlobalRoleAssignment(user.ID, operator.ID)
	require.NoError(t, err)

	assignments, err := store.GetGlobalRoleAssignments(user.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"auditor", "operator"}, []string{assignments[0].RoleSlug, assignments[1].RoleSlug})

	permissions, err = store.GetEffectiveGlobalPermissions(user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanReadGlobalAudit|db.CanManageGlobalSystem, permissions)

	require.NoError(t, store.DeleteGlobalRoleAssignment(user.ID, first.ID))
	permissions, err = store.GetEffectiveGlobalPermissions(user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.CanManageGlobalSystem, permissions)

	assert.ErrorIs(t, store.DeleteGlobalRoleAssignment(user.ID, first.ID), db.ErrNotFound)
	_, err = store.GetGlobalRoleAssignment(user.ID, first.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestGlobalRoleAssignments_AdminHoldsEverything(t *testing.T) {
	store := InitConfigCreateTestStore()
	admin := createGlobalRoleTestUser(t, store, "root", true)

	permissions, err := store.GetEffectiveGlobalPermissions(admin.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AllGlobalPermissions, permissions)
}

func TestCreateGlobalRoleAssignment_Refusals(t *testing.T) {
	store := InitConfigCreateTestStore()
	user := createGlobalRoleTestUser(t, store, "alice", false)
	other := createGlobalRoleTestUser(t, store, "bob", false)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	projectRole, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", ProjectID: &project.ID})
	require.NoError(t, err)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)

	_, err = store.CreateGlobalRoleAssignment(user.ID, projectRole.ID)
	assert.ErrorContains(t, err, "Global role not found")
	owner, err := store.GetGlobalRoleBySlug("owner")
	require.NoError(t, err)
	_, err = store.CreateGlobalRoleAssignment(user.ID, owner.ID)
	assert.ErrorContains(t, err, "Global role not found", "a built-in role grants nothing globally")
	_, err = store.CreateGlobalRoleAssignment(user.ID, 99999)
	assert.ErrorContains(t, err, "Global role not found")
	_, err = store.CreateGlobalRoleAssignment(99999, auditor.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.CreateGlobalRoleAssignment(user.ID, auditor.ID)
	require.NoError(t, err)
	_, err = store.CreateGlobalRoleAssignment(user.ID, auditor.ID)
	assert.ErrorContains(t, err, "already holds")

	// Another user's assignment is not visible through this user.
	assignment, err := store.CreateGlobalRoleAssignment(other.ID, auditor.ID)
	require.NoError(t, err)
	assert.ErrorIs(t, store.DeleteGlobalRoleAssignment(user.ID, assignment.ID), db.ErrNotFound)
}

func TestDeleteUser_RemovesGlobalAssignments(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	user := createGlobalRoleTestUser(t, store, "alice", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	_, err := store.CreateGlobalRoleAssignment(user.ID, auditor.ID)
	require.NoError(t, err)

	require.NoError(t, store.DeleteUser(user.ID))

	assigned, err := store.GetRoleAssignments(auditor.ID)
	require.NoError(t, err)
	assert.Empty(t, assigned.GlobalMembers)
}

// The last user who can manage global roles cannot lose that ability: by losing the assignment, the role, the
// role's permission, the admin flag or the account.
func TestLastGlobalAdministrator(t *testing.T) {
	store := InitConfigCreateTestStore()
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	manager := createGlobalRole(t, store, "role_manager", db.CanManageGlobalRoles|db.CanReadGlobalAudit)
	assignment, err := store.CreateGlobalRoleAssignment(alice.ID, manager.ID)
	require.NoError(t, err)

	assert.ErrorIs(t, store.DeleteGlobalRoleAssignment(alice.ID, assignment.ID), db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.DeleteUser(alice.ID), db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "x", GlobalPermissions: db.CanReadGlobalAudit}), db.ErrLastGlobalAdministrator)
	_, err = store.DeleteRoleWithAssignments(manager.ID)
	assert.ErrorIs(t, err, db.ErrLastGlobalAdministrator)
	assert.ErrorIs(t, store.DeleteRole(manager.ID), db.ErrInvalidOperation, "a plain delete is refused by the foreign key")
	// Keeping the permission and the name change is fine.
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanManageGlobalRoles}))

	// A second administrator lifts every refusal.
	root := createGlobalRoleTestUser(t, store, "root", true)
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanReadGlobalAudit}))
	require.NoError(t, store.UpdateRole(db.Role{ID: manager.ID, Name: "Renamed", GlobalPermissions: db.CanManageGlobalRoles}))
	require.NoError(t, store.DeleteGlobalRoleAssignment(alice.ID, assignment.ID))
	_, err = store.CreateGlobalRoleAssignment(alice.ID, manager.ID)
	require.NoError(t, err)

	// Now root is the last one: the admin flag and the account are protected the same way.
	require.NoError(t, store.DeleteUser(alice.ID))
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

func TestGetRoleAssignments_IncludesGlobalMembers(t *testing.T) {
	store := InitConfigCreateTestStore()
	createGlobalRoleTestUser(t, store, "root", true)
	alice := createGlobalRoleTestUser(t, store, "alice", false)
	bob := createGlobalRoleTestUser(t, store, "bob", false)
	auditor := createGlobalRole(t, store, "auditor", db.CanReadGlobalAudit)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	require.NoError(t, store.CreateProjectUser(project.ID, bob.ID, []int{auditor.ID}))
	_, err = store.CreateGlobalRoleAssignment(alice.ID, auditor.ID)
	require.NoError(t, err)

	assigned, err := store.GetRoleAssignments(auditor.ID)
	require.NoError(t, err)
	require.Len(t, assigned.GlobalMembers, 1)
	assert.Equal(t, alice.ID, assigned.GlobalMembers[0].UserID)
	assert.Equal(t, db.RoleImpact{Members: 1, GlobalMembers: 1, Projects: 1}, assigned.Impact())

	removed, err := store.DeleteRoleWithAssignments(auditor.ID)
	require.NoError(t, err)
	assert.Equal(t, assigned.Impact(), removed.Impact())
	assignments, err := store.GetGlobalRoleAssignments(alice.ID)
	require.NoError(t, err)
	assert.Empty(t, assignments)
	_, err = store.GetGlobalRoleBySlug("auditor")
	assert.ErrorIs(t, err, db.ErrNotFound)
}
