package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpTestToken = "mcptoken1234567890"

type activeLicense struct {
	pro_interfaces.SubscriptionService
}

func (activeLicense) GetToken() (pro_interfaces.SubscriptionToken, error) {
	return pro_interfaces.SubscriptionToken{State: "active", Plan: "pro4"}, nil
}

type failingLicense struct {
	pro_interfaces.SubscriptionService
}

func (failingLicense) GetToken() (pro_interfaces.SubscriptionToken, error) {
	return pro_interfaces.SubscriptionToken{}, errors.New("database is down")
}

type mcpTokensFixture struct {
	store  *sql.SqlDb
	router *mux.Router
	alice  db.User
}

// newMCPTokensFixture gives MCP to every user when mcpAllowed is set, and to nobody otherwise.
func newMCPTokensFixture(t *testing.T, enabled bool, mcpAllowed bool) mcpTokensFixture {
	t.Helper()
	return newMCPTokensFixtureOf(t, enabled, activeLicense{}, mcpAllowed)
}

func newMCPTokensFixtureOf(t *testing.T, enabled bool, license pro_interfaces.SubscriptionService, mcpAllowed bool) mcpTokensFixture {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	alice, err := store.CreateUserWithoutPassword(db.User{Username: "alice", Name: "Alice", Email: "alice@e.st"})
	require.NoError(t, err)

	controller := newMCPTokenController(enabled, license)
	controller.getFeatures = func(*db.User, string) pro_interfaces.Features {
		return pro_interfaces.Features{MCP: mcpAllowed}
	}
	router := mux.NewRouter()
	router.Path("/api/user/mcp-tokens").HandlerFunc(controller.GetTokens).Methods(http.MethodGet)
	router.Path("/api/user/mcp-tokens").HandlerFunc(controller.CreateToken).Methods(http.MethodPost)
	router.Path("/api/user/mcp-tokens/{token_id}").HandlerFunc(controller.DeleteToken).Methods(http.MethodDelete)
	return mcpTokensFixture{store: store, router: router, alice: alice}
}

// send performs the request as alice; authentication is not under test here.
func (f mcpTokensFixture) send(method string, path string, body string) (*httptest.ResponseRecorder, *audittest.Recorder) {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", &f.alice)
	rec := &audittest.Recorder{}
	r = helpers.SetContextValue(r, "audit", rec)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	return w, rec
}

func (f mcpTokensFixture) tokens(t *testing.T) []db.APIToken {
	t.Helper()
	tokens, err := f.store.GetAPITokens(f.alice.ID, db.APITokenTypeMCP)
	require.NoError(t, err)
	return tokens
}

func TestMCPTokens_IssueListRevoke(t *testing.T) {
	f := newMCPTokensFixture(t, true, true)

	w, rec := f.send(http.MethodPost, "/api/user/mcp-tokens", `{"name":"agent"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fields))
	assert.NotContains(t, fields, "type", "the token type is not sent to clients")
	var issued db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &issued))
	tokenID := issued.ID
	assert.Regexp(t, `^[a-z0-9_-]{43}=$`, tokenID)

	stored, err := f.store.GetAPIToken(tokenID)
	require.NoError(t, err)
	assert.Equal(t, db.APITokenTypeMCP, stored.Type, "the route issues an MCP token")
	created, err := rec.Only(audit.IAMAPITokenCreate)
	require.NoError(t, err)
	assert.Equal(t, &audit.Target{Type: audit.TargetMCPToken, ID: audit.TokenFingerprint(tokenID), Name: "agent"}, created.Event.Target)

	w, _ = f.send(http.MethodGet, "/api/user/mcp-tokens", "")
	require.Equal(t, http.StatusOK, w.Code)
	var listed []db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed, 1)
	assert.Equal(t, tokenID[:8], listed[0].ID, "the list shows a prefix only")

	w, rec = f.send(http.MethodDelete, "/api/user/mcp-tokens/"+tokenID[:8], "")
	assert.Equal(t, http.StatusNoContent, w.Code)
	deleted, err := rec.Only(audit.IAMAPITokenDelete)
	require.NoError(t, err)
	assert.Equal(t, &audit.Target{Type: audit.TargetMCPToken, ID: audit.TokenFingerprint(tokenID), Name: "agent"}, deleted.Event.Target)
	_, err = f.store.GetAPIToken(tokenID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// The REST API accepts an MCP token whatever the license, so its owner revokes it whatever the
// license, the seat and the switch say.
func TestMCPTokens_ListedAndRevokedWithoutMCP(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		mcpAllowed bool
	}{
		{"MCP not allowed by the license", true, false},
		{"MCP switched off", false, true},
		{"both", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMCPTokensFixture(t, tt.enabled, tt.mcpAllowed)
			_, err := f.store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: mcpTestToken, UserID: f.alice.ID, Name: "agent"})
			require.NoError(t, err)

			w, _ := f.send(http.MethodGet, "/api/user/mcp-tokens", "")
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), mcpTestToken[:8])

			w, rec := f.send(http.MethodDelete, "/api/user/mcp-tokens/"+mcpTestToken[:8], "")
			assert.Equal(t, http.StatusNoContent, w.Code)
			_, err = rec.Only(audit.IAMAPITokenDelete)
			require.NoError(t, err)
			assert.Empty(t, f.tokens(t))
		})
	}
}

func TestMCPTokens_NotIssuedWhileMCPIsOff(t *testing.T) {
	f := newMCPTokensFixture(t, false, true)

	w, rec := f.send(http.MethodPost, "/api/user/mcp-tokens", `{"name":"agent"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"MCP is switched off on this server"}`, w.Body.String(), "a JSON error, as every other refusal")
	assert.Empty(t, rec.All())
	assert.Empty(t, f.tokens(t))
}

func TestMCPTokens_IssuingNeedsMCPInTheLicense(t *testing.T) {
	f := newMCPTokensFixture(t, true, false)

	w, rec := f.send(http.MethodPost, "/api/user/mcp-tokens", `{"name":"agent"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), pro_interfaces.MCPNeedsLicense)
	assert.Empty(t, f.tokens(t))
	denied, err := rec.Only(audit.AuthAuthorizationDeny)
	require.NoError(t, err)
	assert.Equal(t, audit.ReasonForbidden, denied.Event.Reason)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: pro_interfaces.MCPLicenseFeature}, denied.Event.Metadata)
}

// The plan and the user's seat decide; the build's feature table is asked with both.
func TestMCPTokens_IssuingAsksTheFeatureTableWithThePlanAndTheUser(t *testing.T) {
	f := newMCPTokensFixture(t, true, true)
	controller := newMCPTokenController(true, activeLicense{})
	var askedUser string
	var askedPlan string
	controller.getFeatures = func(user *db.User, plan string) pro_interfaces.Features {
		askedUser, askedPlan = user.Username, plan
		return pro_interfaces.Features{}
	}
	f.router.Path("/check").HandlerFunc(controller.CreateToken).Methods(http.MethodPost)

	w, _ := f.send(http.MethodPost, "/check", `{"name":"agent"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "alice", askedUser)
	assert.Equal(t, "pro4", askedPlan)
}

func TestMCPTokens_FailureToReadTheLicenseIssuesNoToken(t *testing.T) {
	f := newMCPTokensFixtureOf(t, true, failingLicense{}, true)

	w, rec := f.send(http.MethodPost, "/api/user/mcp-tokens", `{"name":"agent"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, rec.All())
	assert.Empty(t, f.tokens(t))
}

func TestMCPTokens_DoNotTouchAPITokens(t *testing.T) {
	f := newMCPTokensFixture(t, true, true)
	_, err := f.store.CreateAPIToken(db.APITokenTypeAPI, db.APIToken{ID: "abcdefgh1111", UserID: f.alice.ID, Name: "rest"})
	require.NoError(t, err)
	_, err = f.store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "abcdefgh2222", UserID: f.alice.ID, Name: "agent"})
	require.NoError(t, err)

	w, _ := f.send(http.MethodGet, "/api/user/mcp-tokens", "")
	var listed []db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed, 1)
	assert.Equal(t, "agent", listed[0].Name, "the list shows MCP tokens only")

	w, rec := f.send(http.MethodDelete, "/api/user/mcp-tokens/abcdefgh", "")
	assert.Equal(t, http.StatusNoContent, w.Code)
	deleted, err := rec.Only(audit.IAMAPITokenDelete)
	require.NoError(t, err)
	assert.Equal(t, "agent", deleted.Event.Target.Name)
	_, err = f.store.GetAPIToken("abcdefgh1111")
	assert.NoError(t, err, "an MCP revoke by prefix does not touch a REST API token")
}

func TestMCPTokens_AreTheUsersOwn(t *testing.T) {
	f := newMCPTokensFixture(t, true, true)
	bob, err := f.store.CreateUserWithoutPassword(db.User{Username: "bob", Name: "Bob", Email: "bob@e.st"})
	require.NoError(t, err)
	_, err = f.store.CreateAPIToken(db.APITokenTypeMCP, db.APIToken{ID: "abcdefgh3333", UserID: bob.ID, Name: "bob-agent"})
	require.NoError(t, err)

	w, _ := f.send(http.MethodGet, "/api/user/mcp-tokens", "")
	require.Equal(t, http.StatusOK, w.Code)
	var listed []db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	assert.Empty(t, listed, "alice does not see bob's tokens")

	w, rec := f.send(http.MethodDelete, "/api/user/mcp-tokens/abcdefgh", "")
	assert.Equal(t, http.StatusNotFound, w.Code, "none of alice's tokens matches")
	assert.JSONEq(t, `{"error":"No MCP token of yours starts with this ID"}`, w.Body.String())
	assert.Empty(t, rec.All())
	_, err = f.store.GetAPIToken("abcdefgh3333")
	assert.NoError(t, err, "alice cannot revoke bob's token")
}

func TestMCPTokens_RefuseBadInput(t *testing.T) {
	f := newMCPTokensFixture(t, true, true)

	tests := []struct {
		name, method, path, body string
		code                     int
	}{
		{"expiry in the past", http.MethodPost, "/api/user/mcp-tokens", `{"name":"agent","expires_at":"2020-01-01T00:00:00Z"}`, http.StatusBadRequest},
		{"body is not JSON", http.MethodPost, "/api/user/mcp-tokens", `{"name":`, http.StatusBadRequest},
		// The store refuses a prefix shorter than 8 characters; helpers.WriteError
		// answers an unclassified error with 400, as on the REST API token route.
		{"prefix too short", http.MethodDelete, "/api/user/mcp-tokens/abc", "", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, rec := f.send(tt.method, tt.path, tt.body)
			assert.Equal(t, tt.code, w.Code)
			assert.Empty(t, rec.All())
		})
	}
	assert.Empty(t, f.tokens(t))
}
