package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/creack/pty"
)

type AnsiblePlaybook struct {
	TemplateID int
	Repository db.Repository
	Logger     task_logger.Logger
}

func (p AnsiblePlaybook) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = p.GetFullPath()

	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")
	cmd.Env = append(cmd.Env, getEnvironmentVars()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	return cmd
}

func (p AnsiblePlaybook) runCmd(command string, args []string) error {
	cmd := p.makeCmd(command, args, nil)
	p.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (p AnsiblePlaybook) RunPlaybook(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	cmd := p.makeCmd("ansible-playbook", args, environmentVars)
	p.Logger.LogCmd(cmd)

	ptmx, err := pty.Start(cmd)

	if err != nil {
		panic(err)
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
	return cmd.Wait()
}

func (p AnsiblePlaybook) RunGalaxy(args []string) error {
	return p.runCmd("ansible-galaxy", args)
}

func (p AnsiblePlaybook) GetFullPath() (path string) {
	path = p.Repository.GetFullPath(p.TemplateID)
	return
}
