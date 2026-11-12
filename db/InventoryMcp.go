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
	// HasToken tells whether the server has a token. Read-only: filled by
	// McpServerManager.GetMcpServers, never stored. The token itself is never
	// returned.
	HasToken bool `db:"has_token" json:"has_token" backup:"-"`
	// Token is the decrypted token, filled only for a task run
	// (InventoryService) and never serialized: not to the API, not to a backup.
	// A remote runner gets it as the access key TokenKeyID of the runner state.
	Token string `db:"-" json:"-" backup:"-"`
	// TokenKeyID is the ID of the token's access key, filled only for a task
	// run, so that a runner can find the token among the decrypted keys the
	// server sends with the job.
	TokenKeyID *int `db:"-" json:"token_key_id,omitempty" backup:"-"`
}

// McpServerWithToken is an MCP server together with the access key of its
// token, as they were at one moment: both come from one statement, so the
// token always belongs to the URL next to it.
type McpServerWithToken struct {
	Server McpServer
	// Token is nil for a server without a token. The secret is still encrypted.
	Token *AccessKey
}

// McpTokenOperation is what to do with the token of an MCP server.
type McpTokenOperation string

const (
	// McpTokenCreate gives a server without a token its token.
	McpTokenCreate McpTokenOperation = "create"
	// McpTokenUpdate replaces the value of the server's token.
	McpTokenUpdate McpTokenOperation = "update"
	// McpTokenDelete removes the server's token.
	McpTokenDelete McpTokenOperation = "delete"
)

// McpTokenChange is a change of the token of one MCP server.
type McpTokenChange struct {
	Operation McpTokenOperation
	// Key carries the encrypted secret (Type and Secret) for create and update.
	// Owner, McpID and ProjectID are set by the store, not taken from here.
	Key AccessKey
}

// McpServerChange is one server of the list given to SetMcpServers.
type McpServerChange struct {
	// Server with ID 0 is created; with an ID it updates that server, which
	// must belong to the inventory.
	Server McpServer
	// Token is nil when the token stays as it is.
	Token *McpTokenChange
}

func (c McpTokenChange) Validate() error {
	switch c.Operation {
	case McpTokenCreate, McpTokenUpdate:
		if c.Key.Secret == nil {
			return common_errors.NewValidationError("mcp token secret can not be empty")
		}
	case McpTokenDelete:
	default:
		return common_errors.NewValidationError("invalid mcp token operation: " + string(c.Operation))
	}

	return nil
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
