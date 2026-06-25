package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClaim(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("email | {{ .id }}@test.com", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, "1234567@test.com", res, "Result should be formatted correctly")
}

func TestParseClaim2(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("username", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, claims["username"], res, "Result should match username claim")
}

func TestParseClaim3(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("email", claims)

	assert.False(t, ok, "parseClaim should fail for empty email")
}

func TestParseClaim4(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("|", claims)

	assert.False(t, ok, "parseClaim should fail for invalid pattern")
}

func TestParseClaim5(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       123456757343.0,
	}

	prepareClaims(claims)

	res, ok := parseClaim("{{ .id }}", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, "123456757343", res, "Result should match formatted ID")
}

func TestGenerateStateOauthCookie(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := "/dashboard"

	stateStr, nonce := generateStateOauthCookie(w, returnPath)

	// Verify nonce is returned and embedded in the state
	assert.NotEmpty(t, nonce, "Nonce should not be empty")

	// Test 1: Verify returned state is valid base64
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	// Test 2: Verify state contains valid JSON
	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Test 3: Verify return path is preserved
	assert.Equal(t, returnPath, state.Return, "Return path should be preserved")

	// Verify the returned nonce matches the one embedded in the state
	assert.Equal(t, nonce, state.Nonce, "Nonce in state should match returned nonce")

	// Test 4: Verify CSRF token is not empty
	assert.NotEmpty(t, state.Csrf, "CSRF token should not be empty")

	// Test 5: Verify CSRF token is valid base64
	_, err = base64.URLEncoding.DecodeString(state.Csrf)
	assert.NoError(t, err, "CSRF token should be valid base64")

	// Test 6: Verify cookie is set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "At least one cookie should be set")

	// Test 7: Verify cookie has correct name
	var oauthCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauthstate" {
			oauthCookie = cookie
			break
		}
	}
	assert.NotNil(t, oauthCookie, "Cookie 'oauthstate' should be set")

	// Test 8: Verify cookie value matches CSRF token in state
	assert.Equal(t, state.Csrf, oauthCookie.Value, "Cookie value should match CSRF token")

	// Test 9: Verify cookie has expiration set (should be ~365 days)
	assert.False(t, oauthCookie.Expires.IsZero(), "Cookie expiration should be set")

	expectedExpiration := time.Now().Add(365 * 24 * time.Hour)
	timeDiff := oauthCookie.Expires.Sub(expectedExpiration)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	// Allow 5 seconds tolerance for test execution time
	assert.LessOrEqual(t, timeDiff, 5*time.Second, "Cookie expiration should be within 5 seconds of expected")
}

func TestGenerateStateOauthCookieEmptyReturnPath(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := ""

	stateStr, _ := generateStateOauthCookie(w, returnPath)

	// Decode and verify state
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Verify empty return path is preserved
	assert.Empty(t, state.Return, "Return path should be empty")
}

func TestGenerateStateOauthCookieUniqueness(t *testing.T) {
	// Generate two states and verify they have different CSRF tokens
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	state1Str, nonce1 := generateStateOauthCookie(w1, "/path1")
	state2Str, nonce2 := generateStateOauthCookie(w2, "/path2")

	// Verify nonces are different between calls
	assert.NotEqual(t, nonce1, nonce2, "Multiple calls should generate different nonces")

	// Decode states
	state1Bytes, err1 := base64.URLEncoding.DecodeString(state1Str)
	state2Bytes, err2 := base64.URLEncoding.DecodeString(state2Str)
	assert.NoError(t, err1, "First state should be valid base64")
	assert.NoError(t, err2, "Second state should be valid base64")

	var state1, state2 oAuthState
	err1 = json.Unmarshal(state1Bytes, &state1)
	err2 = json.Unmarshal(state2Bytes, &state2)
	assert.NoError(t, err1, "First state should be valid JSON")
	assert.NoError(t, err2, "Second state should be valid JSON")

	// Verify CSRF tokens are different
	assert.NotEqual(t, state1.Csrf, state2.Csrf, "Multiple calls should generate different CSRF tokens")

	// Verify states are different
	assert.NotEqual(t, state1Str, state2Str, "Multiple calls should generate different state strings")
}

func TestSameIssuer(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"exact match", "https://idp.example.com", "https://idp.example.com", true},
		{"trailing slash on a", "https://idp.example.com/", "https://idp.example.com", true},
		{"trailing slash on b", "https://idp.example.com", "https://idp.example.com/", true},
		{"with path", "https://idp.example.com/realms/main", "https://idp.example.com/realms/main", true},
		{"different host", "https://idp.example.com", "https://evil.example.com", false},
		{"different path", "https://idp.example.com/a", "https://idp.example.com/b", false},
		{"empty a", "", "https://idp.example.com", false},
		{"empty b", "https://idp.example.com", "", false},
		{"both empty", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sameIssuer(tt.a, tt.b))
		})
	}
}

func TestCheckNonce(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		{"match", "abc", "abc", true},
		{"mismatch", "abc", "xyz", false},
		{"no nonce issued", "", "anything", true},
		{"provider dropped nonce", "abc", "", true},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkNonce(tt.expected, tt.actual))
		})
	}
}

func TestSafeReturnPath(t *testing.T) {
	setupOidcTestConfig()
	util.Config.WebHost = "https://semaphore.example.com"

	tests := []struct {
		name     string
		raw      string
		wantPath string
		wantOk   bool
	}{
		{"relative path", "/dashboard", "/dashboard", true},
		{"relative with query", "/dashboard?tab=1", "/dashboard?tab=1", true},
		{"relative without leading slash", "dashboard", "/dashboard", true},
		{"same-origin absolute", "https://semaphore.example.com/project/1", "/project/1", true},
		{"same-origin root", "https://semaphore.example.com", "/", true},
		{"same-origin case-insensitive host", "https://SEMAPHORE.example.com/x", "/x", true},
		{"foreign host", "https://evil.example.com/x", "", false},
		{"scheme-relative open redirect", "//evil.example.com/x", "", false},
		{"different scheme", "http://semaphore.example.com/x", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, ok := safeReturnPath(tt.raw)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantPath, path)
			}
		})
	}
}

// setupOidcTestConfig initializes util.Config with a single statically
// configured OIDC provider ("test") suitable for exercising the IdP-initiated
// flow without network access. The provider is built via JSON to populate the
// (unexported-typed) Endpoint field without depending on its concrete type.
func setupOidcTestConfig() {
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
	util.Config.WebHost = "https://semaphore.example.com"

	var provider util.OidcProvider
	err := json.Unmarshal([]byte(`{
		"client_id": "client123",
		"redirect_url": "https://semaphore.example.com/api/auth/oidc/test/redirect",
		"scopes": ["openid", "profile", "email"],
		"endpoint": {
			"issuer": "https://idp.example.com",
			"auth": "https://idp.example.com/authorize",
			"token": "https://idp.example.com/token"
		},
		"allow_idp_initiated": true
	}`), &provider)
	if err != nil {
		panic(err)
	}

	util.Config.OidcProviders = map[string]util.OidcProvider{"test": provider}
}

// callOidcInitiate runs the oidcInitiate handler for provider "test" with the
// given query string and returns the recorder.
func callOidcInitiate(t *testing.T, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/test/initiate?"+query, nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "test"})
	w := httptest.NewRecorder()
	oidcInitiate(w, req)
	return w
}

func TestOidcInitiate_Success(t *testing.T) {
	setupOidcTestConfig()

	w := callOidcInitiate(t, "iss=https://idp.example.com&login_hint=alice@example.com&target_link_uri=/project/7")

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)

	loc := w.Header().Get("Location")
	u, err := url.Parse(loc)
	require.NoError(t, err)

	// Redirects to the IdP authorization endpoint.
	assert.Equal(t, "idp.example.com", u.Host)
	assert.Equal(t, "/authorize", u.Path)

	q := u.Query()
	assert.Equal(t, "client123", q.Get("client_id"))
	assert.Equal(t, "alice@example.com", q.Get("login_hint"), "login_hint should be forwarded")
	assert.NotEmpty(t, q.Get("state"), "state should be present")
	assert.NotEmpty(t, q.Get("nonce"), "nonce should be present")

	// State cookie is set.
	var found bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "oauthstate" {
			found = true
		}
	}
	assert.True(t, found, "oauthstate cookie should be set")
}

func TestOidcInitiate_Rejections(t *testing.T) {
	tests := []struct {
		name  string
		query string
		setup func()
	}{
		{
			name:  "feature disabled",
			query: "iss=https://idp.example.com",
			setup: func() {
				setupOidcTestConfig()
				p := util.Config.OidcProviders["test"]
				p.AllowIdPInitiated = false
				util.Config.OidcProviders["test"] = p
			},
		},
		{
			name:  "missing iss",
			query: "",
			setup: setupOidcTestConfig,
		},
		{
			name:  "mismatched iss",
			query: "iss=https://evil.example.com",
			setup: setupOidcTestConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := callOidcInitiate(t, tt.query)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			loc := w.Header().Get("Location")
			assert.Equal(t, "https://semaphore.example.com/auth/login", loc,
				"rejected requests should redirect to the login page")

			for _, c := range w.Result().Cookies() {
				assert.NotEqual(t, "oauthstate", c.Name, "no auth flow should start on rejection")
			}
		})
	}
}

func TestOidcInitiate_UnknownProvider(t *testing.T) {
	setupOidcTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/missing/initiate?iss=https://idp.example.com", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "missing"})
	w := httptest.NewRecorder()
	oidcInitiate(w, req)

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://semaphore.example.com/auth/login", w.Header().Get("Location"))
}

func TestOidcInitiate_RejectsForeignTargetLinkURI(t *testing.T) {
	setupOidcTestConfig()

	// A foreign target_link_uri must be ignored, but the flow still starts.
	w := callOidcInitiate(t, "iss=https://idp.example.com&target_link_uri=https://evil.example.com/x")

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)
	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "idp.example.com/authorize", "flow should still start")
}
