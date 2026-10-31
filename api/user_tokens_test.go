package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIToken_IssuesWorkingToken(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, _ := userRequest(store, http.MethodPost, "/api/user/tokens", `{"name":"ci"}`, user, user)
	w := httptest.NewRecorder()
	createAPIToken(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	var token db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &token))
	assert.Regexp(t, `^[a-z0-9_-]{43}=$`, token.ID)

	ok, req := authenticationHandler(httptest.NewRecorder(), bearerRequest(store, token.ID))
	require.True(t, ok)
	assert.Equal(t, user.ID, helpers.UserFromContext(req).ID)
}

func TestCreateAPIToken_IgnoresTypeInBody(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, _ := userRequest(store, http.MethodPost, "/api/user/tokens", `{"name":"ci","type":"mcp"}`, user, user)
	w := httptest.NewRecorder()
	createAPIToken(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fields))
	assert.NotContains(t, fields, "type", "the token type is not sent to clients")
	var token db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &token))
	stored, err := store.GetAPIToken(token.ID)
	require.NoError(t, err)
	assert.Equal(t, db.APITokenTypeAPI, stored.Type, "the REST API issues REST API tokens whatever the body says")
}

// MCP tools call REST with the token of the MCP request, so REST takes MCP
// tokens as well; the audit tells them apart.
func TestAuthentication_MCPTokenActor(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "mcptoken1234567890", UserID: user.ID, Name: "agent"})
	require.NoError(t, err)

	r, rec := withAuditRecorder(bearerRequest(store, "mcptoken1234567890"))
	ok, req := authenticationHandler(httptest.NewRecorder(), r)

	require.True(t, ok)
	assert.Empty(t, rec.All())
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthMCPToken, audit.TokenFingerprint("mcptoken1234567890")), audit.ActorFrom(req.Context()))
}

func TestAuthentication_ExpiredMCPTokenIsRefused(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	past := time.Now().Add(-time.Hour)
	_, err := store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "mcptoken1234567890", UserID: user.ID, ExpiresAt: &past, Name: "agent"})
	require.NoError(t, err)

	r, rec := withAuditRecorder(bearerRequest(store, "mcptoken1234567890"))
	w := httptest.NewRecorder()
	ok, _ := authenticationHandler(w, r)

	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.ReasonTokenExpired, onlyEvent(t, rec, audit.AuthAPITokenReject).Event.Reason)
}

func TestGetAPITokens_HidesMCPTokens(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateAPIToken(db.APITokenTypeAPI, db.APIToken{ID: "resttoken1234567890", UserID: user.ID, Name: "rest"})
	require.NoError(t, err)
	_, err = store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "mcptoken1234567890", UserID: user.ID, Name: "agent"})
	require.NoError(t, err)

	r, _ := userRequest(store, http.MethodGet, "/api/user/tokens", "", user, user)
	w := httptest.NewRecorder()
	getAPITokens(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var tokens []db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tokens))
	require.Len(t, tokens, 1)
	assert.Equal(t, "rest", tokens[0].Name)
}

func TestDeleteAPIToken_LeavesMCPTokens(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateAPIToken(db.APITokenTypeAPI, db.APIToken{ID: "abcdefgh1111", UserID: user.ID, Name: "rest"})
	require.NoError(t, err)
	_, err = store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "abcdefgh2222", UserID: user.ID, Name: "agent"})
	require.NoError(t, err)

	r, rec := userRequest(store, http.MethodDelete, "/api/user/tokens/x", "", user, user)
	w := httptest.NewRecorder()
	deleteAPIToken(w, mux.SetURLVars(r, map[string]string{"token_id": "abcdefgh"}))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "rest", onlyEvent(t, rec, audit.IAMAPITokenDelete).Event.Target.Name)
	_, err = store.GetAPIToken("abcdefgh1111")
	require.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetAPIToken("abcdefgh2222")
	assert.NoError(t, err, "a REST revoke by prefix does not touch an MCP token")
}
