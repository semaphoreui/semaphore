package lib

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"os/exec"
	"strings"
)

type AnsiblePlaybook struct {
	TemplateID int
	Repository db.Repository
	Logger     Logger
}

func (p AnsiblePlaybook) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = p.GetFullPath()

	if util.Config.AnsibleUsername != "" {
		cmdUser(cmd, util.Config.AnsibleUsername)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")
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

func (p AnsiblePlaybook) RunPlaybook(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	cmd := p.makeCmd("ansible-playbook", args, environmentVars)
	p.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
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
