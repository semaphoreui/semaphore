package db_lib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// maskedSecret replaces a token wherever the agent prints it.
const maskedSecret = "********"

// AgentApp runs an AI agent CLI (Codex or Claude Code) in the template's
// repository directory with the MCP servers of an mcp inventory.
type AgentApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	Inventory  db.Inventory
	Name       db.TemplateApp

	mu sync.Mutex
	// home is the per-task directory with the agent's config, removed by Clear.
	home string
}

func newAgentApp(template db.Template, repository db.Repository, inventory db.Inventory, logger task_logger.Logger) *AgentApp {
	return &AgentApp{
		Logger:     logger,
		Template:   template,
		Repository: repository,
		Inventory:  inventory,
		Name:       template.App,
	}
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

func (t *AgentApp) Run(args LocalAppRunningArgs) error {
	if agentCLIFor(t.Name) == nil {
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

	home, err := os.MkdirTemp(util.Config.GetProjectTmpDir(t.Template.ProjectID), "agent_")
	if err != nil {
		return fmt.Errorf("can't create the agent config directory: %w", err)
	}
	t.mu.Lock()
	t.home = home
	t.mu.Unlock()

	launch, err := NewAgentLaunch(t.Name, home, t.Inventory.McpServers)
	if err != nil {
		return err
	}

	if err = writeAgentHome(home, launch); err != nil {
		return err
	}

	cliArgs := append([]string{}, launch.Args...)
	cliArgs = append(cliArgs, args.CliArgs["default"]...)
	cliArgs = append(cliArgs, launch.TrailingArgs...)

	// The agent's own variables go last: a variable of the environment can not
	// redirect its config directory or replace a token.
	env := append([]string{}, args.EnvironmentVars...)
	env = append(env, launch.Env...)
	env = append(env, launch.TokenEnv...)
	env = append(env, apiKeyAliasEnv(env, launch.APIKeyAlias, launch.APIKeyAliasOf)...)

	t.warnOnMissingAPIKey(env, launch.APIKeyVars)

	cmd := t.makeCmd(launch.Command, cliArgs, env)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.WaitDelay = 250 * time.Millisecond

	t.Logger.Logf("Running %s with %d MCP server(s)", t.Name, launch.Servers)

	finishLog := logCmdMasked(t.Logger, cmd, launch.Secrets)
	defer finishLog()

	err = runCommand(cmd, args.StopCh, t.Logger)
	finishLog()
	if errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}

// writeAgentHome creates the launch's directories and files under home.
func writeAgentHome(home string, launch *AgentLaunch) error {
	for _, dir := range launch.Dirs {
		if err := os.MkdirAll(filepath.Join(home, filepath.FromSlash(dir)), 0o700); err != nil {
			return err
		}
	}
	for _, file := range launch.Files {
		if err := os.WriteFile(filepath.Join(home, filepath.FromSlash(file.Path)), file.Content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (t *AgentApp) makeCmd(command string, args []string, environmentVars []string) *exec.Cmd {
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
func (t *AgentApp) warnOnMissingAPIKey(env []string, apiKeyVars []string) {
	for _, name := range apiKeyVars {
		if envValue(env, name) != "" {
			return
		}
	}

	t.Logger.Logf("No API key for %s: add %s as a secret of the template's Variable Group",
		t.Name, strings.Join(apiKeyVars, " or "))
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

// apiKeyAliasEnv sets alias to the value of of, unless alias is set already.
func apiKeyAliasEnv(env []string, alias string, of string) []string {
	if alias == "" || envValue(env, alias) != "" {
		return nil
	}
	if key := envValue(env, of); key != "" {
		return []string{alias + "=" + key}
	}
	return nil
}
