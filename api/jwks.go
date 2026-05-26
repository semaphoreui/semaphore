package api

import (
	"net/http"

	semjwt "github.com/semaphoreui/semaphore/pkg/jwt"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// jwksHandler serves the JSON Web Key Set for the active JWT signing key.
// External services (e.g. HashiCorp Vault's JWT auth method) point to this
// endpoint via jwks_url to verify Semaphore-issued task tokens.
// Returns 404 when JWT issuance is disabled.
func jwksHandler(w http.ResponseWriter, _ *http.Request) {
	if !util.Config.JWTEnabled {
		http.NotFound(w, nil)
		return
	}

	signer := semjwt.Default()
	if signer == nil {
		http.NotFound(w, nil)
		return
	}

	body, err := signer.JWKS()
	if err != nil {
		log.WithError(err).WithField("context", "jwt").Error("failed to marshal JWKS")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(body)
}
