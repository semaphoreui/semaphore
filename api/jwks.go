package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/pkg/jwt"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type JwksController struct {
	signer jwt.Signer
}

func NewJwksController(signer jwt.Signer) *JwksController {
	return &JwksController{signer: signer}
}

// GetJWKS serves the JSON Web Key Set.
func (c *JwksController) GetJWKS(w http.ResponseWriter, _ *http.Request) {
	if util.Config == nil || util.Config.JWT == nil || !util.Config.JWT.Enabled {
		http.NotFound(w, nil)
		return
	}

	if c.signer == nil {
		http.NotFound(w, nil)
		return
	}

	body, err := c.signer.JWKS()
	if err != nil {
		log.WithError(err).WithField("context", "jwt").Error("failed to marshal JWKS")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}
