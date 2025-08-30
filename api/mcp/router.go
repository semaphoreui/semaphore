package mcp

import (
	"github.com/gorilla/mux"
	mcpservice "github.com/semaphoreui/semaphore/services/mcp"
)

// Route mounts MCP handlers under /mcp.
func Route(r *mux.Router, srv *mcpservice.Server) {
	if srv == nil {
		return
	}
	sub := r.PathPrefix("/mcp").Subrouter()
	sub.HandleFunc("/ws", srv.ServeWS).Methods("GET", "HEAD")
}
