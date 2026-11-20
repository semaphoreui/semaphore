package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Hosts and inventories written before the migration get their counts.
func TestMigration_2_21_7_CountsExistingVariables(t *testing.T) {
	before := "2.21.6"
	store := InitConfigCreateTestStoreAt(&before)

	project, err := store.CreateProject(db.Project{Name: "old hosts"})
	require.NoError(t, err)
	inventoryID, err := store.insert("id",
		"insert into project__inventory (project_id, name, type, inventory, variables) values (?, 'managed', 'managed', '', 'x: 1')", project.ID)
	require.NoError(t, err)
	for name, variables := range map[string]string{"counted": "a: 1\nb: 2", "empty": "", "broken": "a: ["} {
		_, err = store.exec("insert into project__inventory_host (inventory_id, project_id, name, description, enabled, variables, created) "+
			"values (?, ?, ?, '', true, ?, ?)", inventoryID, project.ID, name, variables, project.Created)
		require.NoError(t, err)
	}

	target := "2.21.7"
	require.NoError(t, db.Migrate(store, &target))

	inventory, err := store.GetInventory(project.ID, inventoryID)
	require.NoError(t, err)
	assert.Equal(t, 1, inventory.VariableCount)
	hosts, _, err := store.GetInventoryHosts(project.ID, inventoryID, db.InventoryHostListParams{})
	require.NoError(t, err)
	counts := map[string]int{}
	for _, host := range hosts {
		counts[host.Name] = host.VariableCount
	}
	assert.Equal(t, map[string]int{"counted": 2, "empty": 0, "broken": 0}, counts)
}
