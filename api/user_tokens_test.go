package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotContains(t, body, "type", "the token type is not sent to clients")
	_, err := store.GetAPIToken(body["id"].(string), db.APITokenTypeAPI)
	assert.NoError(t, err, "the REST API issues REST API tokens whatever the body says")
}

func TestAuthentication_MCPTokenIsUnknownToREST(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "mcptoken1234567890", UserID: user.ID, Name: "agent"})
	require.NoError(t, err)

	r, rec := withAuditRecorder(bearerRequest(store, "mcptoken1234567890"))
	w := httptest.NewRecorder()
	ok, _ := authenticationHandler(w, r)

	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	got := onlyEvent(t, rec, audit.AuthAPITokenReject)
	assert.Equal(t, audit.ReasonTokenUnknown, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetAPIToken, ID: audit.TokenFingerprint("mcptoken1234567890")}, got.Event.Target)
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
	_, err = store.GetAPIToken("abcdefgh1111", db.APITokenTypeAPI)
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetAPIToken("abcdefgh2222", db.APITokenTypeMCP)
	assert.NoError(t, err, "a REST revoke by prefix does not touch an MCP token")
}
