package db_lib

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/creack/pty"
	"os"
	"os/exec"
)

type AnsiblePlaybook struct {
	TemplateID int
	Repository db.Repository
	Logger     lib.Logger
}

func (p AnsiblePlaybook) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = p.GetFullPath()

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	cmd.Env = append(cmd.Env, "ANSIBLE_FORCE_COLOR=True")
	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	sensitiveEnvs := []string{
		"SEMAPHORE_ACCESS_KEY_ENCRYPTION",
		"SEMAPHORE_ADMIN_PASSWORD",
		"SEMAPHORE_DB_USER",
		"SEMAPHORE_DB_NAME",
		"SEMAPHORE_DB_HOST",
		"SEMAPHORE_DB_PASS",
		"SEMAPHORE_LDAP_PASSWORD",
	}

	// Remove sensitive env variables from cmd process
	for _, env := range sensitiveEnvs {
		cmd.Env = append(cmd.Env, env+"=")
	}

	return cmd
}

func (p AnsiblePlaybook) runCmd(command string, args []string) error {
	cmd := p.makeCmd(command, args, nil)
	p.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (p AnsiblePlaybook) RunPlaybook(args []string, environmentVars *[]string, inputs []string, cb func(*os.Process)) error {
	cmd := p.makeCmd("ansible-playbook", args, environmentVars)
	p.Logger.LogCmd(cmd)

	ptmx, err := pty.Start(cmd)

	if err != nil {
		panic(err)
	}

	go func() {
		b := make([]byte, 100)
		var e error
		for e == nil {
			var n int
			n, e = ptmx.Read(b)
			s := string(b[0:n])
			fmt.Println(s)
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
