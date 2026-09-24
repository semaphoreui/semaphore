package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_8_SeedsBuiltinRoles(t *testing.T) {
	before := "2.20.7"
	target := "2.20.8"
	store := InitConfigCreateTestStoreAt(&before)

	project, err := store.CreateProject(db.Project{Name: "project"})
	require.NoError(t, err)

	_, err = store.exec(
		"insert into `role` (slug, name, permissions, project_id) values (?, ?, ?, ?)",
		"existing_custom",
		"Existing custom",
		db.CanManageProjectResources,
		project.ID)
	require.NoError(t, err)

	// Stop at 2.20.8 so later migrations can add built-in roles without changing
	// this test of the historical 2.20.8 contract.
	require.NoError(t, db.Migrate(store, &target))

	var existing db.Role
	err = store.selectOne(&existing, "select * from `role` where slug=?", "existing_custom")
	require.NoError(t, err)
	assert.Equal(t, "Existing custom", existing.Name)
	assert.Equal(t, db.CanManageProjectResources, existing.Permissions)
	require.NotNil(t, existing.ProjectID)
	assert.Equal(t, project.ID, *existing.ProjectID)
	assert.False(t, existing.IsBuiltin)

	var builtinRoles []db.Role
	_, err = store.selectAll(&builtinRoles, "select * from `role` where is_builtin=true")
	require.NoError(t, err)

	expected := map[string]struct {
		name        string
		permissions db.ProjectUserPermission
	}{
		"owner":       {name: "Owner", permissions: 15},
		"manager":     {name: "Manager", permissions: 5},
		"task_runner": {name: "Task Runner", permissions: 1},
		"guest":       {name: "Guest", permissions: 0},
	}

	require.Len(t, builtinRoles, len(expected))
	for _, role := range builtinRoles {
		want, ok := expected[role.Slug]
		require.True(t, ok, "unexpected built-in role %q", role.Slug)
		assert.Equal(t, want.name, role.Name)
		assert.Equal(t, want.permissions, role.Permissions)
		assert.Nil(t, role.ProjectID)
		assert.True(t, role.IsBuiltin)
	}
}

func TestMigration_2_20_8_RejectsBuiltinSlugCollision(t *testing.T) {
	before := "2.20.7"
	target := "2.20.8"
	store := InitConfigCreateTestStoreAt(&before)

	_, err := store.exec(
		"insert into `role` (slug, name, permissions, project_id) values (?, ?, ?, ?)",
		"owner",
		"Conflicting owner",
		0,
		nil)
	require.NoError(t, err)

	err = db.Migrate(store, &target)
	require.Error(t, err)

	applied, err := store.IsMigrationApplied(db.Migration{Version: target})
	require.NoError(t, err)
	assert.False(t, applied)

	count, err := store.Sql().SelectInt("select count(*) from `role` where slug='owner'")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
