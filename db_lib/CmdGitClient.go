package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

type CmdGitClient struct {
	keyInstallation db.AccessKeyInstallation
}

func (c CmdGitClient) makeCmd(r GitRepository, targetDir GitRepositoryDirType, args ...string) *exec.Cmd {
	cmd := exec.Command("git") //nolint: gas

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintln("GIT_TERMINAL_PROMPT=0"))
	if r.Repository.SSHKey.Type == db.AccessKeySSH {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_AUTH_SOCK=%s", c.keyInstallation.SSHAgent.SocketFile))
		sshCmd := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
		if util.Config.SshConfigPath != "" {
			sshCmd += " -F " + util.Config.SshConfigPath
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
	}

	switch targetDir {
	case GitRepositoryTmpPath:
		cmd.Dir = util.Config.TmpPath
	case GitRepositoryFullPath:
		cmd.Dir = r.GetFullPath()
	default:
		panic("unknown Repository directory type")
	}

	cmd.Args = append(cmd.Args, args...)

	return cmd
}

func (c CmdGitClient) run(r GitRepository, targetDir GitRepositoryDirType, args ...string) error {
	var err error
	c.keyInstallation, err = r.Repository.SSHKey.Install(db.AccessKeyRoleGit, r.Logger)

	if err != nil {
		return err
	}

	defer c.keyInstallation.Destroy() //nolint: errcheck

	cmd := c.makeCmd(r, targetDir, args...)

	r.Logger.LogCmd(cmd)

	return cmd.Run()
}

func (c CmdGitClient) output(r GitRepository, targetDir GitRepositoryDirType, args ...string) (out string, err error) {
	c.keyInstallation, err = r.Repository.SSHKey.Install(db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return
	}

	defer c.keyInstallation.Destroy() //nolint: errcheck

	bytes, err := c.makeCmd(r, targetDir, args...).Output()
	if err != nil {
		return
	}
	out = strings.Trim(string(bytes), " \n")
	return
}

func (c CmdGitClient) Clone(r GitRepository) error {
	r.Logger.Log("Cloning Repository " + r.Repository.GitURL)

	var dirName string
	if r.TmpDirName == "" {
		dirName = r.Repository.GetDirName(r.TemplateID)
	} else {
		dirName = r.TmpDirName
	}

	return c.run(r, GitRepositoryTmpPath,
		"clone",
		"--recursive",
		"--branch",
		r.Repository.GitBranch,
		r.Repository.GetGitURL(),
		dirName)
}

func (c CmdGitClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Repository " + r.Repository.GitURL)

	return c.run(r, GitRepositoryFullPath, "pull", "--recurse-submodules", "origin", r.Repository.GitBranch)
}

func (c CmdGitClient) Checkout(r GitRepository, target string) error {
	r.Logger.Log("Checkout repository to " + target)

	return c.run(r, GitRepositoryFullPath, "checkout", target)
}

func (c CmdGitClient) CanBePulled(r GitRepository) bool {
	err := c.run(r, GitRepositoryFullPath, "fetch")
	if err != nil {
		return false
	}

	err = c.run(r, GitRepositoryFullPath,
		"merge-base", "--is-ancestor", "HEAD", "origin/"+r.Repository.GitBranch)

	return err == nil
}

func (c CmdGitClient) GetLastCommitMessage(r GitRepository) (msg string, err error) {
	r.Logger.Log("Get current commit message")

	msg, err = c.output(r, GitRepositoryFullPath, "show-branch", "--no-name", "HEAD")
	if err != nil {
		return
	}

	if len(msg) > 100 {
		msg = msg[0:100]
	}

	return
}

func (c CmdGitClient) GetLastCommitHash(r GitRepository) (hash string, err error) {
	r.Logger.Log("Get current commit hash")
	hash, err = c.output(r, GitRepositoryFullPath, "rev-parse", "HEAD")
	return
}

func (c CmdGitClient) GetLastRemoteCommitHash(r GitRepository) (hash string, err error) {
	out, err := c.output(r, GitRepositoryTmpPath, "ls-remote", r.Repository.GetGitURL(), r.Repository.GitBranch)
	if err != nil {
		return
	}

	firstSpaceIndex := strings.IndexAny(out, "\t ")
	if firstSpaceIndex == -1 {
		err = fmt.Errorf("can't retreave remote commit hash")
	}
	if err != nil {
		return
	}

	hash = out[0:firstSpaceIndex]
	return
}
