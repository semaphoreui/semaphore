package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_21_8_AddsLastRunOfInventoryHost(t *testing.T) {
	before := "2.21.7"
	store := InitConfigCreateTestStoreAt(&before)
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	_, err = store.CreateInventoryHost(db.InventoryHost{
		ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true,
	}, nil)
	require.NoError(t, err)

	target := "2.21.8"
	require.NoError(t, db.Migrate(store, &target))

	count, err := store.Sql().SelectInt(
		"select count(*) from project__inventory_host where last_status is null and last_run is null and last_task_id is null and host_id is null")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	_, err = store.Sql().SelectInt("select count(effective_inventory_id) from task")
	require.NoError(t, err)

	require.NoError(t, db.Rollback(store, "2.21.7"))

	_, err = store.Sql().SelectInt("select count(last_status) from project__inventory_host")
	assert.Error(t, err)
	_, err = store.Sql().SelectInt("select count(effective_inventory_id) from task")
	assert.Error(t, err)
}
