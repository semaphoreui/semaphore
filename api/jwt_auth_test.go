package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/semaphoreui/semaphore/util"
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
