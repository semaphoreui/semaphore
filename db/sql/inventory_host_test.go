package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestManagedInventory(t *testing.T, store *SqlDb, projectID int, name string) db.Inventory {
	t.Helper()
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: name, Type: db.InventoryManaged})
	require.NoError(t, err)
	return inventory
}

func TestInventoryHosts_CRUD(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")

	created, err := store.CreateInventoryHost(db.InventoryHost{
		ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true,
	})
	require.NoError(t, err)
	got, err := store.GetInventoryHost(project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "web1", got.Name)
	assert.True(t, got.Enabled)

	got.Name = "web2"
	got.Description = "second"
	got.Enabled = false
	got.Variables = "ansible_port: 2222"
	require.NoError(t, store.UpdateInventoryHost(got))
	updated, err := store.GetInventoryHost(project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "web2", updated.Name)
	assert.Equal(t, "second", updated.Description)
	assert.True(t, updated.Enabled)
	assert.Equal(t, "ansible_port: 2222", updated.Variables)

	changed, err := store.SetInventoryHostEnabled(project.ID, inventory.ID, created.ID, false)
	require.NoError(t, err)
	assert.True(t, changed)
	changed, err = store.SetInventoryHostEnabled(project.ID, inventory.ID, created.ID, false)
	require.NoError(t, err)
	assert.False(t, changed)
	updated, err = store.GetInventoryHost(project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.False(t, updated.Enabled)
	assert.Equal(t, "web2", updated.Name)

	require.NoError(t, store.DeleteInventoryHost(project.ID, inventory.ID, created.ID))
	_, err = store.GetInventoryHost(project.ID, inventory.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestInventoryHosts_NameIsUniqueWithinTheInventory(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	first := createTestManagedInventory(t, store, project.ID, "first")
	second := createTestManagedInventory(t, store, project.ID, "second")
	newHost := func(inventoryID int, name string) (db.InventoryHost, error) {
		return store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventoryID, Name: name, Enabled: true})
	}

	_, err = newHost(first.ID, "web1")
	require.NoError(t, err)
	_, err = newHost(first.ID, "web1")
	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.ErrorContains(t, err, "a host with this name already exists in the inventory")
	_, err = newHost(second.ID, "web1")
	assert.NoError(t, err)
	_, err = newHost(first.ID, "Web1")
	assert.NoError(t, err)

	web2, err := newHost(first.ID, "web2")
	require.NoError(t, err)
	web2.Name = "web1"
	err = store.UpdateInventoryHost(web2)
	require.ErrorAs(t, err, &validationErr)
}

func TestInventoryHosts_ListPagesAndSearch(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	for _, name := range []string{"web1", "db1", "cache1", "app1", "mq1"} {
		_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: name, Enabled: true})
		require.NoError(t, err)
	}

	page, total, err := store.GetInventoryHosts(project.ID, inventory.ID, db.InventoryHostListParams{Count: 2, Offset: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, page, 2)
	assert.Equal(t, "db1", page[0].Name)
	assert.Equal(t, "mq1", page[1].Name)

	found, total, err := store.GetInventoryHosts(project.ID, inventory.ID, db.InventoryHostListParams{Search: "WEB"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, found, 1)
	assert.Equal(t, "web1", found[0].Name)
}

func TestInventoryHosts_AreScopedToTheInventory(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	otherInventory := createTestManagedInventory(t, store, project.ID, "other")
	host, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true})
	require.NoError(t, err)

	_, err = store.GetInventoryHost(project.ID, otherInventory.ID, host.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetInventoryHost(other.ID, inventory.ID, host.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
	assert.ErrorIs(t, store.DeleteInventoryHost(project.ID, otherInventory.ID, host.ID), db.ErrNotFound)
	assert.ErrorIs(t, store.DeleteInventoryHost(other.ID, inventory.ID, host.ID), db.ErrNotFound)
}

func TestInventoryHosts_UpdateMissingIsNotFound(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	otherInventory := createTestManagedInventory(t, store, project.ID, "other")
	host, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true})
	require.NoError(t, err)

	tests := []struct {
		name string
		host db.InventoryHost
	}{
		{"missing host", db.InventoryHost{ID: host.ID + 100, ProjectID: project.ID, InventoryID: inventory.ID, Name: "web9"}},
		{"host of another inventory", db.InventoryHost{ID: host.ID, ProjectID: project.ID, InventoryID: otherInventory.ID, Name: "web1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, store.UpdateInventoryHost(tt.host), db.ErrNotFound)
		})
	}
}

func TestGetEnabledInventoryHosts(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "on", Enabled: true})
	require.NoError(t, err)
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "off", Enabled: false})
	require.NoError(t, err)

	hosts, err := store.GetEnabledInventoryHosts(project.ID, inventory.ID)

	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, "on", hosts[0].Name)
}

func TestDeleteInventory_DeletesItsHosts(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true})
	require.NoError(t, err)

	require.NoError(t, store.DeleteInventory(project.ID, inventory.ID))

	count, err := store.Sql().SelectInt("select count(*) from project__inventory_host")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestInventory_VariablesRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "managed", Type: db.InventoryManaged, Variables: new("a: 1")})
	require.NoError(t, err)
	got, err := store.GetInventory(project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, "a: 1", got.VariablesText())

	got.Name = "renamed"
	got.Variables = new("b: 2")
	require.NoError(t, store.UpdateInventory(got))
	got, err = store.GetInventory(project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)
	assert.Equal(t, "a: 1", got.VariablesText())

	require.NoError(t, store.SetInventoryVariables(project.ID, inventory.ID, "b: 2"))
	got, err = store.GetInventory(project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, "b: 2", got.VariablesText())
}
