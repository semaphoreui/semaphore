package db

import (
	"fmt"
	"net/url"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// mcpServerNameMaxLength bounds a server name: it becomes a config key and
// part of an environment variable name of the agent process.
const mcpServerNameMaxLength = 64

// McpServer is one remote MCP server of an InventoryMcp inventory, a row of
// project__inventory_mcp. Its authorization token, if any, is an access key
// with owner "mcp" and access_key.mcp_id pointing at this row.
type McpServer struct {
	ID          int `db:"id" json:"id" backup:"-"`
	InventoryID int `db:"inventory_id" json:"inventory_id" backup:"-"`
	// Name identifies the server to the agent. Unique within the inventory.
	Name string `db:"name" json:"name"`
	// URL is the HTTP or SSE endpoint of the server.
	URL string `db:"url" json:"url"`
	// Enabled servers are given to the agent; disabled ones are kept but skipped.
	Enabled bool `db:"enabled" json:"enabled"`
}

// ValidateMcpServers checks each server and that names are unique.
// Only remote servers (http/https) are supported.
func ValidateMcpServers(servers []McpServer) error {
	names := make(map[string]struct{}, len(servers))

	for i, s := range servers {
		if err := s.Validate(); err != nil {
			return common_errors.NewValidationError(fmt.Sprintf("mcp server #%d: %s", i+1, err.Error()))
		}

		if _, ok := names[s.Name]; ok {
			return common_errors.NewValidationError("mcp server name is not unique: " + s.Name)
		}
		names[s.Name] = struct{}{}
	}

	return nil
}

// Validate checks one server. The name is limited to letters, digits, '-' and
// '_' because it is written into the agent's config file and its environment;
// anything else could break out of that syntax.
func (s McpServer) Validate() error {
	if s.Name == "" {
		return common_errors.NewValidationError("name can not be empty")
	}

	if len(s.Name) > mcpServerNameMaxLength {
		return common_errors.NewValidationError(fmt.Sprintf("name can not be longer than %d characters", mcpServerNameMaxLength))
	}

	for _, r := range s.Name {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != '-' && r != '_' {
			return common_errors.NewValidationError("name may contain only letters, digits, '-' and '_'")
		}
	}

	if s.URL == "" {
		return common_errors.NewValidationError("url can not be empty")
	}

	u, err := url.Parse(s.URL)
	if err != nil {
		return common_errors.NewValidationError("url is invalid")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return common_errors.NewValidationError("url must use http or https")
	}

	if u.Host == "" {
		return common_errors.NewValidationError("url must have a host")
	}

	// A credential in the URL would be stored and shown in plain text;
	// tokens belong in the server's access key.
	if u.User != nil {
		return common_errors.NewValidationError("url must not contain credentials, use a token instead")
	}

	return nil
}
