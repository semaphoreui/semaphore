package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_14_AddsGlobalPermissionsAndAssignments(t *testing.T) {
	before := "2.20.13"
	store := InitConfigCreateTestStoreAt(&before)
	_, err := store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ops', 'Ops', 1, null, 0, false)")
	require.NoError(t, err)

	target := "2.20.14"
	require.NoError(t, db.Migrate(store, &target))

	role, err := store.GetGlobalRoleBySlug("ops")
	require.NoError(t, err)
	assert.Equal(t, db.GlobalPermission(0), role.GlobalPermissions, "existing roles get no global permissions")
	revision, err := store.Sql().SelectInt(store.PrepareQuery("select revision from global_role_state where id=1"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), revision)
	user, err := store.CreateUserWithoutPassword(db.User{Username: "a", Name: "A", Email: "a@e.st"})
	require.NoError(t, err)
	_, err = store.CreateGlobalRoleAssignment(user.ID, role.ID)
	require.NoError(t, err)
}
