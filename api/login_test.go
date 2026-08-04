package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailVerifiedClaim(t *testing.T) {
	tests := []struct {
		name          string
		claims        map[string]any
		expectedValue bool
		expectedPres  bool
	}{
		{"bool true", map[string]any{"email_verified": true}, true, true},
		{"bool false", map[string]any{"email_verified": false}, false, true},
		{"string true", map[string]any{"email_verified": "true"}, true, true},
		{"string false", map[string]any{"email_verified": "false"}, false, true},
		{"absent claim", map[string]any{"email": "x@y.z"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, present := emailVerifiedClaim(tt.claims)
			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedPres, present)
		})
	}
}

func TestResolveEmailVerified(t *testing.T) {
	tests := []struct {
		name     string
		claims   map[string]any
		require  bool
		expected bool
	}{
		// require_verified_email off (default): absent claim is trusted, but an
		// explicit email_verified=false must NOT be trusted (takeover guard).
		{"off, absent claim trusted", map[string]any{"email": "x@y.z"}, false, true},
		{"off, explicit true", map[string]any{"email_verified": true}, false, true},
		{"off, explicit false not trusted", map[string]any{"email_verified": false}, false, false},
		{"off, string false not trusted", map[string]any{"email_verified": "false"}, false, false},
		// require_verified_email on: claim must be explicitly present and true.
		{"on, absent claim not trusted", map[string]any{"email": "x@y.z"}, true, false},
		{"on, explicit true", map[string]any{"email_verified": true}, true, true},
		{"on, explicit false", map[string]any{"email_verified": false}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := util.OidcProvider{RequireVerifiedEmail: tt.require}
			assert.Equal(t, tt.expected, resolveEmailVerified(tt.claims, provider))
		})
	}
}

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

func TestIsSecureWebHost(t *testing.T) {
	orig := util.WebHostURL
	defer func() { util.WebHostURL = orig }()

	tests := []struct {
		name     string
		webHost  string
		expected bool
	}{
		{"https host is secure", "https://semaphore.example.com", true},
		{"http host is not secure", "http://semaphore.example.com:3000", false},
		{"nil host is not secure", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.webHost == "" {
				util.WebHostURL = nil
			} else {
				u, err := url.Parse(tt.webHost)
				require.NoError(t, err)
				util.WebHostURL = u
			}

			assert.Equal(t, tt.expected, isSecureWebHost())
		})
	}
}

func TestGenerateStateOauthCookie(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := "/dashboard"

	stateStr := generateStateOauthCookie(w, returnPath, false)

	// Test 1: Verify returned state is valid base64
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	// Test 2: Verify state contains valid JSON
	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Test 3: Verify return path is preserved
	assert.Equal(t, returnPath, state.Return, "Return path should be preserved")

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

	stateStr := generateStateOauthCookie(w, returnPath, false)

	// Decode and verify state
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Verify empty return path is preserved
	assert.Empty(t, state.Return, "Return path should be empty")
}

func TestOidcLoginLinkMode(t *testing.T) {
	origConfig := util.Config
	defer func() { util.Config = origConfig }()
	util.Config = &util.ConfigType{
		OidcProviders: map[string]util.OidcProvider{"test": {}},
	}

	tests := []struct {
		name     string
		method   string
		expected int
	}{
		// CSRF protection: SameSite=Lax sends the session cookie on top-level
		// cross-site GET navigations, so link mode must reject GET.
		{"GET link is rejected", http.MethodGet, http.StatusMethodNotAllowed},
		{"POST link without session is unauthorized", http.MethodPost, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/auth/oidc/test/login?link=1", nil)
			req = mux.SetURLVars(req, map[string]string{"provider": "test"})
			w := httptest.NewRecorder()

			oidcLogin(w, req)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestGenerateStateOauthCookieUniqueness(t *testing.T) {
	// Generate two states and verify they have different CSRF tokens
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	state1Str := generateStateOauthCookie(w1, "/path1", false)
	state2Str := generateStateOauthCookie(w2, "/path2", false)

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

func TestOidcClaimsFromUserInfo_UserInfoFailure(t *testing.T) {
	_, err := oidcClaimsFromUserInfo(nil, errors.New("userinfo unavailable"), util.OidcProvider{})
	require.Error(t, err)
	assert.EqualError(t, err, "userinfo unavailable")
}

func TestOidcClaimsFromUserInfo_WithEmail(t *testing.T) {
	userInfo := &oidc.UserInfo{
		Subject: "sub-123",
		Profile: "Jane Doe",
		Email:   "jane@example.com",
	}

	claims, err := oidcClaimsFromUserInfo(userInfo, nil, util.OidcProvider{})
	require.NoError(t, err)
	assert.Equal(t, "sub-123", claims.sub)
	assert.Equal(t, "jane@example.com", claims.email)
	assert.Equal(t, "Jane Doe", claims.name)
	assert.NotEmpty(t, claims.username)
}
