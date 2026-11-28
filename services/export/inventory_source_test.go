package export

import (
	"slices"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventorySourceExport_KeepsOwnership(t *testing.T) {
	src := sql.InitConfigCreateTestStore()
	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	key, err := src.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Name: "none", Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := src.CreateRepository(db.Repository{ProjectID: proj.ID, SSHKeyID: key.ID, Name: "infra", GitURL: "git@example.com:infra", GitBranch: "main"})
	require.NoError(t, err)
	inv, err := src.CreateInventory(db.Inventory{ProjectID: proj.ID, Name: "managed", Type: db.InventoryManaged})
	require.NoError(t, err)
	source, err := src.CreateInventorySource(db.InventorySource{
		ProjectID: proj.ID, InventoryID: inv.ID, Name: "static", Type: db.InventorySourceGit,
		RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "inventory/hosts.yml"}, Variables: "region: eu", VariableCount: 1,
		Status: db.InventorySourceFailed, LastError: "clone failed",
	})
	require.NoError(t, err)
	_, err = src.CreateInventorySource(db.InventorySource{ProjectID: proj.ID, InventoryID: inv.ID, Name: "edge", Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "edge.ini"}})
	require.NoError(t, err)
	_, err = src.CreateInventoryHost(db.InventoryHost{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web1", SourceID: &source.ID}, nil)
	require.NoError(t, err)
	_, err = src.CreateInventoryGroup(db.InventoryGroup{ProjectID: proj.ID, InventoryID: inv.ID, Name: "web", SourceID: &source.ID}, nil)
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
	repos, err := dst.GetRepositories(projects[0].ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, repos, 1)
	sources, err := dst.GetInventorySources(projects[0].ID, invs[0].ID)
	require.NoError(t, err)
	require.Len(t, sources, 2)
	slices.SortFunc(sources, func(a, b db.InventorySource) int { return a.ID - b.ID })
	assert.Equal(t, "static", sources[0].Name)
	assert.Equal(t, &repos[0].ID, sources[0].RepositoryID)
	assert.Equal(t, "inventory/hosts.yml", sources[0].Params["path"])
	assert.Equal(t, "region: eu", sources[0].Variables)
	assert.Equal(t, 1, sources[0].VariableCount)
	assert.Empty(t, sources[0].Status)
	assert.Empty(t, sources[0].LastError)
	hosts, _, err := dst.GetInventoryHosts(projects[0].ID, invs[0].ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, &sources[0].ID, hosts[0].SourceID)
	groups, err := dst.GetInventoryGroups(projects[0].ID, invs[0].ID)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, &sources[0].ID, groups[0].SourceID)
}
