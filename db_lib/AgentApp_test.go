package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/agentlog"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const agentTestToken = "tok-3f9a1c7e5b"

// lineLogger keeps every logged line.
type lineLogger struct {
	task_logger.NopLogger
	mu    sync.Mutex
	lines []string
}

func (l *lineLogger) Log(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, msg)
}

func (l *lineLogger) Logf(format string, a ...any) {
	l.Log(strings.TrimSpace(fmt.Sprintf(format, a...)))
}

func (l *lineLogger) text() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

func agentInventory() db.Inventory {
	return db.Inventory{ID: 7, Type: db.InventoryMcp, McpServers: []db.McpServer{
		{ID: 1, Name: "github", URL: "https://mcp.example.com/github", Enabled: true, Token: agentTestToken},
		{ID: 2, Name: "docs", URL: "https://mcp.example.com/docs", Enabled: true},
		{ID: 3, Name: "off", URL: "https://mcp.example.com/off", Enabled: false, Token: "disabled-token"},
	}}
}

// setupAgentConfig points util.Config at a temporary tmp dir; not parallel.
func setupAgentConfig(t *testing.T, apps map[string]util.App) {
	t.Helper()
	previous := util.Config
	util.Config = &util.ConfigType{TmpPath: t.TempDir(), Apps: apps}
	t.Cleanup(func() { util.Config = previous })
	require.NoError(t, os.MkdirAll(util.Config.GetProjectTmpDir(1), 0o755))
}

// fakeAgent writes a script that reports what the agent would see: its
// arguments, the token variable, the prompt on stdin, and prints the token
// itself to check that the log masks it.
func fakeAgent(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not available")
	}

	script := filepath.Join(t.TempDir(), "agent")
	require.NoError(t, os.WriteFile(script, []byte(`#!/bin/sh
echo "args: $*"
echo "token var: $SEMAPHORE_MCP_TOKEN_1"
echo "codex home: $CODEX_HOME"
echo "claude dir: $CLAUDE_CONFIG_DIR"
echo "codex key: $CODEX_API_KEY"
echo "prompt: $(cat)"
echo "pwd: $(pwd)"
`), 0o755))
	return script
}

func TestAgentApp_RunPassesPromptOnStdinAndMasksTokens(t *testing.T) {
	for _, app := range []db.TemplateApp{db.AppClaude, db.AppCodex} {
		t.Run(string(app), func(t *testing.T) {
			setupAgentConfig(t, map[string]util.App{string(app): {Active: true, AppPath: fakeAgent(t)}})

			repoDir := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(repoDir, "svc"), 0o755))

			logger := &lineLogger{}
			agent := newAgentApp(db.Template{
				ID: 1, ProjectID: 1, App: app, Playbook: "svc",
				TaskParams: db.MapStringAnyField{"prompt": "fix the build"},
			}, db.Repository{ProjectID: 1, GitURL: repoDir}, agentInventory(), logger)

			err := agent.Run(LocalAppRunningArgs{
				CliArgs:         map[string][]string{"default": {"--model", "m1"}},
				EnvironmentVars: []string{"OPENAI_API_KEY=sk-test", "ANTHROPIC_API_KEY=ak-test", "CODEX_HOME=/elsewhere"},
				TaskParams:      &db.AgentTaskParams{},
			})
			require.NoError(t, err)

			out := logger.text()
			assert.Contains(t, out, "prompt: fix the build")
			assert.Contains(t, out, "token var: "+maskedSecret)
			assert.NotContains(t, out, agentTestToken, "the token must not reach the log")
			assert.NotContains(t, out, "disabled-token")
			assert.Contains(t, out, "pwd: "+filepath.Join(repoDir, "svc"))
			assert.Contains(t, out, "Running "+string(app)+" with 2 MCP server(s)")

			argsLine := lineWithPrefix(t, out, "args: ")
			assert.NotContains(t, argsLine, agentTestToken)
			assert.NotContains(t, argsLine, "fix the build", "the prompt goes to stdin, not argv")
			assert.Contains(t, argsLine, "--model m1")

			if app == db.AppCodex {
				assert.True(t, strings.HasSuffix(argsLine, "--model m1 -"), argsLine)
				assert.NotContains(t, out, "codex home: /elsewhere", "the environment can not redirect the config")
				assert.Contains(t, out, "codex key: sk-test")
			} else {
				assert.True(t, strings.HasPrefix(argsLine,
					"args: -p --output-format=stream-json --verbose --include-partial-messages --mcp-config="), argsLine)
			}

			// The config directory lives until Clear and is gone after it.
			entries, err := os.ReadDir(util.Config.GetProjectTmpDir(1))
			require.NoError(t, err)
			require.Len(t, entries, 1)
			agent.Clear()
			entries, err = os.ReadDir(util.Config.GetProjectTmpDir(1))
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

// gateLogger keeps the lines and creates the gate file when a line with the
// marker is logged.
type gateLogger struct {
	lineLogger
	marker string
	gate   string
}

func (l *gateLogger) Log(msg string) {
	l.lineLogger.Log(msg)
	if strings.Contains(msg, l.marker) {
		_ = os.WriteFile(l.gate, nil, 0o600)
	}
}

// The agent goes on only after its tool call has reached the log: the log is
// written while the agent runs, not when it exits.
func TestAgentApp_RunLogsClaudeEventsWhileRunning(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not available")
	}

	gate := filepath.Join(t.TempDir(), "gate")
	script := filepath.Join(t.TempDir(), "agent")
	require.NoError(t, os.WriteFile(script, []byte(`#!/bin/sh
cat > /dev/null
echo '{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"ls"}}]}}'
i=0
while [ ! -f "`+gate+`" ]; do
  i=$((i+1))
  if [ "$i" -gt 200 ]; then
    echo "the tool call did not reach the log while the agent was running"
    exit 3
  fi
  sleep 0.05
done
echo '{"type":"result","subtype":"success","is_error":false,"result":"done","num_turns":1,"duration_ms":1200}'
`), 0o755))

	setupAgentConfig(t, map[string]util.App{"claude": {Active: true, AppPath: script}})

	logger := &gateLogger{marker: `"name":"Bash"`, gate: gate}
	agent := newAgentApp(db.Template{
		ID: 1, ProjectID: 1, App: db.AppClaude,
		TaskParams: db.MapStringAnyField{"prompt": "hi"},
	}, db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{}, logger)
	t.Cleanup(agent.Clear)

	require.NoError(t, agent.Run(LocalAppRunningArgs{
		EnvironmentVars: []string{"ANTHROPIC_API_KEY=ak-test"},
		TaskParams:      &db.AgentTaskParams{},
	}))

	// The log holds the agent's events in the format of agentlog, not the
	// CLI's own; as plain text they read as a transcript.
	out := logger.text()
	assert.Contains(t, out, `{"@agent":"tool","id":"toolu_1","name":"Bash","input":{"command":"ls"}}`)
	assert.NotContains(t, out, `"type"`, "the CLI's events are not dumped")

	text := []string{}
	for _, line := range strings.Split(out, "\n") {
		text = append(text, agentlog.Lines(line)...)
	}
	assert.Contains(t, text, `> Bash {"command":"ls"}`)
	assert.Contains(t, text, "done")
	assert.Contains(t, text, "Agent finished: 1 turn(s), 1.2s")
}

func TestAgentApp_RunWarnsWithoutAPIKey(t *testing.T) {
	setupAgentConfig(t, map[string]util.App{"claude": {Active: true, AppPath: fakeAgent(t)}})

	logger := &lineLogger{}
	agent := newAgentApp(db.Template{
		ID: 1, ProjectID: 1, App: db.AppClaude,
		TaskParams: db.MapStringAnyField{"prompt": "hi"},
	}, db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{}, logger)
	t.Cleanup(agent.Clear)

	require.NoError(t, agent.Run(LocalAppRunningArgs{TaskParams: &db.AgentTaskParams{}}))
	assert.Contains(t, logger.text(), "No API key for claude: add ANTHROPIC_API_KEY")
}

func TestAgentApp_RunRejectsForbiddenPromptAndWrongInventory(t *testing.T) {
	setupAgentConfig(t, map[string]util.App{"claude": {Active: true, AppPath: fakeAgent(t)}})

	t.Run("task prompt without the flag", func(t *testing.T) {
		agent := newAgentApp(db.Template{ID: 1, ProjectID: 1, App: db.AppClaude,
			TaskParams: db.MapStringAnyField{"prompt": "hi"}},
			db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{}, &lineLogger{})

		err := agent.Run(LocalAppRunningArgs{TaskParams: &db.AgentTaskParams{Prompt: "other"}})
		assert.ErrorContains(t, err, "does not allow overriding the prompt")
	})

	t.Run("static inventory", func(t *testing.T) {
		agent := newAgentApp(db.Template{ID: 1, ProjectID: 1, App: db.AppClaude,
			TaskParams: db.MapStringAnyField{"prompt": "hi"}},
			db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{ID: 3, Type: db.InventoryStatic}, &lineLogger{})

		err := agent.Run(LocalAppRunningArgs{TaskParams: &db.AgentTaskParams{}})
		assert.ErrorContains(t, err, "needs an mcp inventory")
	})
}

func TestCreateApp_Agents(t *testing.T) {
	for _, app := range []db.TemplateApp{db.AppClaude, db.AppCodex} {
		created := CreateApp(db.Template{App: app}, db.Repository{}, db.Inventory{}, task_logger.NopLogger{})
		agent, ok := created.(*AgentApp)
		require.True(t, ok, "%s must run as AgentApp", app)
		assert.Equal(t, app, agent.Name)
	}
}

func lineWithPrefix(t *testing.T, text string, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	require.Failf(t, "line not found", "no line starting with %q in:\n%s", prefix, text)
	return ""
}
