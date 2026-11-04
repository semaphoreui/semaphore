package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_17_LiftsMasksAndAddsWorkflowGrants(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)
	for _, r := range []struct {
		slug string
		mask int
	}{{"ops", 5}, {"runner2", 1}, {"nobody", 0}} {
		_, err := store.exec("insert into `role` (slug, name, permissions, project_id, scope, builtin) values (?, ?, ?, null, 0, false)", r.slug, r.slug, r.mask)
		require.NoError(t, err)
	}

	target := "2.20.17"
	require.NoError(t, db.Migrate(store, &target))

	want := map[string]db.ProjectUserPermission{
		"owner": 127, "manager": 117, "task_runner": 81, "guest": 0,
		"ops": 117, "runner2": 81, "nobody": 0,
	}
	for slug, mask := range want {
		role, err := store.GetSharedRoleBySlug(slug)
		require.NoError(t, err)
		assert.Equal(t, mask, role.Permissions, slug)
	}
	n, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from project__workflow_role"))
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	require.NoError(t, db.Rollback(store, "2.20.16"))
	role, err := store.GetSharedRoleBySlug("ops")
	require.NoError(t, err)
	assert.Equal(t, db.ProjectUserPermission(5), role.Permissions, "rollback keeps the old bits only")
}
