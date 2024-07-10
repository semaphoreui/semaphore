package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
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

	cmd.Env = removeSensitiveEnvs(os.Environ())

	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")

	// TODO: Following option doesn't work when password authentication used.
	// 		 So, we need to check args for --ask-pass, --ask-become-pass or remove this code completely.
	//       What reason to use this code: prevent hanging of semaphore when host key confirmation required.
	//cmd.Env = append(cmd.Env, "ANSIBLE_SSH_ARGS=\"-o BatchMode=yes\"")

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
