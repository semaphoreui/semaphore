package server

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An mcp inventory for a task run carries its enabled servers with the
// decrypted tokens; disabled servers are left out.
func TestInventoryService_McpServersWithDecryptedTokens(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	encryption := NewAccessKeyEncryptionService(store, store, store, store)

	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	encrypt := func(secret string) db.AccessKey {
		key := db.AccessKey{Name: "t", Type: db.AccessKeyString, String: secret, ProjectID: &project.ID, Owner: db.AccessKeyMcp}
		require.NoError(t, encryption.SerializeSecret(&key))
		return db.AccessKey{Type: key.Type, Secret: key.Secret}
	}

	inv, _, err := store.CreateMcpInventory(db.Inventory{ProjectID: project.ID, Name: "mcp", Type: db.InventoryMcp}, []db.McpServerChange{
		{Server: db.McpServer{Name: "github", URL: "https://mcp.example.com/github", Enabled: true},
			Token: &db.McpTokenChange{Operation: db.McpTokenCreate, Key: encrypt("gh-token")}},
		{Server: db.McpServer{Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true}},
		{Server: db.McpServer{Name: "off", URL: "https://mcp.example.com/off", Enabled: false},
			Token: &db.McpTokenChange{Operation: db.McpTokenCreate, Key: encrypt("off-token")}},
	})
	require.NoError(t, err)

	got, err := NewInventoryService(store, store, store, encryption).GetInventory(project.ID, inv.ID)
	require.NoError(t, err)

	require.Len(t, got.McpServers, 2)
	assert.Equal(t, "github", got.McpServers[0].Name)
	assert.Equal(t, "gh-token", got.McpServers[0].Token)
	require.NotNil(t, got.McpServers[0].TokenKeyID)
	assert.Equal(t, "docs", got.McpServers[1].Name)
	assert.Empty(t, got.McpServers[1].Token)
	assert.Nil(t, got.McpServers[1].TokenKeyID)

	tokens, err := store.GetMcpServerTokens(project.ID, inv.ID)
	require.NoError(t, err)
	tokenKeyIDs := make([]int, 0, len(tokens))
	for _, key := range tokens {
		tokenKeyIDs = append(tokenKeyIDs, key.ID)
	}
	assert.Contains(t, tokenKeyIDs, *got.McpServers[0].TokenKeyID)
}

// editingMcpRepository changes the URL and the token of every server right
// after a read of the servers, like an inventory edit landing between two
// reads of a task about to start.
type editingMcpRepository struct {
	*sql.SqlDb
	t        *testing.T
	newToken db.AccessKey
}

func (r *editingMcpRepository) edit(projectID int, inventoryID int, servers []db.McpServer) {
	changes := make([]db.McpServerChange, 0, len(servers))
	for _, server := range servers {
		server.URL += "/v2"
		changes = append(changes, db.McpServerChange{
			Server: server,
			Token:  &db.McpTokenChange{Operation: db.McpTokenUpdate, Key: r.newToken},
		})
	}

	_, err := r.SqlDb.SetMcpServers(projectID, inventoryID, changes)
	require.NoError(r.t, err)
}

func (r *editingMcpRepository) GetMcpServers(projectID int, inventoryID int) ([]db.McpServer, error) {
	servers, err := r.SqlDb.GetMcpServers(projectID, inventoryID)
	if err == nil {
		r.edit(projectID, inventoryID, servers)
	}
	return servers, err
}

func (r *editingMcpRepository) GetMcpServersWithTokens(projectID int, inventoryID int) ([]db.McpServerWithToken, error) {
	items, err := r.SqlDb.GetMcpServersWithTokens(projectID, inventoryID)
	if err == nil {
		servers := make([]db.McpServer, 0, len(items))
		for _, item := range items {
			servers = append(servers, item.Server)
		}
		r.edit(projectID, inventoryID, servers)
	}
	return items, err
}

// A task never gets the new token of a server together with its old URL: an
// edit of the inventory while the task loads it gives either the old pair or
// the new one.
func TestInventoryService_McpServerTokenMatchesItsURL(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	encryption := NewAccessKeyEncryptionService(store, store, store, store)

	project, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)

	encrypt := func(secret string) db.AccessKey {
		key := db.AccessKey{Name: "t", Type: db.AccessKeyString, String: secret, ProjectID: &project.ID, Owner: db.AccessKeyMcp}
		require.NoError(t, encryption.SerializeSecret(&key))
		return db.AccessKey{Type: key.Type, Secret: key.Secret}
	}

	const oldURL = "https://old.example.com/mcp"

	inv, _, err := store.CreateMcpInventory(db.Inventory{ProjectID: project.ID, Name: "mcp", Type: db.InventoryMcp}, []db.McpServerChange{
		{Server: db.McpServer{Name: "github", URL: oldURL, Enabled: true},
			Token: &db.McpTokenChange{Operation: db.McpTokenCreate, Key: encrypt("old-token")}},
	})
	require.NoError(t, err)

	repo := &editingMcpRepository{SqlDb: store, t: t, newToken: encrypt("new-token")}

	got, err := NewInventoryService(store, store, repo, encryption).GetInventory(project.ID, inv.ID)
	require.NoError(t, err)
	require.Len(t, got.McpServers, 1)

	pairs := map[string]string{oldURL: "old-token", oldURL + "/v2": "new-token"}
	assert.Equal(t, pairs[got.McpServers[0].URL], got.McpServers[0].Token)

	// The edit did land: the next task runs with the new pair.
	next, err := NewInventoryService(store, store, store, encryption).GetInventory(project.ID, inv.ID)
	require.NoError(t, err)
	require.Len(t, next.McpServers, 1)
	assert.Equal(t, oldURL+"/v2", next.McpServers[0].URL)
	assert.Equal(t, "new-token", next.McpServers[0].Token)
}
