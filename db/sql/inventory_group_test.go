package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inventoryGroupFixture struct {
	store     *SqlDb
	project   db.Project
	inventory db.Inventory
}

func newInventoryGroupFixture(t *testing.T) inventoryGroupFixture {
	t.Helper()
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)
	return inventoryGroupFixture{store: store, project: project, inventory: createTestManagedInventory(t, store, project.ID, "managed")}
}

func (f inventoryGroupFixture) group(t *testing.T, name string, children ...int) db.InventoryGroup {
	t.Helper()
	group, err := f.store.CreateInventoryGroup(db.InventoryGroup{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: name}, children)
	require.NoError(t, err)
	return group
}

func (f inventoryGroupFixture) host(t *testing.T, name string, groups ...int) db.InventoryHost {
	t.Helper()
	host, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: name, Enabled: true}, groups)
	require.NoError(t, err)
	return host
}

func (f inventoryGroupFixture) linkCount(t *testing.T) int64 {
	t.Helper()
	count, err := f.store.Sql().SelectInt("select count(*) from project__inventory_group_host")
	require.NoError(t, err)
	return count
}

func TestInventoryGroups_CreateAndNameIsUnique(t *testing.T) {
	f := newInventoryGroupFixture(t)
	web := f.group(t, "web")
	parent := f.group(t, "parent", web.ID)

	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, "parent", got.Name)
	assert.Equal(t, []int{web.ID}, got.Children)

	_, err = f.store.CreateInventoryGroup(db.InventoryGroup{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "web"}, nil)
	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.ErrorContains(t, err, "a group with this name already exists in the inventory")
}

func TestChangeInventoryHostGroups_ReportsOnlyChangedLinks(t *testing.T) {
	f := newInventoryGroupFixture(t)
	web, db1 := f.group(t, "web"), f.group(t, "db")
	h1, h2 := f.host(t, "h1"), f.host(t, "h2")

	added, removed, err := f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{h1.ID, h2.ID}, []int{web.ID, db1.ID}, nil)
	require.NoError(t, err)
	assert.Equal(t, map[int][]int{web.ID: {h1.ID, h2.ID}, db1.ID: {h1.ID, h2.ID}}, added)
	assert.Empty(t, removed)

	added, removed, err = f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{h1.ID, h2.ID}, []int{web.ID, db1.ID}, nil)
	require.NoError(t, err)
	assert.Empty(t, added)
	assert.Empty(t, removed)

	got, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, h1.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{web.ID, db1.ID}, got.Groups)
}

func TestChangeInventoryHostGroups_ForeignHostWritesNothing(t *testing.T) {
	f := newInventoryGroupFixture(t)
	web := f.group(t, "web")
	h1 := f.host(t, "h1")
	other := createTestManagedInventory(t, f.store, f.project.ID, "other")
	foreign, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: other.ID, Name: "h1", Enabled: true}, nil)
	require.NoError(t, err)

	_, _, err = f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{h1.ID, foreign.ID}, []int{web.ID}, nil)

	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, int64(0), f.linkCount(t))
}

func TestCreateInventoryHost_UnknownGroupCreatesNothing(t *testing.T) {
	f := newInventoryGroupFixture(t)
	web := f.group(t, "web")

	_, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "h1", Enabled: true}, []int{web.ID, web.ID + 100})

	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	_, total, err := f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Equal(t, int64(0), f.linkCount(t))
}

func TestGetInventoryGroups_DiamondCountsHostOnce(t *testing.T) {
	f := newInventoryGroupFixture(t)
	shared := f.group(t, "web-shared")
	eu := f.group(t, "web-eu", shared.ID)
	us := f.group(t, "web-us", shared.ID)
	web := f.group(t, "web", eu.ID, us.ID)
	f.host(t, "h1", shared.ID, eu.ID)
	f.host(t, "h2", us.ID)

	groups, err := f.store.GetInventoryGroups(f.project.ID, f.inventory.ID)

	require.NoError(t, err)
	require.Len(t, groups, 4)
	assert.Equal(t, "web", groups[0].Name)
	assert.Equal(t, []int{eu.ID, us.ID}, groups[0].Children)
	assert.Equal(t, 2, groups[0].HostCount)
	assert.Equal(t, web.ID, groups[0].ID)
	assert.Equal(t, "web-shared", groups[2].Name)
	assert.Equal(t, 1, groups[2].HostCount)
}

func TestGetInventoryGroup_CountsAsTheList(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, f inventoryGroupFixture)
	}{
		{"diamond", func(t *testing.T, f inventoryGroupFixture) {
			shared := f.group(t, "web-shared")
			eu := f.group(t, "web-eu", shared.ID)
			us := f.group(t, "web-us", shared.ID)
			f.group(t, "web", eu.ID, us.ID)
			f.host(t, "h1", shared.ID, eu.ID)
			f.host(t, "h2", us.ID)
		}},
		{"cycle in the table", func(t *testing.T, f inventoryGroupFixture) {
			a, b := f.group(t, "a"), f.group(t, "b")
			f.host(t, "h1", a.ID)
			_, err := f.store.exec("insert into project__inventory_group_child (parent_id, child_id) values (?, ?), (?, ?)", a.ID, b.ID, b.ID, a.ID)
			require.NoError(t, err)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newInventoryGroupFixture(t)
			tt.setup(t, f)

			groups, err := f.store.GetInventoryGroups(f.project.ID, f.inventory.ID)

			require.NoError(t, err)
			for _, want := range groups {
				got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, want.ID)
				require.NoError(t, err)
				assert.Positive(t, got.HostCount, want.Name)
				assert.Equal(t, want.HostCount, got.HostCount, want.Name)
				assert.Equal(t, want.Children, got.Children, want.Name)
			}
		})
	}
}

func TestUpdateInventoryGroup_RejectsCycle(t *testing.T) {
	f := newInventoryGroupFixture(t)
	a, b := f.group(t, "a"), f.group(t, "b")

	added, removed, err := f.store.UpdateInventoryGroup(a, []int{b.ID}, nil)
	require.NoError(t, err)
	assert.Equal(t, []int{b.ID}, added)
	assert.Empty(t, removed)

	b.Name = "renamed"
	_, _, err = f.store.UpdateInventoryGroup(b, []int{a.ID}, nil)
	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.ErrorContains(t, err, "a group can not be its own ancestor")
	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, b.ID)
	require.NoError(t, err)
	assert.Equal(t, "b", got.Name)
	assert.Empty(t, got.Children)
}

func TestUpdateInventoryGroup_CycleInTheTableAllowsOtherChanges(t *testing.T) {
	f := newInventoryGroupFixture(t)
	a, b := f.group(t, "a"), f.group(t, "b")
	_, err := f.store.exec("insert into project__inventory_group_child (parent_id, child_id) values (?, ?), (?, ?)", a.ID, b.ID, b.ID, a.ID)
	require.NoError(t, err)

	a.Name = "renamed"
	_, _, err = f.store.UpdateInventoryGroup(a, nil, nil)
	require.NoError(t, err)
	c, d := f.group(t, "c"), f.group(t, "d")
	added, _, err := f.store.UpdateInventoryGroup(c, []int{d.ID}, nil)
	require.NoError(t, err)
	assert.Equal(t, []int{d.ID}, added)
	_, removed, err := f.store.UpdateInventoryGroup(a, nil, []int{b.ID})
	require.NoError(t, err)
	assert.Equal(t, []int{b.ID}, removed)
}

func TestGetInventoryGroups_ReturnsWithCycleInTheTable(t *testing.T) {
	f := newInventoryGroupFixture(t)
	a, b := f.group(t, "a"), f.group(t, "b")
	f.host(t, "h1", a.ID)
	_, err := f.store.exec("insert into project__inventory_group_child (parent_id, child_id) values (?, ?), (?, ?)", a.ID, b.ID, b.ID, a.ID)
	require.NoError(t, err)

	groups, err := f.store.GetInventoryGroups(f.project.ID, f.inventory.ID)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, 1, groups[0].HostCount)
	assert.Equal(t, 1, groups[1].HostCount)
}

func TestGetInventoryHosts_FiltersByGroup(t *testing.T) {
	f := newInventoryGroupFixture(t)
	child := f.group(t, "child")
	parent := f.group(t, "parent", child.ID)
	f.host(t, "in-child", child.ID)
	f.host(t, "alone")

	hosts, total, err := f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{GroupID: parent.ID, Count: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, hosts, 1)
	assert.Equal(t, "in-child", hosts[0].Name)
	assert.Equal(t, []int{child.ID}, hosts[0].Groups)

	hosts, total, err = f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{Ungrouped: true})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, hosts, 1)
	assert.Equal(t, "alone", hosts[0].Name)

	_, _, err = f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{GroupID: parent.ID + 100})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestDeleteInventoryGroup_KeepsHostsAndChildren(t *testing.T) {
	f := newInventoryGroupFixture(t)
	child := f.group(t, "child")
	parent := f.group(t, "parent", child.ID)
	host := f.host(t, "h1", parent.ID, child.ID)

	require.NoError(t, f.store.DeleteInventoryGroup(f.project.ID, f.inventory.ID, parent.ID))
	_, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, host.ID)
	require.NoError(t, err)
	_, err = f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), f.linkCount(t))

	require.NoError(t, f.store.DeleteInventoryHost(f.project.ID, f.inventory.ID, host.ID))
	assert.Equal(t, int64(0), f.linkCount(t))
}

func TestInventoryGroups_StoreTheVariableCount(t *testing.T) {
	f := newInventoryGroupFixture(t)

	group, err := f.store.CreateInventoryGroup(db.InventoryGroup{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "web", Variables: "a: 1\nb: 2", VariableCount: 99}, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, group.VariableCount)
	groups, err := f.store.GetInventoryGroups(f.project.ID, f.inventory.ID)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, 2, groups[0].VariableCount)

	group.Variables = "a: 1"
	_, _, err = f.store.UpdateInventoryGroup(group, nil, nil)
	require.NoError(t, err)
	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.VariableCount)
}
