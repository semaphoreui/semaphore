package pro_interfaces

import "net/http"

// MCPController serves the built-in MCP endpoint. A build without MCP and a
// server with MCP switched off answer 501 Not Implemented; any failed sign-in
// gets one and the same 401 Unauthorized.
type MCPController interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
