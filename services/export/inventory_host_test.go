package export

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportKeepsInventoryHosts(t *testing.T) {
	src := sql.InitConfigCreateTestStore()
	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	inv, err := src.CreateInventory(db.Inventory{ProjectID: proj.ID, Name: "managed", Type: db.InventoryManaged, Variables: new("a: 1")})
	require.NoError(t, err)
	_, err = src.CreateInventoryHost(db.InventoryHost{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web1", Enabled: true, Variables: "port: 80"})
	require.NoError(t, err)
	_, err = src.CreateInventoryHost(db.InventoryHost{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web2", Description: "spare"})
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
	assert.Equal(t, "a: 1", invs[0].VariablesText())
	hosts, _, err := dst.GetInventoryHosts(projects[0].ID, invs[0].ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	assert.Equal(t, "web1", hosts[0].Name)
	assert.True(t, hosts[0].Enabled)
	assert.Equal(t, "port: 80", hosts[0].Variables)
	assert.Equal(t, "spare", hosts[1].Description)
}
