package db_lib

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// mcpTokenEnvPrefix names the environment variables the MCP server tokens are
// passed in. The agent's config refers to a variable, never holds a token.
const mcpTokenEnvPrefix = "SEMAPHORE_MCP_TOKEN_"

// maskedSecret replaces a token wherever the agent prints it.
const maskedSecret = "********"

// agentMcpServer is an MCP server as an agent CLI gets it. TokenEnv is the
// name of the variable holding its token, empty for a server without one.
type agentMcpServer struct {
	Name     string
	URL      string
	TokenEnv string
}

// agentCLI is what differs between the agent CLIs; everything else is AgentApp.
type agentCLI interface {
	// command is the default binary, replaced by util.Config.Apps[...].AppPath.
	command() string
	// prepare writes the CLI's config into home, a directory of this task only,
	// and returns the arguments of a non-interactive run reading the prompt
	// from stdin, and the environment that points the CLI at home.
	prepare(home string, servers []agentMcpServer) (args []string, env []string, err error)
	// trailingArgs go after the template and task arguments.
	trailingArgs() []string
	// apiKeyVars lists the variables any of which authenticates the CLI.
	apiKeyVars() []string
}

// AgentApp runs an AI agent CLI (Codex or Claude Code) in the template's
// repository directory with the MCP servers of an mcp inventory.
type AgentApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	Inventory  db.Inventory
	Name       db.TemplateApp

	cli agentCLI

	mu sync.Mutex
	// home is the per-task directory with the agent's config, removed by Clear.
	home string
}

func newAgentApp(template db.Template, repository db.Repository, inventory db.Inventory, logger task_logger.Logger) *AgentApp {
	app := &AgentApp{
		Logger:     logger,
		Template:   template,
		Repository: repository,
		Inventory:  inventory,
		Name:       template.App,
	}

	switch template.App {
	case db.AppClaude:
		app.cli = claudeCLI{}
	case db.AppCodex:
		app.cli = codexCLI{}
	}

	return app
}

func (t *AgentApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger
	return logger
}

func (t *AgentApp) InstallRequirements(LocalAppInstallingArgs) error {
	return nil
}

// GetFullPath is the agent's working directory: the repository, or the
// template's subdirectory of it.
func (t *AgentApp) GetFullPath() string {
	return path.Join(t.Repository.GetFullPath(t.Template.ID), strings.TrimPrefix(t.Template.Playbook, "/"))
}

// Clear removes the agent's config directory.
func (t *AgentApp) Clear() {
	t.mu.Lock()
	home := t.home
	t.home = ""
	t.mu.Unlock()

	if home == "" {
		return
	}

	if err := os.RemoveAll(home); err != nil {
		t.Logger.Logf("Can't remove the agent config directory: %s", err.Error())
	}
}

// mcpServers lists the enabled servers in a stable order and assigns each
// token its variable. A disabled server is not given to the agent even if a
// caller left it in the inventory.
func (t *AgentApp) mcpServers() (servers []agentMcpServer, tokenEnv []string, secrets []string, err error) {
	for i, s := range t.Inventory.McpServers {
		if !s.Enabled {
			continue
		}

		// A server that has a token must not silently run without it, e.g.
		// where the token was not delivered to the machine running the task.
		if s.HasToken && s.Token == "" {
			err = fmt.Errorf("the token of mcp server %s was not delivered to the task", s.Name)
			return
		}

		server := agentMcpServer{Name: s.Name, URL: s.URL}
		if s.Token != "" {
			server.TokenEnv = mcpTokenEnvPrefix + strconv.Itoa(i+1)
			tokenEnv = append(tokenEnv, server.TokenEnv+"="+s.Token)
			secrets = append(secrets, s.Token)
		}

		servers = append(servers, server)
	}

	return
}

func (t *AgentApp) Run(args LocalAppRunningArgs) error {
	if t.cli == nil {
		return fmt.Errorf("unsupported agent app: %s", t.Name)
	}

	taskParams, ok := args.TaskParams.(*db.AgentTaskParams)
	if !ok || taskParams == nil {
		taskParams = &db.AgentTaskParams{}
	}

	prompt, err := t.Template.AgentPrompt(*taskParams)
	if err != nil {
		return err
	}

	if t.Inventory.ID != 0 && t.Inventory.Type != db.InventoryMcp {
		return fmt.Errorf("an agent template needs an mcp inventory, got %s", t.Inventory.Type)
	}

	servers, tokenEnv, secrets, err := t.mcpServers()
	if err != nil {
		return err
	}

	home, err := os.MkdirTemp(util.Config.GetProjectTmpDir(t.Template.ProjectID), "agent_")
	if err != nil {
		return fmt.Errorf("can't create the agent config directory: %w", err)
	}
	t.mu.Lock()
	t.home = home
	t.mu.Unlock()

	cliArgs, cliEnv, err := t.cli.prepare(home, servers)
	if err != nil {
		return err
	}

	cliArgs = append(cliArgs, args.CliArgs["default"]...)
	cliArgs = append(cliArgs, t.cli.trailingArgs()...)

	// The agent's own variables go last: a variable of the environment can not
	// redirect its config directory or replace a token.
	env := append([]string{}, args.EnvironmentVars...)
	env = append(env, cliEnv...)
	env = append(env, tokenEnv...)

	if _, isCodex := t.cli.(codexCLI); isCodex {
		env = append(env, codexAPIKeyEnv(env)...)
	}

	t.warnOnMissingAPIKey(env)

	cmd := t.makeCmd(cliArgs, env)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.WaitDelay = 250 * time.Millisecond

	t.Logger.Logf("Running %s with %d MCP server(s)", t.Name, len(servers))

	finishLog := logCmdMasked(t.Logger, cmd, secrets)
	defer finishLog()

	err = runCommand(cmd, args.StopCh, t.Logger)
	finishLog()
	if errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}

func (t *AgentApp) makeCmd(args []string, environmentVars []string) *exec.Cmd {
	command := t.cli.command()

	if app, ok := util.Config.Apps[string(t.Name)]; ok {
		if app.AppPath != "" {
			command = app.AppPath
		}
		if app.AppArgs != nil {
			args = append(append([]string{}, app.AppArgs...), args...)
		}
	}

	cmd := exec.Command(command, args...) //nolint:gosec
	cmd.Dir = t.GetFullPath()

	cmd.Env = getEnvironmentVars()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", getHomeDir(t.Repository, t.Template.ID)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))
	cmd.Env = append(cmd.Env, environmentVars...)

	cmd.SysProcAttr = util.Config.GetAppSysProcAttr()

	return cmd
}

// warnOnMissingAPIKey tells in the task log which variable to add when the
// agent has no API key; the run goes on, the CLI reports its own error.
func (t *AgentApp) warnOnMissingAPIKey(env []string) {
	for _, name := range t.cli.apiKeyVars() {
		if envValue(env, name) != "" {
			return
		}
	}

	t.Logger.Logf("No API key for %s: add %s as a secret of the template's Variable Group",
		t.Name, strings.Join(t.cli.apiKeyVars(), " or "))
}

// logCmdMasked logs the command's output line by line with every secret
// replaced, so a token the agent prints does not reach the task log.
func logCmdMasked(logger task_logger.Logger, cmd *exec.Cmd, secrets []string) func() {
	pairs := make([]string, 0, len(secrets)*2)
	for _, s := range secrets {
		if s != "" {
			pairs = append(pairs, s, maskedSecret)
		}
	}
	replacer := strings.NewReplacer(pairs...)

	reader, writer := io.Pipe()
	cmd.Stdout = writer
	cmd.Stderr = writer

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer reader.Close() //nolint:errcheck

		scanner := bufio.NewScanner(reader)
		const maxCapacity = 10 * 1024 * 1024
		scanner.Buffer(make([]byte, 64*1024), maxCapacity)
		for scanner.Scan() {
			logger.Log(replacer.Replace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			logger.Logf("Failed to read the agent output: %s", err.Error())
			_, _ = io.Copy(io.Discard, reader)
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			_ = writer.Close()
			<-done
		})
	}
}

// claudeCLI runs Claude Code: claude -p, the prompt on stdin.
type claudeCLI struct{}

func (claudeCLI) command() string { return string(db.AppClaude) }

func (claudeCLI) apiKeyVars() []string { return []string{"ANTHROPIC_API_KEY"} }

func (claudeCLI) trailingArgs() []string { return nil }

type claudeMcpServer struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type claudeMcpConfig struct {
	McpServers map[string]claudeMcpServer `json:"mcpServers"`
}

func (claudeCLI) prepare(home string, servers []agentMcpServer) (args []string, env []string, err error) {
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

	configDir := filepath.Join(home, "claude")
	if err = os.Mkdir(configDir, 0o700); err != nil {
		return
	}

	configFile := filepath.Join(home, "mcp.json")
	if err = writeAgentConfig(configFile, config); err != nil {
		return
	}

	args = []string{"-p", "--mcp-config=" + configFile, "--strict-mcp-config", "--no-session-persistence"}
	if len(allowed) > 0 {
		args = append(args, "--allowedTools="+strings.Join(allowed, ","))
	}

	env = []string{"CLAUDE_CONFIG_DIR=" + configDir}
	return
}

// codexCLI runs Codex: codex exec, the prompt on stdin ("-").
type codexCLI struct{}

func (codexCLI) command() string { return string(db.AppCodex) }

func (codexCLI) apiKeyVars() []string { return []string{"CODEX_API_KEY", "OPENAI_API_KEY"} }

func (codexCLI) trailingArgs() []string { return []string{"-"} }

func (codexCLI) prepare(home string, servers []agentMcpServer) (args []string, env []string, err error) {
	codexHome := filepath.Join(home, "codex")
	if err = os.Mkdir(codexHome, 0o700); err != nil {
		return
	}

	if err = os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(codexConfig(servers)), 0o600); err != nil {
		return
	}

	args = []string{"exec", "--skip-git-repo-check"}
	env = []string{"CODEX_HOME=" + codexHome}
	return
}

// envValue returns the last value of name in env, or else the value a child
// process gets from the base environment.
func envValue(env []string, name string) string {
	value := BaseEnvironmentVar(name)
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, name+"="); ok {
			value = v
		}
	}
	return value
}

// codexAPIKeyEnv passes OPENAI_API_KEY as CODEX_API_KEY, the variable codex
// exec reads, unless CODEX_API_KEY is set already.
func codexAPIKeyEnv(env []string) []string {
	if envValue(env, "CODEX_API_KEY") != "" {
		return nil
	}
	if key := envValue(env, "OPENAI_API_KEY"); key != "" {
		return []string{"CODEX_API_KEY=" + key}
	}
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

func writeAgentConfig(file string, config any) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o600)
}
