package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMcpServers_Valid(t *testing.T) {
	servers := []McpServer{
		{Name: "github", URL: "https://api.githubcopilot.com/mcp/", Enabled: true},
		{Name: "docs_1", URL: "http://mcp.example.com:8080/sse"},
	}

	assert.NoError(t, ValidateMcpServers(servers))
	assert.NoError(t, ValidateMcpServers(nil))
}

func TestMcpServer_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		server  McpServer
		errPart string
	}{
		{"empty name", McpServer{Name: "", URL: "https://a"}, "name can not be empty"},
		{"name with space", McpServer{Name: "my server", URL: "https://a"}, "only letters"},
		{"name with quote", McpServer{Name: `a"b`, URL: "https://a"}, "only letters"},
		{"name with dot", McpServer{Name: "a.b", URL: "https://a"}, "only letters"},
		{"too long name", McpServer{Name: strings.Repeat("a", 65), URL: "https://a"}, "longer than 64"},
		{"empty url", McpServer{Name: "a", URL: ""}, "url can not be empty"},
		{"stdio command", McpServer{Name: "a", URL: "npx -y server"}, "http or https"},
		{"file scheme", McpServer{Name: "a", URL: "file:///etc/passwd"}, "http or https"},
		{"no host", McpServer{Name: "a", URL: "https:///path"}, "must have a host"},
		{"credentials in url", McpServer{Name: "a", URL: "https://user:pass@a.com"}, "credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorContains(t, tt.server.Validate(), tt.errPart)
		})
	}
}

func TestValidateMcpServers_DuplicateName(t *testing.T) {
	err := ValidateMcpServers([]McpServer{
		{Name: "a", URL: "https://a"},
		{Name: "a", URL: "https://b"},
	})

	assert.ErrorContains(t, err, "not unique")
}

func TestValidateMcpServers_ReportsPosition(t *testing.T) {
	err := ValidateMcpServers([]McpServer{
		{Name: "ok", URL: "https://a"},
		{Name: "bad", URL: "ftp://b"},
	})

	assert.ErrorContains(t, err, "mcp server #2")
}
