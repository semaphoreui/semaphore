package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMcpToken stores an MCP server with a token the way the inventory API does.
func newMcpToken(t *testing.T, f resourceFixture) db.AccessKey {
	t.Helper()

	inv, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "mcp", Type: db.InventoryMcp})
	require.NoError(t, err)

	secret := "encrypted"
	_, err = f.store.SetMcpServers(f.project.ID, inv.ID, []db.McpServerChange{{
		Server: db.McpServer{Name: "github", URL: "https://api.githubcopilot.com/mcp/", Enabled: true},
		Token:  &db.McpTokenChange{Operation: db.McpTokenCreate, Key: db.AccessKey{Secret: &secret}},
	}})
	require.NoError(t, err)

	tokens, err := f.store.GetMcpServerTokens(f.project.ID, inv.ID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	return tokens[0]
}

func TestKeyMiddleware_HidesMcpTokens(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			f := newResourceFixture(t)
			token := newMcpToken(t, f)

			reached := false
			handler := KeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
			}))

			r, _ := f.request(method, "", nil, map[string]string{"key_id": strconv.Itoa(token.ID)})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.False(t, reached, "an MCP token is managed only through the inventory")
		})
	}
}

func TestKeyMiddleware_PassesSharedKeys(t *testing.T) {
	f := newResourceFixture(t)

	reached := false
	handler := KeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	}))

	r, _ := f.request(http.MethodGet, "", nil, map[string]string{"key_id": strconv.Itoa(f.key.ID)})
	handler.ServeHTTP(httptest.NewRecorder(), r)

	assert.True(t, reached)
}

func TestAddKey_RejectsMcpOwner(t *testing.T) {
	f, ctrl := newKeyFixture(t)
	body := `{"name":"t","type":"string","string":"s","owner":"mcp","project_id":` + strconv.Itoa(f.project.ID) + `}`
	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()

	ctrl.AddKey(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid key owner")
	assert.Empty(t, rec.All(), "nothing was created, nothing to audit")
}

func TestGetKeys_DoesNotListMcpTokens(t *testing.T) {
	f := newResourceFixture(t)
	token := newMcpToken(t, f)

	r, _ := f.request(http.MethodGet, "", nil, nil)
	w := httptest.NewRecorder()
	GetKeys(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"id":`+strconv.Itoa(token.ID)+`,`)
	assert.NotContains(t, w.Body.String(), "github")
}
