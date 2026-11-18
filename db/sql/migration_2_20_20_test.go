package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_20_AddsInventoryHostsAndVariables(t *testing.T) {
	before := "2.20.19"
	store := InitConfigCreateTestStoreAt(&before)
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	_, err = store.exec("insert into project__inventory (project_id, name, type, inventory) values (?, 'old', 'static', 'all')", project.ID)
	require.NoError(t, err)

	target := "2.20.20"
	require.NoError(t, db.Migrate(store, &target))
	_, err = store.exec("insert into project__inventory (project_id, name, type, inventory) values (?, 'from old node', 'static', 'all')", project.ID)
	require.NoError(t, err)

	inventories, err := store.GetInventories(project.ID, db.RetrieveQueryParams{}, nil)
	require.NoError(t, err)
	require.Len(t, inventories, 2)
	assert.Nil(t, inventories[0].Variables)
	assert.Nil(t, inventories[1].Variables)
	hosts, err := store.Sql().SelectInt("select count(*) from project__inventory_host")
	require.NoError(t, err)
	assert.Equal(t, int64(0), hosts)

	require.NoError(t, db.Rollback(store, "2.20.19"))
}
