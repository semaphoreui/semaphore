package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMcpTestInventory(t *testing.T, store *SqlDb) (projectID int, inventoryID int) {
	t.Helper()

	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "mcp", Type: db.InventoryMcp})
	require.NoError(t, err)

	return project.ID, inv.ID
}

func mcpSecret(s string) *string {
	return &s
}

func mcpTokenCreate(secret string) *db.McpTokenChange {
	return &db.McpTokenChange{Operation: db.McpTokenCreate, Key: db.AccessKey{Secret: mcpSecret(secret)}}
}

func serverByName(t *testing.T, servers []db.McpServer, name string) db.McpServer {
	t.Helper()

	for _, s := range servers {
		if s.Name == name {
			return s
		}
	}

	require.Failf(t, "server not found", "no server %q in %v", name, servers)
	return db.McpServer{}
}

func countAccessKeys(t *testing.T, store *SqlDb, where string, args ...any) int64 {
	t.Helper()

	n, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from access_key where "+where), args...)
	require.NoError(t, err)
	return n
}

func TestSetMcpServers_CreateWithAndWithoutToken(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "github", URL: "https://api.githubcopilot.com/mcp/", Enabled: true}, Token: mcpTokenCreate("enc-1")},
		{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/sse", Enabled: false}},
	})
	require.NoError(t, err)
	require.Len(t, servers, 2)

	github := serverByName(t, servers, "github")
	assert.True(t, github.HasToken)
	assert.True(t, github.Enabled)
	assert.Equal(t, inventoryID, github.InventoryID)

	docs := serverByName(t, servers, "docs")
	assert.False(t, docs.HasToken)
	assert.False(t, docs.Enabled)

	tokens, err := store.GetMcpServerTokens(projectID, inventoryID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, db.AccessKeyMcp, tokens[0].Owner)
	require.NotNil(t, tokens[0].McpID)
	assert.Equal(t, github.ID, *tokens[0].McpID)
	require.NotNil(t, tokens[0].ProjectID)
	assert.Equal(t, projectID, *tokens[0].ProjectID)
	assert.Equal(t, "enc-1", *tokens[0].Secret)
	assert.Equal(t, db.AccessKeyString, tokens[0].Type)
	assert.Equal(t, "github", tokens[0].Name)

	read, err := store.GetMcpServers(projectID, inventoryID)
	require.NoError(t, err)
	assert.Equal(t, servers, read)
}

func TestSetMcpServers_OwnerAndServerComeFromStore(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)
	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	// A caller cannot redirect the key to another project, owner or server.
	foreignServer := 9999
	_, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{{
		Server: db.McpServer{Name: "a", URL: "https://a"},
		Token: &db.McpTokenChange{Operation: db.McpTokenCreate, Key: db.AccessKey{
			Secret:    mcpSecret("x"),
			ProjectID: &otherProject.ID,
			Owner:     db.AccessKeyShared,
			McpID:     &foreignServer,
		}},
	}})
	require.NoError(t, err)

	assert.Equal(t, int64(1), countAccessKeys(t, store, "owner = ? and project_id = ? and mcp_id is not null", db.AccessKeyMcp, projectID))
	assert.Equal(t, int64(0), countAccessKeys(t, store, "project_id = ?", otherProject.ID))
}

func TestSetMcpServers_TokenUpdateAndDelete(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}, Token: mcpTokenCreate("old")},
	})
	require.NoError(t, err)
	a := servers[0]

	// No token change keeps the token.
	servers, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{ID: a.ID, Name: "a", URL: "https://a2"}},
	})
	require.NoError(t, err)
	assert.True(t, servers[0].HasToken)
	assert.Equal(t, "https://a2", servers[0].URL)

	_, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{ID: a.ID, Name: "a", URL: "https://a2"},
			Token: &db.McpTokenChange{Operation: db.McpTokenUpdate, Key: db.AccessKey{Secret: mcpSecret("new")}}},
	})
	require.NoError(t, err)

	tokens, err := store.GetMcpServerTokens(projectID, inventoryID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "new", *tokens[0].Secret)

	servers, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{ID: a.ID, Name: "a", URL: "https://a2"}, Token: &db.McpTokenChange{Operation: db.McpTokenDelete}},
	})
	require.NoError(t, err)
	assert.False(t, servers[0].HasToken)
	assert.Equal(t, int64(0), countAccessKeys(t, store, "owner = ?", db.AccessKeyMcp))
}

func TestSetMcpServers_TokenOperationMismatch(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "with", URL: "https://a"}, Token: mcpTokenCreate("t")},
		{Server: db.McpServer{Name: "without", URL: "https://b"}},
	})
	require.NoError(t, err)
	with := serverByName(t, servers, "with")
	without := serverByName(t, servers, "without")

	tests := []struct {
		name    string
		change  db.McpServerChange
		errText string
	}{
		{"second token", db.McpServerChange{Server: with, Token: mcpTokenCreate("t2")}, "already has a token"},
		{"update missing token", db.McpServerChange{Server: without,
			Token: &db.McpTokenChange{Operation: db.McpTokenUpdate, Key: db.AccessKey{Secret: mcpSecret("t")}}}, "has no token"},
		{"create without secret", db.McpServerChange{Server: without,
			Token: &db.McpTokenChange{Operation: db.McpTokenCreate}}, "secret can not be empty"},
		{"unknown operation", db.McpServerChange{Server: without,
			Token: &db.McpTokenChange{Operation: "rotate", Key: db.AccessKey{Secret: mcpSecret("t")}}}, "invalid mcp token operation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := []db.McpServerChange{{Server: with}, {Server: without}}
			if tt.change.Server.ID == with.ID {
				changes[0] = tt.change
			} else {
				changes[1] = tt.change
			}

			_, err := store.SetMcpServers(projectID, inventoryID, changes)
			var validationErr *common_errors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.ErrorContains(t, err, tt.errText)

			// The failed request changed nothing.
			read, err := store.GetMcpServers(projectID, inventoryID)
			require.NoError(t, err)
			assert.Equal(t, servers, read)
			assert.Equal(t, int64(1), countAccessKeys(t, store, "owner = ?", db.AccessKeyMcp))
		})
	}
}

func TestSetMcpServers_MissingServerIsDeletedWithToken(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "keep", URL: "https://a"}, Token: mcpTokenCreate("k")},
		{Server: db.McpServer{Name: "drop", URL: "https://b"}, Token: mcpTokenCreate("d")},
	})
	require.NoError(t, err)
	keep := serverByName(t, servers, "keep")

	servers, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{{Server: keep}})
	require.NoError(t, err)
	require.Len(t, servers, 1)
	assert.Equal(t, "keep", servers[0].Name)
	assert.Equal(t, int64(1), countAccessKeys(t, store, "owner = ?", db.AccessKeyMcp))
	assert.Equal(t, int64(1), countAccessKeys(t, store, "mcp_id = ?", keep.ID))
}

func TestSetMcpServers_SwapNames(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}},
		{Server: db.McpServer{Name: "b", URL: "https://b"}},
	})
	require.NoError(t, err)
	a := serverByName(t, servers, "a")
	b := serverByName(t, servers, "b")

	a.Name, b.Name = "b", "a"
	servers, err = store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{{Server: a}, {Server: b}})
	require.NoError(t, err)

	assert.Equal(t, a.ID, serverByName(t, servers, "b").ID)
	assert.Equal(t, b.ID, serverByName(t, servers, "a").ID)
}

func TestSetMcpServers_RejectsInvalidLists(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}},
	})
	require.NoError(t, err)
	a := servers[0]

	// A server of another inventory of the same project.
	otherInv, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "other", Type: db.InventoryMcp})
	require.NoError(t, err)
	otherServers, err := store.SetMcpServers(projectID, otherInv.ID, []db.McpServerChange{
		{Server: db.McpServer{Name: "x", URL: "https://x"}},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		changes []db.McpServerChange
		errText string
	}{
		{"duplicate name", []db.McpServerChange{{Server: a}, {Server: db.McpServer{Name: "a", URL: "https://b"}}}, "not unique"},
		{"invalid url", []db.McpServerChange{{Server: db.McpServer{Name: "b", URL: "ftp://b"}}}, "http or https"},
		{"server of another inventory", []db.McpServerChange{{Server: otherServers[0]}}, "does not belong to the inventory"},
		{"same server twice", []db.McpServerChange{{Server: a}, {Server: db.McpServer{ID: a.ID, Name: "a2", URL: "https://a"}}}, "listed twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.SetMcpServers(projectID, inventoryID, tt.changes)
			assert.ErrorContains(t, err, tt.errText)

			read, err := store.GetMcpServers(projectID, inventoryID)
			require.NoError(t, err)
			assert.Equal(t, servers, read)
		})
	}
}

func TestSetMcpServers_ScopedByProjectAndType(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	_, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}, Token: mcpTokenCreate("t")},
	})
	require.NoError(t, err)

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	// The inventory read through another project is not there.
	_, err = store.SetMcpServers(otherProject.ID, inventoryID, nil)
	assert.ErrorIs(t, err, db.ErrNotFound)

	servers, err := store.GetMcpServers(otherProject.ID, inventoryID)
	require.NoError(t, err)
	assert.Empty(t, servers)

	tokens, err := store.GetMcpServerTokens(otherProject.ID, inventoryID)
	require.NoError(t, err)
	assert.Empty(t, tokens)

	// An inventory of another type takes no MCP servers.
	static, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "hosts", Type: db.InventoryStatic})
	require.NoError(t, err)
	_, err = store.SetMcpServers(projectID, static.ID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}},
	})
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestDeleteInventory_DeletesMcpServersAndTokens(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	_, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}, Token: mcpTokenCreate("t")},
	})
	require.NoError(t, err)

	shared, err := store.CreateAccessKey(db.AccessKey{Name: "shared", ProjectID: &projectID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	require.NoError(t, store.DeleteInventory(projectID, inventoryID))

	n, err := store.Sql().SelectInt("select count(*) from project__inventory_mcp")
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, int64(0), countAccessKeys(t, store, "owner = ?", db.AccessKeyMcp))

	_, err = store.GetAccessKey(projectID, shared.ID)
	assert.NoError(t, err, "keys outside the inventory stay")
}

func TestMcpTokens_HiddenFromProjectKeys(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	servers, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "a", URL: "https://a"}, Token: mcpTokenCreate("t")},
	})
	require.NoError(t, err)

	keys, err := store.GetAccessKeys(projectID, db.GetAccessKeyOptions{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Empty(t, keys, "the project Key Store lists shared keys only")

	_, err = store.GetAccessKeys(projectID, db.GetAccessKeyOptions{Owner: db.AccessKeyMcp}, db.RetrieveQueryParams{})
	assert.ErrorContains(t, err, "mcp_id is required")

	keys, err = store.GetAccessKeys(projectID, db.GetAccessKeyOptions{Owner: db.AccessKeyMcp, McpID: &servers[0].ID}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestCreateAccessKey_RejectsMcpOwner(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newMcpTestInventory(t, store)
	serverID := 1

	tests := []struct {
		name string
		key  db.AccessKey
	}{
		{"mcp owner", db.AccessKey{Name: "k", ProjectID: &projectID, Type: db.AccessKeyString, Owner: db.AccessKeyMcp}},
		{"mcp server id", db.AccessKey{Name: "k", ProjectID: &projectID, Type: db.AccessKeyString, McpID: &serverID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.CreateAccessKey(tt.key)
			assert.ErrorContains(t, err, "created through the inventory")
		})
	}
}

func countInventories(t *testing.T, store *SqlDb, projectID int) int64 {
	t.Helper()

	n, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from project__inventory where project_id = ?"), projectID)
	require.NoError(t, err)
	return n
}

func TestCreateMcpInventory_StoresInventoryServersAndTokens(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	inv, servers, err := store.CreateMcpInventory(
		db.Inventory{ProjectID: project.ID, Name: "agent mcp", Type: db.InventoryMcp},
		[]db.McpServerChange{
			{Server: db.McpServer{Name: "github", URL: "https://mcp.example.com/github", Enabled: true}, Token: mcpTokenCreate("enc-1")},
			{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true}},
		})
	require.NoError(t, err)
	require.NotZero(t, inv.ID)
	require.Len(t, servers, 2)
	assert.True(t, serverByName(t, servers, "github").HasToken)
	assert.False(t, serverByName(t, servers, "docs").HasToken)

	read, err := store.GetInventory(project.ID, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, db.InventoryMcp, read.Type)
	assert.Equal(t, "agent mcp", read.Name)
	assert.Equal(t, int64(1), countAccessKeys(t, store, "owner = ? and project_id = ?", db.AccessKeyMcp, project.ID))
}

// A failure in the middle of the list leaves nothing behind: neither the
// inventory row nor the servers and tokens written before it.
func TestCreateMcpInventory_RollsBackOnFailure(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	inv, servers, err := store.CreateMcpInventory(
		db.Inventory{ProjectID: project.ID, Name: "agent mcp", Type: db.InventoryMcp},
		[]db.McpServerChange{
			{Server: db.McpServer{Name: "github", URL: "https://mcp.example.com/github", Enabled: true}, Token: mcpTokenCreate("enc-1")},
			// A new server has no token, so update fails inside the transaction.
			{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/docs"},
				Token: &db.McpTokenChange{Operation: db.McpTokenUpdate, Key: db.AccessKey{Secret: mcpSecret("enc-2")}}},
		})
	require.Error(t, err)
	var validationErr *common_errors.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Zero(t, inv.ID)
	assert.Nil(t, servers)

	assert.Equal(t, int64(0), countInventories(t, store, project.ID))
	assert.Equal(t, int64(0), countAccessKeys(t, store, "project_id = ?", project.ID))
}

func TestCreateMcpInventory_RejectsOtherTypes(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	_, _, err = store.CreateMcpInventory(db.Inventory{ProjectID: project.ID, Name: "x", Type: db.InventoryStatic}, nil)
	require.Error(t, err)
	assert.Equal(t, int64(0), countInventories(t, store, project.ID))
}

func TestUpdateMcpInventory_UpdatesRowAndServers(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	created, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "github", URL: "https://mcp.example.com/github", Enabled: true}, Token: mcpTokenCreate("enc-1")},
		{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true}},
	})
	require.NoError(t, err)
	github := serverByName(t, created, "github")

	servers, err := store.UpdateMcpInventory(
		db.Inventory{ID: inventoryID, ProjectID: projectID, Name: "renamed", Type: db.InventoryMcp},
		[]db.McpServerChange{
			{Server: db.McpServer{ID: github.ID, Name: "github", URL: github.URL, Enabled: false},
				Token: &db.McpTokenChange{Operation: db.McpTokenDelete}},
		})
	require.NoError(t, err)
	require.Len(t, servers, 1)
	assert.False(t, servers[0].Enabled)
	assert.False(t, servers[0].HasToken)

	read, err := store.GetInventory(projectID, inventoryID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", read.Name)
	assert.Equal(t, int64(0), countAccessKeys(t, store, "owner = ? and project_id = ?", db.AccessKeyMcp, projectID))
}

// A failing server change leaves the inventory row as it was, not renamed.
func TestUpdateMcpInventory_RollsBackOnFailure(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	created, err := store.SetMcpServers(projectID, inventoryID, []db.McpServerChange{
		{Server: db.McpServer{Name: "github", URL: "https://mcp.example.com/github", Enabled: true}},
	})
	require.NoError(t, err)

	_, err = store.UpdateMcpInventory(
		db.Inventory{ID: inventoryID, ProjectID: projectID, Name: "renamed", Type: db.InventoryMcp},
		[]db.McpServerChange{
			{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/docs"}, Token: mcpTokenCreate("enc-1")},
			{Server: db.McpServer{ID: created[0].ID, Name: "github", URL: created[0].URL},
				Token: &db.McpTokenChange{Operation: db.McpTokenUpdate, Key: db.AccessKey{Secret: mcpSecret("enc-2")}}},
		})
	require.Error(t, err)

	read, err := store.GetInventory(projectID, inventoryID)
	require.NoError(t, err)
	assert.Equal(t, "mcp", read.Name)

	servers, err := store.GetMcpServers(projectID, inventoryID)
	require.NoError(t, err)
	assert.Equal(t, created, servers)
	assert.Equal(t, int64(0), countAccessKeys(t, store, "project_id = ?", projectID))
}

func TestUpdateMcpInventory_TypeAndProjectScoped(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, inventoryID := newMcpTestInventory(t, store)

	static, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "hosts", Type: db.InventoryStatic})
	require.NoError(t, err)

	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	t.Run("static inventory can not become mcp", func(t *testing.T) {
		_, err := store.UpdateMcpInventory(db.Inventory{ID: static.ID, ProjectID: projectID, Name: "hosts", Type: db.InventoryMcp}, nil)
		assert.ErrorIs(t, err, db.ErrNotFound)

		read, err := store.GetInventory(projectID, static.ID)
		require.NoError(t, err)
		assert.Equal(t, db.InventoryStatic, read.Type)
	})

	t.Run("inventory of another project", func(t *testing.T) {
		_, err := store.UpdateMcpInventory(db.Inventory{ID: inventoryID, ProjectID: other.ID, Name: "stolen", Type: db.InventoryMcp}, nil)
		assert.ErrorIs(t, err, db.ErrNotFound)

		read, err := store.GetInventory(projectID, inventoryID)
		require.NoError(t, err)
		assert.Equal(t, "mcp", read.Name)
	})
}
