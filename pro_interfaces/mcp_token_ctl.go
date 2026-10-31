package pro_interfaces

import "net/http"

// MCPTokenController serves the routes where a user lists, issues and revokes
// their own MCP tokens. A build without MCP answers 404 on all of them.
type MCPTokenController interface {
	GetTokens(w http.ResponseWriter, r *http.Request)
	CreateToken(w http.ResponseWriter, r *http.Request)
	DeleteToken(w http.ResponseWriter, r *http.Request)
}
