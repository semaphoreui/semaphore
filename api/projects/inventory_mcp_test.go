package projects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpTestSecret = "mcp-s3cr3t-token"

func (f resourceFixture) inventoryController() *InventoryController {
	return NewInventoryController(server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store))
}

// addMcpInventory posts an mcp inventory and returns the response.
func (f resourceFixture) addMcpInventory(t *testing.T, servers string) (*httptest.ResponseRecorder, db.Inventory) {
	t.Helper()

	body := fmt.Sprintf(`{"name":"agent mcp","project_id":%d,"type":"mcp","mcp_servers":%s}`, f.project.ID, servers)
	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	f.inventoryController().AddInventory(w, r)

	var inv db.Inventory
	if w.Code == http.StatusCreated {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &inv))
	}
	return w, inv
}

// putInventory sends PUT for an existing inventory with the given body tail.
func (f resourceFixture) putInventory(t *testing.T, inv db.Inventory, body string) *httptest.ResponseRecorder {
	t.Helper()

	stored, err := f.store.GetInventory(f.project.ID, inv.ID)
	require.NoError(t, err)

	r, _ := f.request(http.MethodPut, body, map[string]any{"inventory": stored}, nil)
	w := httptest.NewRecorder()
	f.inventoryController().UpdateInventory(w, r)
	return w
}

func mcpPutBody(f resourceFixture, inv db.Inventory, servers string) string {
	return fmt.Sprintf(`{"id":%d,"name":"agent mcp","project_id":%d,"type":"mcp","mcp_servers":%s}`, inv.ID, f.project.ID, servers)
}

func mcpTokens(t *testing.T, f resourceFixture, inventoryID int) []db.AccessKey {
	t.Helper()
	tokens, err := f.store.GetMcpServerTokens(f.project.ID, inventoryID)
	require.NoError(t, err)
	return tokens
}

func mcpServer(t *testing.T, servers []db.McpServer, name string) db.McpServer {
	t.Helper()
	for _, s := range servers {
		if s.Name == name {
			return s
		}
	}
	require.Failf(t, "server not found", "no server %q", name)
	return db.McpServer{}
}

func TestAddInventory_McpStoresServersAndEncryptedTokens(t *testing.T) {
	f := newResourceFixture(t)

	body := fmt.Sprintf(`{"name":"agent mcp","project_id":%d,"type":"mcp","mcp_servers":[
		{"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"create","secret":"%s"}},
		{"name":"docs","url":"https://mcp.example.com/docs","enabled":false}]}`, f.project.ID, mcpTestSecret)
	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	f.inventoryController().AddInventory(w, r)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	assert.NotContains(t, w.Body.String(), mcpTestSecret)

	var inv db.Inventory
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &inv))
	assert.Equal(t, db.InventoryMcp, inv.Type)
	require.Len(t, inv.McpServers, 2)
	assert.True(t, mcpServer(t, inv.McpServers, "github").HasToken)
	assert.False(t, mcpServer(t, inv.McpServers, "docs").HasToken)
	assert.False(t, mcpServer(t, inv.McpServers, "docs").Enabled)

	tokens := mcpTokens(t, f, inv.ID)
	require.Len(t, tokens, 1)
	assert.Equal(t, db.AccessKeyMcp, tokens[0].Owner)
	require.NotNil(t, tokens[0].Secret)
	assert.NotEqual(t, mcpTestSecret, *tokens[0].Secret, "the token must not be stored as plaintext")

	key := tokens[0]
	require.NoError(t, server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store).DeserializeSecret(&key))
	assert.Equal(t, mcpTestSecret, key.String)

	created := only(t, rec, audit.ResourceInventoryCreate)
	assert.Equal(t, audit.InventoryMetadata{Type: "mcp", McpServers: 2, TokensCreated: 1}, created.Event.Metadata)
}

func TestGetInventory_McpListsServersWithoutTokens(t *testing.T) {
	f := newResourceFixture(t)
	w, inv := f.addMcpInventory(t, `[{"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"create","secret":"`+mcpTestSecret+`"}}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	stored, err := f.store.GetInventory(f.project.ID, inv.ID)
	require.NoError(t, err)

	r, _ := f.request(http.MethodGet, "", map[string]any{"inventory": stored}, nil)
	w = httptest.NewRecorder()
	f.inventoryController().GetInventory(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), mcpTestSecret)
	assert.NotContains(t, w.Body.String(), `"secret"`)

	var got db.Inventory
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got.McpServers, 1)
	assert.Equal(t, "github", got.McpServers[0].Name)
	assert.True(t, got.McpServers[0].HasToken)
}

func TestUpdateInventory_McpTokenAndServerChanges(t *testing.T) {
	f := newResourceFixture(t)
	w, inv := f.addMcpInventory(t, `[
		{"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"create","secret":"old"}},
		{"name":"docs","url":"https://mcp.example.com/docs","enabled":true,"token":{"operation":"create","secret":"docs"}},
		{"name":"jira","url":"https://mcp.example.com/jira","enabled":true}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	github := mcpServer(t, inv.McpServers, "github")
	docs := mcpServer(t, inv.McpServers, "docs")
	jira := mcpServer(t, inv.McpServers, "jira")

	// github: token replaced; docs: token removed and server disabled;
	// jira: left out of the list, so deleted.
	stored, err := f.store.GetInventory(f.project.ID, inv.ID)
	require.NoError(t, err)
	r, rec := f.request(http.MethodPut, mcpPutBody(f, inv, fmt.Sprintf(`[
		{"id":%d,"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"update","secret":"new"}},
		{"id":%d,"name":"docs","url":"https://mcp.example.com/docs","enabled":false,"token":{"operation":"delete"}}]`, github.ID, docs.ID)),
		map[string]any{"inventory": stored}, nil)
	w = httptest.NewRecorder()
	f.inventoryController().UpdateInventory(w, r)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

	servers, err := f.store.GetMcpServers(f.project.ID, inv.ID)
	require.NoError(t, err)
	require.Len(t, servers, 2)
	assert.True(t, mcpServer(t, servers, "github").HasToken)
	assert.False(t, mcpServer(t, servers, "docs").HasToken)
	assert.False(t, mcpServer(t, servers, "docs").Enabled)
	for _, s := range servers {
		assert.NotEqual(t, jira.ID, s.ID)
	}

	tokens := mcpTokens(t, f, inv.ID)
	require.Len(t, tokens, 1)
	key := tokens[0]
	require.NoError(t, server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store).DeserializeSecret(&key))
	assert.Equal(t, "new", key.String)

	updated := only(t, rec, audit.ResourceInventoryUpdate)
	assert.Equal(t, audit.InventoryMetadata{Type: "mcp", McpServers: 2, TokensUpdated: 1, TokensDeleted: 1}, updated.Event.Metadata)
}

func TestUpdateInventory_McpWithoutServersKeepsThem(t *testing.T) {
	f := newResourceFixture(t)
	w, inv := f.addMcpInventory(t, `[{"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"create","secret":"x"}}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	w = f.putInventory(t, inv, fmt.Sprintf(`{"id":%d,"name":"renamed","project_id":%d,"type":"mcp"}`, inv.ID, f.project.ID))
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

	read, err := f.store.GetInventory(f.project.ID, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", read.Name)

	servers, err := f.store.GetMcpServers(f.project.ID, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, inv.McpServers, servers)
}

// A server of another inventory can not be taken over by naming its ID, and
// the failed request changes nothing.
func TestUpdateInventory_McpForeignServerIDIsRejected(t *testing.T) {
	f := newResourceFixture(t)
	w, own := f.addMcpInventory(t, `[{"name":"github","url":"https://mcp.example.com/github","enabled":true}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	w, foreign := f.addMcpInventory(t, `[{"name":"secret","url":"https://mcp.example.com/secret","enabled":true,"token":{"operation":"create","secret":"x"}}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	w = f.putInventory(t, own, mcpPutBody(f, own, fmt.Sprintf(
		`[{"id":%d,"name":"stolen","url":"https://evil.example.com","enabled":true,"token":{"operation":"update","secret":"y"}}]`,
		foreign.McpServers[0].ID)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	servers, err := f.store.GetMcpServers(f.project.ID, own.ID)
	require.NoError(t, err)
	assert.Equal(t, own.McpServers, servers)

	servers, err = f.store.GetMcpServers(f.project.ID, foreign.ID)
	require.NoError(t, err)
	assert.Equal(t, foreign.McpServers, servers)
}

// The inventory row and the servers are one transaction: a failing server
// leaves no inventory and no tokens behind, and nothing is audited.
func TestAddInventory_McpFailureLeavesNothing(t *testing.T) {
	f := newResourceFixture(t)

	body := fmt.Sprintf(`{"name":"agent mcp","project_id":%d,"type":"mcp","mcp_servers":[
		{"name":"github","url":"https://mcp.example.com/github","enabled":true,"token":{"operation":"create","secret":"x"}},
		{"name":"docs","url":"https://mcp.example.com/docs","enabled":true,"token":{"operation":"update","secret":"y"}}]}`, f.project.ID)
	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	f.inventoryController().AddInventory(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "has no token")
	assert.Empty(t, rec.All())

	inventories, err := f.store.GetInventories(f.project.ID, db.RetrieveQueryParams{}, nil)
	require.NoError(t, err)
	assert.Empty(t, inventories)

	keys, err := f.store.Sql().SelectInt(f.store.PrepareQuery("select count(*) from access_key where owner = ?"), db.AccessKeyMcp)
	require.NoError(t, err)
	assert.Zero(t, keys)
}

func TestAddInventory_McpRejectsInvalidRequests(t *testing.T) {
	f := newResourceFixture(t)
	pid := f.project.ID

	tests := []struct {
		name string
		body string
	}{
		{"inventory text", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","inventory":"[all]"}`, pid)},
		{"ssh key", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","ssh_key_id":%d}`, pid, f.key.ID)},
		{"servers on static", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"static","inventory":"[all]","mcp_servers":[{"name":"x","url":"https://mcp.example.com","enabled":true}]}`, pid)},
		{"stdio url", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"x","url":"file:///bin/sh","enabled":true}]}`, pid)},
		{"credentials in url", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"x","url":"https://u:p@mcp.example.com","enabled":true}]}`, pid)},
		{"bad name", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"a b","url":"https://mcp.example.com","enabled":true}]}`, pid)},
		{"duplicate names", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"x","url":"https://a.example.com","enabled":true},{"name":"x","url":"https://b.example.com","enabled":true}]}`, pid)},
		{"empty secret", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"x","url":"https://mcp.example.com","enabled":true,"token":{"operation":"create","secret":""}}]}`, pid)},
		{"unknown operation", fmt.Sprintf(`{"name":"m","project_id":%d,"type":"mcp","mcp_servers":[{"name":"x","url":"https://mcp.example.com","enabled":true,"token":{"operation":"steal","secret":"s"}}]}`, pid)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, rec := f.request(http.MethodPost, tt.body, nil, nil)
			w := httptest.NewRecorder()
			f.inventoryController().AddInventory(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.Empty(t, rec.All())
		})
	}

	inventories, err := f.store.GetInventories(pid, db.RetrieveQueryParams{}, nil)
	require.NoError(t, err)
	assert.Empty(t, inventories)
}

func TestUpdateInventory_TypeCanNotSwitchToOrFromMcp(t *testing.T) {
	f := newResourceFixture(t)

	w, mcp := f.addMcpInventory(t, `[{"name":"github","url":"https://mcp.example.com/github","enabled":true}]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	static, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "hosts", Type: db.InventoryStatic, Inventory: "[all]"})
	require.NoError(t, err)

	t.Run("mcp to static", func(t *testing.T) {
		w := f.putInventory(t, mcp, fmt.Sprintf(`{"id":%d,"name":"x","project_id":%d,"type":"static","inventory":"[all]"}`, mcp.ID, f.project.ID))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		read, err := f.store.GetInventory(f.project.ID, mcp.ID)
		require.NoError(t, err)
		assert.Equal(t, db.InventoryMcp, read.Type)
	})

	t.Run("static to mcp", func(t *testing.T) {
		w := f.putInventory(t, static, fmt.Sprintf(`{"id":%d,"name":"x","project_id":%d,"type":"mcp","mcp_servers":[]}`, static.ID, f.project.ID))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		read, err := f.store.GetInventory(f.project.ID, static.ID)
		require.NoError(t, err)
		assert.Equal(t, db.InventoryStatic, read.Type)
	})
}

func TestUpdateInventory_McpIDMustMatchURL(t *testing.T) {
	f := newResourceFixture(t)
	w, inv := f.addMcpInventory(t, `[]`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	w = f.putInventory(t, inv, fmt.Sprintf(`{"id":%d,"name":"x","project_id":%d,"type":"mcp","mcp_servers":[]}`, inv.ID+100, f.project.ID))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
