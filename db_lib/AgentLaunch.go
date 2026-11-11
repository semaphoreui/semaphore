package db_lib

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// mcpTokenEnvPrefix names the environment variables the MCP server tokens are
// passed in. The agent's config refers to a variable, never holds a token.
const mcpTokenEnvPrefix = "SEMAPHORE_MCP_TOKEN_"

// AgentFile is a file of an agent's config directory. It never holds a
// secret: tokens are referred to by the name of their variable.
type AgentFile struct {
	// Path is relative to the agent home, with forward slashes.
	Path    string
	Content []byte
}

// AgentLaunch is how to start an agent CLI, whatever runs it: AgentApp on
// the machine of the server or a runner, or a Pro container executor. The
// caller creates Dirs and writes Files under the home given to
// NewAgentLaunch, then runs
//
//	Command Args <template and task args> TrailingArgs
//
// with the prompt on stdin and Env, TokenEnv added last to the environment.
type AgentLaunch struct {
	Command      string
	Args         []string
	TrailingArgs []string

	// Env points the CLI at its config in home. Not secret.
	Env []string
	// TokenEnv holds the MCP server tokens as NAME=VALUE. Secret: pass it in
	// the process environment only, never in argv, a file or a log.
	TokenEnv []string
	// Secrets are the token values, to be masked in the task log.
	Secrets []string

	// Dirs are created (0700) before Files are written (0600). Relative to home.
	Dirs  []string
	Files []AgentFile

	// APIKeyVars lists the variables any of which authenticates the CLI.
	APIKeyVars []string
	// APIKeyAlias, when set, is a variable the CLI reads which takes the value
	// of APIKeyAliasOf if the user did not set it.
	APIKeyAlias   string
	APIKeyAliasOf string

	// Servers is the number of MCP servers given to the agent.
	Servers int

	// output makes the reader of the CLI's output; nil for a CLI whose output
	// is the task log as it is.
	output func(mask func(string) string) AgentOutput
}

// NewOutput returns the reader of this run's output: it makes task log lines,
// with Secrets masked, of the lines the CLI prints. Whoever runs the CLI
// passes every line of its stdout and stderr through one AgentOutput.
func (l *AgentLaunch) NewOutput() AgentOutput {
	mask := newSecretMasker(l.Secrets)
	if l.output == nil {
		return plainAgentOutput{mask: mask}
	}
	return l.output(mask)
}

// agentMcpServer is an MCP server as an agent CLI gets it. TokenEnv is the
// name of the variable holding its token, empty for a server without one.
type agentMcpServer struct {
	Name     string
	URL      string
	TokenEnv string
}

// agentCLI is what differs between the agent CLIs.
type agentCLI interface {
	// launch fills the CLI-specific part of l for a config directory in home:
	// Command, Args, TrailingArgs, Env, Dirs, Files, the API key fields and
	// the reader of the output.
	launch(l *AgentLaunch, home string, servers []agentMcpServer) error
}

func agentCLIFor(app db.TemplateApp) agentCLI {
	switch app {
	case db.AppClaude:
		return claudeCLI{}
	case db.AppCodex:
		return codexCLI{}
	default:
		return nil
	}
}

// NewAgentLaunch describes the run of the agent app with the given MCP
// servers and its config directory at home. It only computes, it writes
// nothing. Disabled servers are skipped. A server whose token was expected
// but did not arrive is an error: the agent must not silently run without it.
func NewAgentLaunch(app db.TemplateApp, home string, mcpServers []db.McpServer) (*AgentLaunch, error) {
	cli := agentCLIFor(app)
	if cli == nil {
		return nil, fmt.Errorf("unsupported agent app: %s", app)
	}

	l := &AgentLaunch{}

	var servers []agentMcpServer
	for i, s := range mcpServers {
		if !s.Enabled {
			continue
		}

		if s.HasToken && s.Token == "" {
			return nil, fmt.Errorf("the token of mcp server %s was not delivered to the task", s.Name)
		}

		server := agentMcpServer{Name: s.Name, URL: s.URL}
		if s.Token != "" {
			server.TokenEnv = mcpTokenEnvPrefix + strconv.Itoa(i+1)
			l.TokenEnv = append(l.TokenEnv, server.TokenEnv+"="+s.Token)
			l.Secrets = append(l.Secrets, s.Token)
		}

		servers = append(servers, server)
	}

	l.Servers = len(servers)

	if err := cli.launch(l, home, servers); err != nil {
		return nil, err
	}

	return l, nil
}

// claudeCLI runs Claude Code: claude -p, the prompt on stdin. In print mode
// the default text format prints nothing until the run is over, so the run
// asks for the stream of events, which claudeStreamOutput makes the log of.
type claudeCLI struct{}

type claudeMcpServer struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type claudeMcpConfig struct {
	McpServers map[string]claudeMcpServer `json:"mcpServers"`
}

func (claudeCLI) launch(l *AgentLaunch, home string, servers []agentMcpServer) error {
	config := claudeMcpConfig{McpServers: make(map[string]claudeMcpServer, len(servers))}
	allowed := make([]string, 0, len(servers))

	for _, s := range servers {
		server := claudeMcpServer{Type: "http", URL: s.URL}
		if s.TokenEnv != "" {
			// Claude Code expands ${VAR} in headers: the file holds the name only.
			server.Headers = map[string]string{"Authorization": "Bearer ${" + s.TokenEnv + "}"}
		}
		config.McpServers[s.Name] = server
		// The servers were chosen for this agent, so their tools need no
		// approval; in print mode an unapproved tool call is refused.
		allowed = append(allowed, "mcp__"+s.Name+"__*")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	l.Command = string(db.AppClaude)
	l.Dirs = []string{"claude"}
	l.Files = []AgentFile{{Path: "mcp.json", Content: data}}
	l.Args = []string{
		"-p",
		// stream-json needs --verbose in print mode.
		"--output-format=stream-json", "--verbose", "--include-partial-messages",
		"--mcp-config=" + path.Join(home, "mcp.json"), "--strict-mcp-config", "--no-session-persistence",
	}
	if len(allowed) > 0 {
		l.Args = append(l.Args, "--allowedTools="+strings.Join(allowed, ","))
	}
	l.Env = []string{"CLAUDE_CONFIG_DIR=" + path.Join(home, "claude")}
	l.APIKeyVars = []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN"}
	l.output = func(mask func(string) string) AgentOutput {
		return &claudeStreamOutput{mask: mask}
	}
	return nil
}

// codexCLI runs Codex: codex exec, the prompt on stdin ("-"). It reports its
// progress in plain text as it goes, so its output is the log.
type codexCLI struct{}

func (codexCLI) launch(l *AgentLaunch, home string, servers []agentMcpServer) error {
	l.Command = string(db.AppCodex)
	l.Dirs = []string{"codex"}
	l.Files = []AgentFile{{Path: "codex/config.toml", Content: []byte(codexConfig(servers))}}
	l.Args = []string{"exec", "--skip-git-repo-check"}
	l.TrailingArgs = []string{"-"}
	l.Env = []string{"CODEX_HOME=" + path.Join(home, "codex")}
	l.APIKeyVars = []string{"CODEX_API_KEY", "OPENAI_API_KEY"}
	// codex exec reads CODEX_API_KEY; users know the key as OPENAI_API_KEY.
	l.APIKeyAlias = "CODEX_API_KEY"
	l.APIKeyAliasOf = "OPENAI_API_KEY"
	return nil
}

// codexConfig renders config.toml. Names are limited to [A-Za-z0-9_-] by
// McpServer.Validate, so they are valid bare keys; values are quoted.
func codexConfig(servers []agentMcpServer) string {
	var b strings.Builder

	for _, s := range servers {
		fmt.Fprintf(&b, "[mcp_servers.%s]\n", s.Name)
		fmt.Fprintf(&b, "url = %s\n", tomlString(s.URL))
		if s.TokenEnv != "" {
			fmt.Fprintf(&b, "bearer_token_env_var = %s\n", tomlString(s.TokenEnv))
		}
		// The servers were chosen for this agent; exec can not ask anyone.
		b.WriteString("default_tools_approval_mode = \"auto\"\n\n")
	}

	return b.String()
}

// tomlString quotes s as a TOML basic string.
func tomlString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\u%04X`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
