package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func signTestJWT(t *testing.T, key *ecdsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = "test-key-1"

	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func setupTestJWTAuth(t *testing.T, key *ecdsa.PrivateKey, config *util.JWTAuthConfig) {
	t.Helper()

	jwk, err := jwkset.NewJWKFromKey(&key.PublicKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{
			KID: "test-key-1",
			ALG: jwkset.AlgES256,
			USE: jwkset.UseSig,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	storage := jwkset.NewMemoryStorage()
	if err := storage.KeyWrite(t.Context(), jwk); err != nil {
		t.Fatal(err)
	}

	kf, err := keyfunc.New(keyfunc.Options{
		Storage: storage,
	})
	if err != nil {
		t.Fatal(err)
	}

	oldKF := globalKeyfunc
	oldParser := globalJWTParser
	t.Cleanup(func() {
		globalKeyfunc = oldKF
		globalJWTParser = oldParser
	})
	globalKeyfunc = kf
	globalJWTParser = newJWTParser(config)
}

func setupEmptyJWTAuth(t *testing.T, config *util.JWTAuthConfig) {
	t.Helper()

	kf, err := keyfunc.New(keyfunc.Options{
		Storage: jwkset.NewMemoryStorage(),
	})
	if err != nil {
		t.Fatal(err)
	}

	oldKF := globalKeyfunc
	oldParser := globalJWTParser
	t.Cleanup(func() {
		globalKeyfunc = oldKF
		globalJWTParser = oldParser
	})
	globalKeyfunc = kf
	globalJWTParser = newJWTParser(config)
}

func TestValidateProxyJWT_Valid(t *testing.T) {
	key := generateTestKey(t)
	config := &util.JWTAuthConfig{
		Enabled:  true,
		Audience: "https://semaphore.example.com",
		Issuer:   "https://auth.example.com",
	}
	setupTestJWTAuth(t, key, config)

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"name":  "Test User",
		"aud":   "https://semaphore.example.com",
		"iss":   "https://auth.example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		"iat":   jwt.NewNumericDate(time.Now()),
	})

	claims, err := validateProxyJWT(token)
	if err != nil {
		t.Fatal("expected valid token, got error:", err)
	}

	if claims["email"] != "user@example.com" {
		t.Errorf("expected email user@example.com, got %v", claims["email"])
	}
}

func TestValidateProxyJWT_Expired(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{Enabled: true})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	})

	_, err := validateProxyJWT(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateProxyJWT_WrongAudience(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{
		Enabled:  true,
		Audience: "https://semaphore.example.com",
	})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"aud":   "https://wrong.example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	_, err := validateProxyJWT(token)
	if err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestValidateProxyJWT_WrongIssuer(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{
		Enabled: true,
		Issuer:  "https://auth.example.com",
	})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"iss":   "https://evil.example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	_, err := validateProxyJWT(token)
	if err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestValidateProxyJWT_UnknownKey(t *testing.T) {
	key := generateTestKey(t)
	setupEmptyJWTAuth(t, &util.JWTAuthConfig{Enabled: true})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	_, err := validateProxyJWT(token)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestValidateProxyJWT_MissingExp(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{Enabled: true})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
	})

	_, err := validateProxyJWT(token)
	if err == nil {
		t.Fatal("expected error for missing exp")
	}
}

func TestValidateProxyJWT_AudienceArray(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{
		Enabled:  true,
		Audience: "https://semaphore.example.com",
	})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"aud":   jwt.ClaimStrings{"https://other.example.com", "https://semaphore.example.com"},
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	claims, err := validateProxyJWT(token)
	if err != nil {
		t.Fatal("expected valid token with audience array, got error:", err)
	}

	if claims["email"] != "user@example.com" {
		t.Errorf("expected email user@example.com, got %v", claims["email"])
	}
}

func TestValidateProxyJWT_NoAudienceValidation(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{Enabled: true})

	token := signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@example.com",
		"aud":   "https://anything.example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	_, err := validateProxyJWT(token)
	if err != nil {
		t.Fatal("expected valid token when no audience configured, got error:", err)
	}
}

func TestJWTAuthConfig_Defaults(t *testing.T) {
	config := &util.JWTAuthConfig{}

	if config.GetHeader() != "" {
		t.Errorf("expected empty default header, got %s", config.GetHeader())
	}
	if config.GetEmailClaim() != "email" {
		t.Errorf("expected default email claim 'email', got %s", config.GetEmailClaim())
	}
	if config.GetNameClaim() != "name" {
		t.Errorf("expected default name claim 'name', got %s", config.GetNameClaim())
	}
	if config.GetUsernameClaim() != "email" {
		t.Errorf("expected default username claim 'email', got %s", config.GetUsernameClaim())
	}
}

func TestJWTAuthConfig_CustomValues(t *testing.T) {
	config := &util.JWTAuthConfig{
		Header:        "X-Custom-JWT",
		EmailClaim:    "mail",
		NameClaim:     "display_name",
		UsernameClaim: "preferred_username",
	}

	if config.GetHeader() != "X-Custom-JWT" {
		t.Errorf("expected header X-Custom-JWT, got %s", config.GetHeader())
	}
	if config.GetEmailClaim() != "mail" {
		t.Errorf("expected email claim 'mail', got %s", config.GetEmailClaim())
	}
	if config.GetNameClaim() != "display_name" {
		t.Errorf("expected name claim 'display_name', got %s", config.GetNameClaim())
	}
	if config.GetUsernameClaim() != "preferred_username" {
		t.Errorf("expected username claim 'preferred_username', got %s", config.GetUsernameClaim())
	}
}

func TestValidateProxyJWT_ParserNotInitialized(t *testing.T) {
	key := generateTestKey(t)
	setupTestJWTAuth(t, key, &util.JWTAuthConfig{Enabled: true})
	globalJWTParser = nil

	_, err := validateProxyJWT("not-a-real-token")
	if err == nil {
		t.Fatal("expected an error when the JWT parser is not initialized")
	}
}

func TestAuthentication_JWTActor(t *testing.T) {
	store := setupSessionTest(t)
	key := generateTestKey(t)
	config := &util.JWTAuthConfig{Enabled: true, Header: "X-Test-Jwt"}
	setupTestJWTAuth(t, key, config)
	util.Config.Auth = &util.AuthConfig{JWT: config}

	r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/user", nil), "store", store)
	r.Header.Set("X-Test-Jwt", signTestJWT(t, key, jwt.MapClaims{
		"sub":   "user123",
		"email": "jwtuser@example.com",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}))

	ok, req := authenticationHandler(httptest.NewRecorder(), r)

	require.True(t, ok)
	actor := audit.ActorFrom(req.Context())
	assert.Equal(t, audit.AuthProxyJWT, actor.Auth)
}

func setupJWTIdentityTest(t *testing.T, emailMatching string) (db.Store, *ecdsa.PrivateKey) {
	t.Helper()
	store := setupIdentityTest(t, emailMatching)
	key := generateTestKey(t)
	config := &util.JWTAuthConfig{Enabled: true, Header: "X-Test-Jwt"}
	setupTestJWTAuth(t, key, config)
	util.Config.Auth = &util.AuthConfig{JWT: config}
	return store, key
}

func jwtRequest(t *testing.T, store db.Store, key *ecdsa.PrivateKey, sub string, email string) *http.Request {
	t.Helper()
	claims := jwt.MapClaims{
		"email": email,
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	if sub != "" {
		claims["sub"] = sub
	}
	r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/user", nil), "store", store)
	r.Header.Set("X-Test-Jwt", signTestJWT(t, key, claims))
	return r
}

func TestAuthenticateByJWT_LinksIdentityBySub(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "auto")

	first, err := authenticateByJWT(jwtRequest(t, store, key, "sub-1", "jwtuser@example.com"))
	require.NoError(t, err)

	ids, err := store.GetUserExternalIdentities(first)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, db.IdentityTypeJwt, ids[0].Type)
	assert.Equal(t, "jwt", ids[0].Provider)
	assert.Equal(t, "sub-1", ids[0].ExternalUID)

	// The email changed at the proxy: the same sub finds the same user.
	second, err := authenticateByJWT(jwtRequest(t, store, key, "sub-1", "renamed@example.com"))
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestAuthenticateByJWT_RequiresSub(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "auto")

	_, err := authenticateByJWT(jwtRequest(t, store, key, "", "jwtuser@example.com"))
	assert.Error(t, err)
}

func TestAuthenticateByJWT_DoesNotAdoptPinnedUser(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "auto")

	pinned, _, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "oidc-sub",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", EmailVerified: true,
	})
	require.NoError(t, err)

	userID, err := authenticateByJWT(jwtRequest(t, store, key, "jwt-sub", "jdoe@example.com"))
	if err == nil {
		assert.NotEqual(t, pinned.ID, userID)
	}

	ids, err := store.GetUserExternalIdentities(pinned.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestAuthenticateByJWT_NeverMatchesByEmail(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "never")

	existing, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	userID, err := authenticateByJWT(jwtRequest(t, store, key, "jwt-sub", "jdoe@example.com"))
	if err == nil {
		assert.NotEqual(t, existing.ID, userID)
	}
}

func TestAuthenticateByJWT_AutoLinksExistingJWTUser(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "auto")

	// Created by an earlier version, which matched JWT users by email only.
	existing, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	userID, err := authenticateByJWT(jwtRequest(t, store, key, "jwt-sub", "jdoe@example.com"))
	require.NoError(t, err)
	assert.Equal(t, existing.ID, userID)
}

func TestAuthenticateByJWT_DoesNotAdoptLocalUser(t *testing.T) {
	store, key := setupJWTIdentityTest(t, "always")

	local, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	userID, err := authenticateByJWT(jwtRequest(t, store, key, "jwt-sub", "jdoe@example.com"))
	if err == nil {
		assert.NotEqual(t, local.ID, userID)
	}
}
