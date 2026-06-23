package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/pkg/jwt"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwksController_GetJWKS(t *testing.T) {
	signer := newTestSigner(t)
	controller := NewJwksController(signer)

	t.Run("enabled", func(t *testing.T) {
		util.Config = &util.ConfigType{
			JWT: &util.JWTConfig{Enabled: true},
		}

		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rr := httptest.NewRecorder()

		controller.GetJWKS(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), signer.KeyID())
	})

	t.Run("disabled", func(t *testing.T) {
		util.Config = &util.ConfigType{
			JWT: &util.JWTConfig{Enabled: false},
		}

		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rr := httptest.NewRecorder()

		controller.GetJWKS(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func newTestSigner(t *testing.T) jwt.Signer {
	t.Helper()

	pemBytes, err := jwt.GenerateKeyPEM()
	require.NoError(t, err)

	signer, err := jwt.NewECDSASignerFromPEM(pemBytes, jwt.SignerOptions{})
	require.NoError(t, err)

	return signer
}
