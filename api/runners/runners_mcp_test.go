package runners

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/services/runners"
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
// the task's own job, keyed by its ID, and never inside the serialized
// inventory or among the keys the jobs of a response share.
func TestCollectMcpTokenKeys(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	c := &RunnerController{encryptionService: server.NewAccessKeyEncryptionService(store, store, store, store)}

	keyID := 41
	tr := mcpTaskRunner([]db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, HasToken: true, Token: runnerMcpToken, TokenKeyID: &keyID},
		{ID: 2, Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true},
	})

	keys, err := c.collectMcpTokenKeys(tr, 1)
	require.NoError(t, err)

	require.Len(t, keys, 1)
	require.Contains(t, keys, keyID)
	assert.Equal(t, runnerMcpToken, keys[keyID].String)
	assert.Equal(t, db.AccessKeyString, keys[keyID].Type)
	assert.Equal(t, db.AccessKeyMcp, keys[keyID].Owner)

	inventory, err := json.Marshal(tr.Inventory)
	require.NoError(t, err)
	assert.NotContains(t, string(inventory), runnerMcpToken)
	assert.Contains(t, string(inventory), `"token_key_id":41`)

	shared := map[int]db.AccessKey{}
	require.NoError(t, c.collectTaskAccessKeys(tr, 1, shared))
	assert.NotContains(t, shared, keyID)
}

// A token that was not decrypted fails the dispatch instead of sending the
// job without it.
func TestCollectMcpTokenKeys_TokenMissing(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	c := &RunnerController{encryptionService: server.NewAccessKeyEncryptionService(store, store, store, store)}

	keyID := 41
	tr := mcpTaskRunner([]db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, HasToken: true, TokenKeyID: &keyID},
	})

	keys, err := c.collectMcpTokenKeys(tr, 1)
	assert.ErrorContains(t, err, "mcp server github")
	assert.Empty(t, keys)
}

// Two tasks of one runner were queued before and after an edit of the same
// MCP server: each keeps the URL and the token it was queued with, and both
// tokens share one access key ID, because an edit keeps the key. The runner
// must get, for each job, the token that belongs to that job's URL.
func TestPrepareRemoteJob_McpTokensOfTwoJobsStayApart(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	c := &RunnerController{encryptionService: server.NewAccessKeyEncryptionService(store, store, store, store)}

	const oldURL, oldToken = "https://old.example.com/mcp", "mcp-tok-old-51c7"
	const newURL, newToken = "https://new.example.com/mcp", "mcp-tok-new-a94f"

	keyID := 41
	queued := []struct {
		taskID     int
		url, token string
	}{
		{taskID: 5, url: oldURL, token: oldToken},
		{taskID: 6, url: newURL, token: newToken},
	}

	runner := db.Runner{ID: 1}
	state := runners.RunnerState{AccessKeys: make(map[int]db.AccessKey)}

	for _, q := range queued {
		tr := tasks.NewTaskRunner(db.Task{ID: q.taskID, ProjectID: 3}, nil, "", nil)
		tr.Inventory = db.Inventory{ID: 9, ProjectID: 3, Type: db.InventoryMcp, McpServers: []db.McpServer{
			{ID: 1, Name: "github", URL: q.url, Enabled: true, HasToken: true, Token: q.token, TokenKeyID: &keyID},
		}}
		c.prepareRemoteJob(tr, &runner, &state)
	}

	// The runner sees only what survives the wire.
	body, err := json.Marshal(state)
	require.NoError(t, err)

	var received runners.RunnerState
	require.NoError(t, json.Unmarshal(body, &received))
	require.Len(t, received.NewJobs, len(queued))

	for i, q := range queued {
		job := received.NewJobs[i]
		require.Equal(t, q.taskID, job.Task.ID)
		require.Len(t, job.Inventory.McpServers, 1)

		srv := job.Inventory.McpServers[0]
		require.NotNil(t, srv.TokenKeyID)
		assert.Equal(t, q.url, srv.URL)
		assert.Equal(t, q.token, job.AccessKeys[*srv.TokenKeyID].String,
			"task %d sends this token to %s", q.taskID, srv.URL)
	}

	assert.NotContains(t, received.AccessKeys, keyID, "a token is never shared between jobs")
}
