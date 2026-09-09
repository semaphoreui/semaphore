package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

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

	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = cmdDir

	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=False")
	//cmd.Env = append(cmd.Env, "ANSIBLE_SSH_ARGS=-o UserKnownHostsFile=/dev/null")
	cmd.Env = append(cmd.Env, getEnvironmentVars()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", getHomeDir(p.Repository, p.TemplateID)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if util.Config.HomeDirMode == util.HomeDirModeTemplateDir {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_HOME=%s", path.Join(p.Repository.GetHomePath(p.TemplateID), ".ansible")))
	}

	cmd.Env = append(cmd.Env, environmentVars...)

	cmd.SysProcAttr = util.Config.GetAppSysProcAttr()

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

func (p AnsiblePlaybook) runCmd(command string, args []string, environmentVars []string) error {
	cmd, err := p.makeCmd(command, args, environmentVars)
	if err != nil {
		return err
	}

	finishLog := p.Logger.LogCmd(cmd)
	defer finishLog()
	return cmd.Run()
}

func (p AnsiblePlaybook) RunPlaybook(args []string, environmentVars []string, inputs map[string]string, cb func(*os.Process)) error {
	cmd, err := p.makeCmd("ansible-playbook", args, environmentVars)
	if err != nil {
		return err
	}

	finishLog := p.Logger.LogCmd(cmd)
	defer finishLog()

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
	cb(cmd.Process)
	err = cmd.Wait()
	return err
}

func (p AnsiblePlaybook) RunGalaxy(args []string, environmentVars []string) error {
	return p.runCmd("ansible-galaxy", args, environmentVars)
}
