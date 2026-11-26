package sql

import (
	"testing"
	"time"

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
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, []int{}, created.Groups)
	got, err := store.GetInventoryHost(project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "web1", got.Name)
	assert.True(t, got.Enabled)

	got.Name = "web2"
	got.Description = "second"
	got.Enabled = false
	got.Variables = "ansible_port: 2222"
	_, _, err = store.UpdateInventoryHost(got, nil, nil)
	require.NoError(t, err)
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
		return store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventoryID, Name: name, Enabled: true}, nil)
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
	_, _, err = store.UpdateInventoryHost(web2, nil, nil)
	require.ErrorAs(t, err, &validationErr)
}

func TestInventoryHosts_ListPagesAndSearch(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	for _, name := range []string{"web1", "db1", "cache1", "app1", "mq1"} {
		_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: name, Enabled: true}, nil)
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
	host, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true}, nil)
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
	host, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true}, nil)
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
			_, _, err := store.UpdateInventoryHost(tt.host, nil, nil)
			assert.ErrorIs(t, err, db.ErrNotFound)
		})
	}
}

func TestGetEnabledInventoryHosts(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "on", Enabled: true}, nil)
	require.NoError(t, err)
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "off", Enabled: false}, nil)
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
	_, err = store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true}, nil)
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

func TestInventoryHosts_StoreTheVariableCount(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "managed", Type: db.InventoryManaged})
	require.NoError(t, err)

	host, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Variables: "a: 1\nb: 2"}, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, host.VariableCount)
	hosts, _, err := store.GetInventoryHosts(project.ID, inventory.ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, 2, hosts[0].VariableCount)

	host.Variables = "a: 1"
	_, _, err = store.UpdateInventoryHost(host, nil, nil)
	require.NoError(t, err)
	got, err := store.GetInventoryHost(project.ID, inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.VariableCount)

	host.Variables = ""
	_, _, err = store.UpdateInventoryHost(host, nil, nil)
	require.NoError(t, err)
	got, err = store.GetInventoryHost(project.ID, inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, got.VariableCount)

	other, err := store.CreateInventoryHost(db.InventoryHost{ProjectID: project.ID, InventoryID: inventory.ID, Name: "web2", VariableCount: 99}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, other.VariableCount)
	got, err = store.GetInventoryHost(project.ID, inventory.ID, other.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, got.VariableCount)
}

func TestInventory_StoresTheVariableCount(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)

	inventory, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "managed", Type: db.InventoryManaged, Variables: new("a: 1\nb: 2")})
	require.NoError(t, err)
	assert.Equal(t, 2, inventory.VariableCount)
	inventories, err := store.GetInventories(project.ID, db.RetrieveQueryParams{}, nil)
	require.NoError(t, err)
	require.Len(t, inventories, 1)
	assert.Equal(t, 2, inventories[0].VariableCount)

	require.NoError(t, store.SetInventoryVariables(project.ID, inventory.ID, "a: 1"))
	got, err := store.GetInventory(project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.VariableCount)

	got.Name = "renamed"
	require.NoError(t, store.UpdateInventory(got))
	got, err = store.GetInventory(project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.VariableCount)

	static, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "static", Type: db.InventoryStatic, Inventory: "localhost"})
	require.NoError(t, err)
	assert.Equal(t, 0, static.VariableCount)
}

func TestUpdateInventoryHost_KeepsLastRun(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	inventory := createTestManagedInventory(t, store, project.ID, "managed")
	host, err := store.CreateInventoryHost(db.InventoryHost{
		ProjectID: project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true,
	}, nil)
	require.NoError(t, err)
	lastRun := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	_, err = store.Sql().Exec(
		"update project__inventory_host set last_status = ?, last_run = ?, last_task_id = ?, host_id = ? where id = ?",
		"ok", lastRun, 9, 7, host.ID)
	require.NoError(t, err)

	host.Name = "web2"
	host.Variables = "a: 1"
	_, _, err = store.UpdateInventoryHost(host, nil, nil)
	require.NoError(t, err)
	_, err = store.SetInventoryHostEnabled(project.ID, inventory.ID, host.ID, false)
	require.NoError(t, err)

	got, err := store.GetInventoryHost(project.ID, inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, "web2", got.Name)
	require.NotNil(t, got.LastStatus)
	assert.Equal(t, "ok", *got.LastStatus)
	require.NotNil(t, got.LastRun)
	assert.True(t, lastRun.Equal(*got.LastRun))
	require.NotNil(t, got.LastTaskID)
	assert.Equal(t, 9, *got.LastTaskID)
	require.NotNil(t, got.HostID)
	assert.Equal(t, 7, *got.HostID)
}
