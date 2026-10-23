package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// NewMCPTokenController returns the MCP token routes of a build without MCP:
// every one of them answers 404.
func NewMCPTokenController(enabled bool, subscription pro_interfaces.SubscriptionService) pro_interfaces.MCPTokenController {
	return &mcpTokenControllerImpl{}
}

type mcpTokenControllerImpl struct {
}

func (c *mcpTokenControllerImpl) GetTokens(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *mcpTokenControllerImpl) CreateToken(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *mcpTokenControllerImpl) DeleteToken(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
