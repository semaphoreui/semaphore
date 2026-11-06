package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// NewMCPController returns the MCP endpoints of a build without MCP: every
// one of them answers 501 Not Implemented.
func NewMCPController(config pro_interfaces.MCPConfig, store db.Store, globalRoles db.GlobalRoleStore, subscription pro_interfaces.SubscriptionService, rest http.Handler) pro_interfaces.MCPController {
	return &mcpControllerImpl{}
}

type mcpControllerImpl struct {
}

func (ctrl *mcpControllerImpl) ProjectHandler() http.Handler {
	return http.HandlerFunc(mcpNotImplemented)
}

func (ctrl *mcpControllerImpl) AdminHandler() http.Handler {
	return http.HandlerFunc(mcpNotImplemented)
}

func (ctrl *mcpControllerImpl) ToolsHandler() http.Handler {
	return http.HandlerFunc(mcpToolsNotImplemented)
}

func mcpNotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}

// mcpToolsNotImplemented is the 501 of the REST route GET /api/mcp/tools: JSON, as every REST error.
func mcpToolsNotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	helpers.WriteErrorStatus(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
