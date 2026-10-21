package pro_interfaces

import "net/http"

// MCPController serves the built-in MCP endpoint. A request it cannot serve —
// MCP switched off, no access, or a build without MCP — gets the same
// 501 Not Implemented, so a caller cannot tell these cases apart.
type MCPController interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
