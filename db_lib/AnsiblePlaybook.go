package db_lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type AnsiblePlaybook struct {
	TemplateID       int
	WorkingDirectory *string
	Repository       db.Repository
	Logger           task_logger.Logger
}

func (p AnsiblePlaybook) makeCmd(command string, args []string, environmentVars []string) (*exec.Cmd, error) {
	cmdDir, err := p.resolveWorkingDirectory()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(command, args...) //nolint:gosec
	cmd.Dir = cmdDir

	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=False")
	//cmd.Env = append(cmd.Env, "ANSIBLE_SSH_ARGS=-o UserKnownHostsFile=/dev/null")
	cmd.Env = append(cmd.Env, getEnvironmentVars()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", GetHomeDir(p.Repository, p.TemplateID)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if util.Config.HomeDirMode == util.HomeDirModeTemplateDir {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_HOME=%s", path.Join(p.Repository.GetHomePath(p.TemplateID), ".ansible")))
	}

	cmd.Env = append(cmd.Env, environmentVars...)

	cmd.SysProcAttr = appSysProcAttr()

	return cmd, nil
}

func (p AnsiblePlaybook) resolveWorkingDirectory() (string, error) {
	repoRoot := p.Repository.GetFullPath(p.TemplateID)
	if p.WorkingDirectory == nil {
		return repoRoot, nil
	}

	wd := filepath.Join(repoRoot, *p.WorkingDirectory)
	relPath, err := filepath.Rel(repoRoot, wd)
	if err != nil {
		return "", fmt.Errorf("resolve working directory relative to repository: %w", err)
	}
	if !filepath.IsLocal(relPath) {
		return "", fmt.Errorf("working directory %q is outside repository %q", wd, repoRoot)
	}

	return wd, nil
}

const ansibleConfigTimeout = 30 * time.Second

// CallbackPluginPaths asks Ansible where it would look for callback plugins
// with this environment, so ansible.cfg and its defaults are read by Ansible itself.
func (p AnsiblePlaybook) CallbackPluginPaths(environmentVars []string) ([]string, error) {
	// Verbose and debug output of the task would go to stdout and break the JSON
	quiet := append(slices.Clone(environmentVars), "ANSIBLE_VERBOSITY=0", "ANSIBLE_DEBUG=0")
	cmd, err := p.makeCmd("ansible-config", []string{"dump", "--format", "json"}, quiet)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.WaitDelay = 250 * time.Millisecond
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	stopCh := make(chan struct{})
	timer := time.AfterFunc(ansibleConfigTimeout, func() { close(stopCh) })
	defer timer.Stop()
	if err := waitCommand(cmd, stopCh, p.Logger); err != nil {
		return nil, err
	}

	var settings []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(out.Bytes(), &settings); err != nil {
		return nil, err
	}

	for _, setting := range settings {
		if setting.Name == "DEFAULT_CALLBACK_PLUGIN_PATH" {
			var paths []string
			err := json.Unmarshal(setting.Value, &paths)
			return paths, err
		}
	}

	return nil, errors.New("ansible-config does not report the callback plugin path")
}

func (p AnsiblePlaybook) runCmd(command string, args []string, environmentVars []string) error {
	cmd, err := p.makeCmd(command, args, environmentVars)
	if err != nil {
		return err
	}

	finishLog := p.Logger.LogCmd(cmd)
	defer finishLog()
	return cmd.Run()
}

func (p AnsiblePlaybook) RunPlaybook(args []string, environmentVars []string, inputs map[string]string, stopCh <-chan struct{}) error {
	cmd, err := p.makeCmd("ansible-playbook", args, environmentVars)
	if err != nil {
		return err
	}

	cmd.WaitDelay = 250 * time.Millisecond
	finishLog := p.Logger.LogCmd(cmd)
	defer finishLog()

	// This is the last pre-run cancellation check. The check and pty.Start are
	// not atomic; if cancellation wins between them, waitCommand stops the process.
	select {
	case <-stopCh:
		return nil
	default:
	}

	ptmx, err := pty.Start(cmd)

	if err != nil {
		return err
	}

	go func() {

		b := make([]byte, 100)

		var e error

		for {
			var n int
			n, e = ptmx.Read(b)
			if e != nil {
				break
			}

			s := strings.TrimSpace(string(b[0:n]))

			for k, v := range inputs {
				if strings.HasPrefix(s, k) {
					_, _ = ptmx.WriteString(v + "\n")
				}
			}
		}

	}()

	defer func() { _ = ptmx.Close() }()
	err = waitCommand(cmd, stopCh, p.Logger)
	finishLog()
	if errors.Is(err, exec.ErrWaitDelay) {
		p.Logger.Logf("Ansible command output draining exceeded %s and was stopped", cmd.WaitDelay)
		return nil
	}
	return err
}

func (p AnsiblePlaybook) RunGalaxy(args []string, environmentVars []string) error {
	return p.runCmd("ansible-galaxy", args, environmentVars)
}
