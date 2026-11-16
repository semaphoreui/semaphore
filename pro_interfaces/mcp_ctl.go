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
	// ToolsHandler serves GET /api/mcp/tools, a REST route behind the session or
	// API-token sign-in: the tools a connection of the signed-in user to a set
	// would get. A build without MCP and a server with MCP switched off answer
	// 501 Not Implemented.
	ToolsHandler() http.Handler
}

// MCPConfig is what the built-in MCP server takes from the server
// configuration when it is built. The configuration does not change while the
// server runs; what may change and must apply at once on every node — the MCP
// settings, the license, a user's permissions — the server reads on every
// request instead.
type MCPConfig struct {
	// Enabled switches the MCP endpoints on; off, they answer 501.
	Enabled bool
	// WebPath is the path prefix the router serves the API under, ending in "/".
	WebPath string
	// WebHost is the public address of the web UI; empty when not configured.
	WebHost string
	// NonAdminCanCreateProject lets users who are not administrators create projects.
	NonAdminCanCreateProject bool
	// ScheduleTimezone is the time zone of the cron expressions of schedules; empty means UTC.
	ScheduleTimezone string
}
