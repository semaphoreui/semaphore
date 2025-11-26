package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseClaim(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("email | {{ .id }}@test.com", claims)

	if !ok {
		t.Fail()
	}

	if res != "1234567@test.com" {
		t.Fatalf("%s must be %d@test.com", res, claims["id"])
	}
}

func TestParseClaim2(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("username", claims)

	if !ok {
		t.Fail()
	}

	if res != claims["username"] {
		t.Fail()
	}
}

func TestParseClaim3(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("email", claims)

	if ok {
		t.Fail()
	}
}

func TestParseClaim4(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("|", claims)

	if ok {
		t.Fail()
	}
}

func TestParseClaim5(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       123456757343.0,
	}

	prepareClaims(claims)

	res, ok := parseClaim("{{ .id }}", claims)

	if !ok || res != "123456757343" {
		t.Fatalf("Expected: %v, Got: %v", "123456757343", res)
	}
}

func TestGenerateStateOauthCookie(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := "/dashboard"

	stateStr := generateStateOauthCookie(w, returnPath)

	// Test 1: Verify returned state is valid base64
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	if err != nil {
		t.Fatalf("Returned state is not valid base64: %v", err)
	}

	// Test 2: Verify state contains valid JSON
	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		t.Fatalf("State does not contain valid JSON: %v", err)
	}

	// Test 3: Verify return path is preserved
	if state.Return != returnPath {
		t.Fatalf("Expected return path %s, got %s", returnPath, state.Return)
	}

	// Test 4: Verify CSRF token is not empty
	if state.Csrf == "" {
		t.Fatal("CSRF token should not be empty")
	}

	// Test 5: Verify CSRF token is valid base64
	_, err = base64.URLEncoding.DecodeString(state.Csrf)
	if err != nil {
		t.Fatalf("CSRF token is not valid base64: %v", err)
	}

	// Test 6: Verify cookie is set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies were set")
	}

	// Test 7: Verify cookie has correct name
	var oauthCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauthstate" {
			oauthCookie = cookie
			break
		}
	}
	if oauthCookie == nil {
		t.Fatal("Cookie 'oauthstate' was not set")
	}

	// Test 8: Verify cookie value matches CSRF token in state
	if oauthCookie.Value != state.Csrf {
		t.Fatalf("Cookie value %s does not match CSRF token %s", oauthCookie.Value, state.Csrf)
	}

	// Test 9: Verify cookie has expiration set (should be ~365 days)
	if oauthCookie.Expires.IsZero() {
		t.Fatal("Cookie expiration should be set")
	}

	expectedExpiration := time.Now().Add(365 * 24 * time.Hour)
	timeDiff := oauthCookie.Expires.Sub(expectedExpiration)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	// Allow 5 seconds tolerance for test execution time
	if timeDiff > 5*time.Second {
		t.Fatalf("Cookie expiration is off by more than 5 seconds: %v", timeDiff)
	}
}

func TestGenerateStateOauthCookieEmptyReturnPath(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := ""

	stateStr := generateStateOauthCookie(w, returnPath)

	// Decode and verify state
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	if err != nil {
		t.Fatalf("Returned state is not valid base64: %v", err)
	}

	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		t.Fatalf("State does not contain valid JSON: %v", err)
	}

	// Verify empty return path is preserved
	if state.Return != "" {
		t.Fatalf("Expected empty return path, got %s", state.Return)
	}
}

func TestGenerateStateOauthCookieUniqueness(t *testing.T) {
	// Generate two states and verify they have different CSRF tokens
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	state1Str := generateStateOauthCookie(w1, "/path1")
	state2Str := generateStateOauthCookie(w2, "/path2")

	// Decode states
	state1Bytes, _ := base64.URLEncoding.DecodeString(state1Str)
	state2Bytes, _ := base64.URLEncoding.DecodeString(state2Str)

	var state1, state2 oAuthState
	json.Unmarshal(state1Bytes, &state1)
	json.Unmarshal(state2Bytes, &state2)

	// Verify CSRF tokens are different
	if state1.Csrf == state2.Csrf {
		t.Fatal("Multiple calls should generate different CSRF tokens")
	}

	// Verify states are different
	if state1Str == state2Str {
		t.Fatal("Multiple calls should generate different state strings")
	}
}
