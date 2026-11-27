package export

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportKeepsInventoryGroups(t *testing.T) {
	src := sql.InitConfigCreateTestStore()
	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	inv, err := src.CreateInventory(db.Inventory{ProjectID: proj.ID, Name: "managed", Type: db.InventoryManaged})
	require.NoError(t, err)
	web1, err := src.CreateInventoryHost(db.InventoryHost{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web1"}, nil)
	require.NoError(t, err)
	web2, err := src.CreateInventoryHost(db.InventoryHost{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web2"}, nil)
	require.NoError(t, err)
	web, err := src.CreateInventoryGroup(db.InventoryGroup{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web"}, nil)
	require.NoError(t, err)
	edge, err := src.CreateInventoryGroup(db.InventoryGroup{ProjectID: proj.ID, InventoryID: inv.ID, Name: "edge", Description: "front", Variables: "tier: front"}, []int{web.ID})
	require.NoError(t, err)
	_, _, err = src.ChangeInventoryHostGroups(proj.ID, inv.ID, []int{web1.ID}, []int{edge.ID}, nil, false)
	require.NoError(t, err)
	_, _, err = src.ChangeInventoryHostGroups(proj.ID, inv.ID, []int{web2.ID}, []int{web.ID}, nil, false)
	require.NoError(t, err)

	dst := sql.InitConfigCreateTestStore()
	chain := InitProjectExporters(NewKeyMapper(), true, false)
	require.NoError(t, chain.Load(src))
	require.NoError(t, chain.Restore(dst, 10))

	projects, err := dst.GetAllProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	invs, err := dst.GetInventories(projects[0].ID, db.RetrieveQueryParams{}, []db.InventoryType{})
	require.NoError(t, err)
	require.Len(t, invs, 1)
	groups, err := dst.GetInventoryGroups(projects[0].ID, invs[0].ID)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "edge", groups[0].Name)
	assert.Equal(t, "front", groups[0].Description)
	assert.Equal(t, "tier: front", groups[0].Variables)
	assert.Equal(t, []int{groups[1].ID}, groups[0].Children)
	assert.Equal(t, "web", groups[1].Name)
	hosts, _, err := dst.GetInventoryHosts(projects[0].ID, invs[0].ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	assert.Equal(t, []int{groups[0].ID}, hosts[0].Groups)
	assert.Equal(t, []int{groups[1].ID}, hosts[1].Groups)
}
