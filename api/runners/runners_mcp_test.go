package runners

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const runnerMcpToken = "mcp-tok-8d2e41"

func mcpTaskRunner(servers []db.McpServer) *tasks.TaskRunner {
	tr := tasks.NewTaskRunner(db.Task{ID: 5, ProjectID: 3}, nil, "", nil)
	tr.Inventory = db.Inventory{ID: 9, ProjectID: 3, Type: db.InventoryMcp, McpServers: servers}
	return tr
}

// The token of an MCP server goes to the runner as a decrypted access key of
// the task, keyed by its ID, and never inside the serialized inventory.
func TestCollectTaskAccessKeys_McpTokens(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	c := &RunnerController{encryptionService: server.NewAccessKeyEncryptionService(store, store, store, store)}

	keyID := 41
	tr := mcpTaskRunner([]db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, HasToken: true, Token: runnerMcpToken, TokenKeyID: &keyID},
		{ID: 2, Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true},
	})

	keys := map[int]db.AccessKey{}
	require.NoError(t, c.collectTaskAccessKeys(tr, 1, keys))

	require.Contains(t, keys, keyID)
	assert.Equal(t, runnerMcpToken, keys[keyID].String)
	assert.Equal(t, db.AccessKeyString, keys[keyID].Type)
	assert.Equal(t, db.AccessKeyMcp, keys[keyID].Owner)

	inventory, err := json.Marshal(tr.Inventory)
	require.NoError(t, err)
	assert.NotContains(t, string(inventory), runnerMcpToken)
	assert.Contains(t, string(inventory), `"token_key_id":41`)
}

// A token that was not decrypted fails the dispatch instead of sending the
// job without it.
func TestCollectTaskAccessKeys_McpTokenMissing(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	c := &RunnerController{encryptionService: server.NewAccessKeyEncryptionService(store, store, store, store)}

	keyID := 41
	tr := mcpTaskRunner([]db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, HasToken: true, TokenKeyID: &keyID},
	})

	keys := map[int]db.AccessKey{}
	assert.ErrorContains(t, c.collectTaskAccessKeys(tr, 1, keys), "mcp server github")
	assert.NotContains(t, keys, keyID)
}
