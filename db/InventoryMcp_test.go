package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMcpServers_Valid(t *testing.T) {
	servers, err := ParseMcpServers(`[
		{"name": "github", "url": "https://api.githubcopilot.com/mcp/", "token": "q9r34543-qe74tas-wr74tegert"},
		{"name": "docs_1", "url": "http://mcp.example.com:8080/sse"}
	]`)

	require.NoError(t, err)
	assert.Equal(t, []McpServer{
		{Name: "github", URL: "https://api.githubcopilot.com/mcp/", Token: "q9r34543-qe74tas-wr74tegert"},
		{Name: "docs_1", URL: "http://mcp.example.com:8080/sse"},
	}, servers)
}

func TestParseMcpServers_Empty(t *testing.T) {
	for _, content := range []string{"", "  ", "[]", "null"} {
		t.Run(content, func(t *testing.T) {
			servers, err := ParseMcpServers(content)
			require.NoError(t, err)
			assert.Empty(t, servers)
			assert.NotNil(t, servers)
		})
	}
}

func TestParseMcpServers_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errPart string
	}{
		{"not json", `not json`, "JSON array"},
		{"object instead of array", `{"name": "a", "url": "https://a"}`, "JSON array"},
		{"empty name", `[{"name": "", "url": "https://a"}]`, "name can not be empty"},
		{"name with space", `[{"name": "my server", "url": "https://a"}]`, "only letters"},
		{"name with quote", `[{"name": "a\"b", "url": "https://a"}]`, "only letters"},
		{"name with dot", `[{"name": "a.b", "url": "https://a"}]`, "only letters"},
		{"too long name", `[{"name": "` + strings.Repeat("a", 65) + `", "url": "https://a"}]`, "longer than 64"},
		{"empty url", `[{"name": "a", "url": ""}]`, "url can not be empty"},
		{"stdio command", `[{"name": "a", "url": "npx -y server"}]`, "http or https"},
		{"file scheme", `[{"name": "a", "url": "file:///etc/passwd"}]`, "http or https"},
		{"no host", `[{"name": "a", "url": "https:///path"}]`, "must have a host"},
		{"credentials in url", `[{"name": "a", "url": "https://user:pass@a.com"}]`, "credentials"},
		{"duplicate name", `[{"name": "a", "url": "https://a"}, {"name": "a", "url": "https://b"}]`, "not unique"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMcpServers(tt.content)
			assert.ErrorContains(t, err, tt.errPart)
		})
	}
}
