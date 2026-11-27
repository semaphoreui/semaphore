package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_16_AddsGlobalRolesAndAssignments(t *testing.T) {
	before := "2.20.13"
	store := InitConfigCreateTestStoreAt(&before)
	_, err := store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values ('ops', 'Ops', 1, null, 0, false)")
	require.NoError(t, err)

	target := "2.20.16"
	require.NoError(t, db.Migrate(store, &target))

	role, err := store.GetSharedRoleBySlug("ops")
	require.NoError(t, err)
	assert.Equal(t, db.CanRunProjectTasks, role.Permissions, "project roles are untouched")
	revision, err := store.Sql().SelectInt(store.PrepareQuery("select revision from global_role_state where id=1"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), revision)

	auditor, err := store.insert("id", "insert into global_role (slug, name, permissions) values (?, ?, ?)", "auditor", "auditor", db.CanReadGlobalAudit)
	require.NoError(t, err)
	_, err = store.insert("id", "insert into global_role (slug, name, permissions) values (?, ?, ?)", "auditor", "Again", 0)
	assert.Error(t, err, "a global role slug is unique")
	user, err := store.CreateUserWithoutPassword(db.User{Username: "a", Name: "A", Email: "a@e.st"})
	require.NoError(t, err)
	_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", user.ID, auditor)
	require.NoError(t, err)
	_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", user.ID, auditor)
	assert.Error(t, err, "a user holds a role once")
	_, err = store.exec("insert into user__global_role (user_id, role_id) values (?, ?)", user.ID, role.ID+1000)
	assert.Error(t, err, "an assignment needs a global role")
	held, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from user__global_role where user_id=?"), user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), held)
}
