package jwt

import (
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allowedAlgs = []jose.SignatureAlgorithm{jose.ES256}

func newTestSigner(t *testing.T, opts SignerOptions) Signer {
	t.Helper()
	pemBytes, err := GenerateKeyPEM()
	require.NoError(t, err)
	s, err := NewECDSASignerFromPEM(pemBytes, opts)
	require.NoError(t, err)
	return s
}

func TestSigner_SignRoundtrip(t *testing.T) {
	s := newTestSigner(t, SignerOptions{
		Issuer:     "semaphore",
		DefaultTTL: time.Hour,
		MaxTTL:     time.Hour,
	})

	token, err := s.Sign(TaskInfo{
		TaskID:     42,
		ProjectID:  7,
		TemplateID: 3,
		Audience:   Audience{"openbao", "vault"},
	})
	require.NoError(t, err)

	parsed, err := josejwt.ParseSigned(token, allowedAlgs)
	require.NoError(t, err)

	var claims TaskClaims
	require.NoError(t, parsed.UnsafeClaimsWithoutVerification(&claims))

	assert.Equal(t, "semaphore", claims.Issuer)
	assert.Equal(t, "task:42", claims.Subject)
	assert.Equal(t, 42, claims.TaskID)
	assert.Equal(t, 7, claims.ProjectID)
	assert.Equal(t, 3, claims.TemplateID)
	assert.Equal(t, Audience{"openbao", "vault"}, claims.Audience)
	assert.NotEmpty(t, claims.JWTID)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestSigner_TTLClamping(t *testing.T) {
	s := newTestSigner(t, SignerOptions{
		DefaultTTL: time.Minute,
		MaxTTL:     time.Hour,
	})

	token, err := s.Sign(TaskInfo{
		TaskID: 1,
		TTL:    24 * time.Hour, // way above MaxTTL
	})
	require.NoError(t, err)

	parsed, err := josejwt.ParseSigned(token, allowedAlgs)
	require.NoError(t, err)

	var claims TaskClaims
	require.NoError(t, parsed.UnsafeClaimsWithoutVerification(&claims))

	lifetime := time.Duration(claims.ExpiresAt-claims.IssuedAt) * time.Second
	assert.Equal(t, time.Hour, lifetime)
}

func TestSigner_JWKSExposesKey(t *testing.T) {
	s := newTestSigner(t, SignerOptions{DefaultTTL: time.Hour, MaxTTL: time.Hour})

	jwks, err := s.JWKS()
	require.NoError(t, err)
	assert.Contains(t, string(jwks), `"kty":"EC"`)
	assert.Contains(t, string(jwks), `"crv":"P-256"`)
	assert.Contains(t, string(jwks), s.KeyID())
}
