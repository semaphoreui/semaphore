package db

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// mcpServerNameMaxLength bounds a server name: it becomes a config key and
// part of an environment variable name of the agent process.
const mcpServerNameMaxLength = 64

// McpServer is one remote MCP server of an InventoryMcp inventory.
type McpServer struct {
	// Name identifies the server to the agent. Unique within the inventory.
	Name string `json:"name"`
	// URL is the HTTP or SSE endpoint of the server.
	URL string `json:"url"`
	// Token is not the token itself but the uid of the inventory secret that
	// holds it. Empty means the server needs no authorization.
	Token string `json:"token,omitempty"`
}

// ParseMcpServers decodes and validates the content of an InventoryMcp
// inventory. An empty content is an empty list.
func ParseMcpServers(content string) ([]McpServer, error) {
	if strings.TrimSpace(content) == "" {
		return []McpServer{}, nil
	}

	var servers []McpServer
	if err := json.Unmarshal([]byte(content), &servers); err != nil {
		return nil, common_errors.NewValidationError("mcp inventory must be a JSON array of servers: " + err.Error())
	}

	if servers == nil {
		servers = []McpServer{}
	}

	if err := ValidateMcpServers(servers); err != nil {
		return nil, err
	}

	return servers, nil
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
	// tokens belong in an inventory secret.
	if u.User != nil {
		return common_errors.NewValidationError("url must not contain credentials, use a token instead")
	}

	return nil
}
