package pro_interfaces

import "net/http"

// MCPController serves the built-in MCP server: two tool sets, each on its own
// endpoint. A build without MCP and a server with MCP switched off answer 501
// Not Implemented on both; any failed sign-in gets one and the same 401
// Unauthorized.
type MCPController interface {
	// ProjectHandler serves the tool set for the work in projects, at /api/mcp/project.
	ProjectHandler() http.Handler
	// AdminHandler serves the tool set for the administration of the server, at /api/mcp/admin.
	AdminHandler() http.Handler
}
