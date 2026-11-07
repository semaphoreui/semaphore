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
	assert.Equal(t, "docs", got.McpServers[1].Name)
	assert.Empty(t, got.McpServers[1].Token)
}
