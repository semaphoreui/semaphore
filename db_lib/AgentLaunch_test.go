package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentLaunch_SkipsDisabledServersAndUsesTokenVars(t *testing.T) {
	l, err := NewAgentLaunch(db.AppClaude, "/home/agent", agentInventory().McpServers)
	require.NoError(t, err)

	assert.Equal(t, []string{"SEMAPHORE_MCP_TOKEN_1=" + agentTestToken}, l.TokenEnv)
	assert.Equal(t, []string{agentTestToken}, l.Secrets)
	assert.Equal(t, 2, l.Servers)
	for _, f := range l.Files {
		assert.NotContains(t, string(f.Content), agentTestToken)
		assert.NotContains(t, string(f.Content), "disabled-token")
		assert.NotContains(t, string(f.Content), "mcp.example.com/off")
	}
}

func TestNewAgentLaunch_TokenNotDelivered(t *testing.T) {
	_, err := NewAgentLaunch(db.AppClaude, "/home/agent", []db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, HasToken: true},
	})
	assert.ErrorContains(t, err, "token of mcp server github was not delivered")
}

func TestNewAgentLaunch_UnsupportedApp(t *testing.T) {
	_, err := NewAgentLaunch(db.AppBash, "/home/agent", nil)
	assert.ErrorContains(t, err, "unsupported agent app")
}

func TestNewAgentLaunch_Claude(t *testing.T) {
	l, err := NewAgentLaunch(db.AppClaude, "/home/agent", agentInventory().McpServers)
	require.NoError(t, err)

	assert.Equal(t, "claude", l.Command)
	assert.Equal(t, []string{
		"-p",
		"--output-format=stream-json",
		"--verbose",
		"--include-partial-messages",
		"--mcp-config=/home/agent/mcp.json",
		"--strict-mcp-config",
		"--no-session-persistence",
		"--allowedTools=mcp__github__*,mcp__docs__*",
	}, l.Args)
	assert.Empty(t, l.TrailingArgs)
	assert.Equal(t, []string{"CLAUDE_CONFIG_DIR=/home/agent/claude"}, l.Env)
	assert.Equal(t, []string{"claude"}, l.Dirs)
	assert.Equal(t, []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN"}, l.APIKeyVars)
	assert.Empty(t, l.APIKeyAlias)

	require.Len(t, l.Files, 1)
	assert.Equal(t, "mcp.json", l.Files[0].Path)
	assert.JSONEq(t, `{"mcpServers":{
		"github":{"type":"http","url":"https://mcp.example.com/github","headers":{"Authorization":"Bearer ${SEMAPHORE_MCP_TOKEN_1}"}},
		"docs":{"type":"http","url":"https://mcp.example.com/docs"}}}`, string(l.Files[0].Content))
}

func TestNewAgentLaunch_ClaudeNoServers(t *testing.T) {
	l, err := NewAgentLaunch(db.AppClaude, "/home/agent", nil)
	require.NoError(t, err)

	assert.NotContains(t, l.Args, "--allowedTools")
	assert.JSONEq(t, `{"mcpServers":{}}`, string(l.Files[0].Content))
	assert.Equal(t, 0, l.Servers)
}

func TestNewAgentLaunch_Codex(t *testing.T) {
	l, err := NewAgentLaunch(db.AppCodex, "/home/agent", []db.McpServer{
		{Name: "github", URL: "https://mcp.example.com/github", Enabled: true, Token: agentTestToken},
		{Name: "docs", URL: "https://mcp.example.com/docs?q=\"x\"\\y", Enabled: true},
	})
	require.NoError(t, err)

	assert.Equal(t, "codex", l.Command)
	assert.Equal(t, []string{"exec", "--skip-git-repo-check"}, l.Args)
	assert.Equal(t, []string{"-"}, l.TrailingArgs)
	assert.Equal(t, []string{"CODEX_HOME=/home/agent/codex"}, l.Env)
	assert.Equal(t, []string{"codex"}, l.Dirs)
	assert.Equal(t, "CODEX_API_KEY", l.APIKeyAlias)
	assert.Equal(t, "OPENAI_API_KEY", l.APIKeyAliasOf)

	require.Len(t, l.Files, 1)
	assert.Equal(t, "codex/config.toml", l.Files[0].Path)
	assert.Equal(t, `[mcp_servers.github]
url = "https://mcp.example.com/github"
bearer_token_env_var = "SEMAPHORE_MCP_TOKEN_1"
default_tools_approval_mode = "auto"

[mcp_servers.docs]
url = "https://mcp.example.com/docs?q=\"x\"\\y"
default_tools_approval_mode = "auto"

`, string(l.Files[0].Content))
}

func TestTomlString_EscapesControlCharacters(t *testing.T) {
	assert.Equal(t, `"a\u000Ab\"c\\"`, tomlString("a\nb\"c\\"))
}

func TestAPIKeyAliasEnv(t *testing.T) {
	setupAgentConfig(t, nil)

	assert.Equal(t, []string{"CODEX_API_KEY=sk-1"}, apiKeyAliasEnv([]string{"OPENAI_API_KEY=sk-1"}, "CODEX_API_KEY", "OPENAI_API_KEY"))
	assert.Nil(t, apiKeyAliasEnv([]string{"OPENAI_API_KEY=sk-1", "CODEX_API_KEY=ck-2"}, "CODEX_API_KEY", "OPENAI_API_KEY"))
	assert.Nil(t, apiKeyAliasEnv(nil, "CODEX_API_KEY", "OPENAI_API_KEY"))
	assert.Nil(t, apiKeyAliasEnv([]string{"OPENAI_API_KEY=sk-1"}, "", ""))
}

func TestWriteAgentHome_PrivateModes(t *testing.T) {
	home := t.TempDir()
	l, err := NewAgentLaunch(db.AppCodex, home, agentInventory().McpServers)
	require.NoError(t, err)

	require.NoError(t, writeAgentHome(home, l))

	dir, err := os.Stat(filepath.Join(home, "codex"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dir.Mode().Perm())

	file, err := os.Stat(filepath.Join(home, "codex", "config.toml"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), file.Mode().Perm())
}
