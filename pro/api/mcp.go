package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func NewMCPController(enabled bool, store db.Store, rest http.Handler, webPath string) pro_interfaces.MCPController {
	return &mcpControllerImpl{}
}

type mcpControllerImpl struct {
}

func (ctrl *mcpControllerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
