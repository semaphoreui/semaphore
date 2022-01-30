package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type AnsiblePlaybook struct {
	TemplateID  int
	Logger      Logger
	Repository  db.Repository
	Environment db.Environment
}

func (p AnsiblePlaybook) makeCmd(command string, args []string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = p.GetFullPath()
	p.setCmdEnvironment(cmd, p.Repository.SSHKey.GetSshCommand())
	return cmd
}

func (p AnsiblePlaybook) runCmd(command string, args []string) error {
	cmd := p.makeCmd(command, args)
	p.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (p AnsiblePlaybook) GetHosts(args []string) (hosts []string, err error) {
	args = append(args, "--list-hosts")
	cmd := p.makeCmd("ansible-playbook", args)

	var errb bytes.Buffer
	cmd.Stderr = &errb

	out, err := cmd.Output()
	if err != nil {
		return
	}

	re := regexp.MustCompile(`(?m)^\\s{6}(.*)$`)
	matches := re.FindAllSubmatch(out, 20)
	hosts = make([]string, len(matches))
	for i := range matches {
		hosts[i] = string(matches[i][1])
	}

	return
}

func (p AnsiblePlaybook) MakeRunCmd(args []string) (cmd *exec.Cmd, err error) {
	cmd = p.makeCmd("ansible-playbook", args)
	p.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err = cmd.Start()
	if err != nil {
		return
	}
	return
}

func (p AnsiblePlaybook) RunGalaxy(args []string) error {
	return p.runCmd("ansible-galaxy", args)
}

func (p AnsiblePlaybook) GetFullPath() (path string) {
	path = p.Repository.GetFullPath(p.TemplateID)
	return
}

func (p AnsiblePlaybook) setCmdEnvironment(cmd *exec.Cmd, gitSSHCommand string) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	env = append(env, fmt.Sprintf("PWD=%s", cmd.Dir))
	env = append(env, fmt.Sprintln("PYTHONUNBUFFERED=1"))
	env = append(env, fmt.Sprintln("GIT_TERMINAL_PROMPT=0"))
	env = append(env, extractCommandEnvironment(p.Environment.JSON)...)

	if gitSSHCommand != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", gitSSHCommand))
	}
	cmd.Env = env
}

func extractCommandEnvironment(envJSON string) []string {
	env := make([]string, 0)
	var js map[string]interface{}
	err := json.Unmarshal([]byte(envJSON), &js)
	if err == nil {
		if cfg, ok := js["ENV"]; ok {
			switch v := cfg.(type) {
			case map[string]interface{}:
				for key, val := range v {
					env = append(env, fmt.Sprintf("%s=%s", key, val))
				}
			}
		}
	}
	return env
}
