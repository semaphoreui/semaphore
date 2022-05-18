package lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

type GitRepositoryDirType int

const (
	GitRepositoryTmpDir GitRepositoryDirType = iota
	GitRepositoryRepoDir
)

type GitRepository struct {
	TemplateID int
	Repository db.Repository
	Logger     Logger
}

func (r GitRepository) makeCmd(targetDir GitRepositoryDirType, args ...string) *exec.Cmd {
	cmd := exec.Command("git") //nolint: gas

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintln("GIT_TERMINAL_PROMPT=0"))
	if r.Repository.SSHKey.Type == db.AccessKeySSH {
		sshCmd := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i " + r.Repository.SSHKey.GetPath()
		if util.Config.SshConfigPath != "" {
			sshCmd += " -F " + util.Config.SshConfigPath
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
	}

	switch targetDir {
	case GitRepositoryTmpDir:
		cmd.Dir = util.Config.TmpPath
	case GitRepositoryRepoDir:
		cmd.Dir = r.GetFullPath()
	default:
		panic("unknown Repository directory type")
	}

	cmd.Args = append(cmd.Args, args...)

	return cmd
}

func (r GitRepository) run(targetDir GitRepositoryDirType, args ...string) error {
	err := r.Repository.SSHKey.Install(db.AccessKeyRoleGit)
	if err != nil {
		return err
	}

	defer r.Repository.SSHKey.Destroy() //nolint: errcheck

	cmd := r.makeCmd(targetDir, args...)

	r.Logger.LogCmd(cmd)

	return cmd.Run()
}

func (r GitRepository) output(targetDir GitRepositoryDirType, args ...string) (out string, err error) {
	err = r.Repository.SSHKey.Install(db.AccessKeyRoleGit)
	if err != nil {
		return
	}

	defer r.Repository.SSHKey.Destroy() //nolint: errcheck

	bytes, err := r.makeCmd(targetDir, args...).Output()
	if err != nil {
		return
	}
	out = strings.Trim(string(bytes), " \n")
	return
}

func (r GitRepository) Clone() error {
	r.Logger.Log("Cloning Repository " + r.Repository.GitURL)

	return r.run(GitRepositoryTmpDir,
		"clone",
		"--recursive",
		"--branch",
		r.Repository.GitBranch,
		r.Repository.GetGitURL(),
		r.Repository.GetDirName(r.TemplateID))
}

func (r GitRepository) Pull() error {
	r.Logger.Log("Updating Repository " + r.Repository.GitURL)

	return r.run(GitRepositoryRepoDir, "pull", "origin", r.Repository.GitBranch)
}

func (r GitRepository) Checkout(target string) error {
	r.Logger.Log("Checkout repository to " + target)

	return r.run(GitRepositoryRepoDir, "checkout", target)
}

func (r GitRepository) CanBePulled() bool {
	err := r.run(GitRepositoryRepoDir, "fetch")
	if err != nil {
		return false
	}

	err = r.run(GitRepositoryRepoDir,
		"merge-base", "--is-ancestor", "HEAD", "origin/"+r.Repository.GitBranch)

	return err == nil
}

func (r GitRepository) GetLastCommitMessage() (msg string, err error) {
	r.Logger.Log("Get current commit message")

	msg, err = r.output(GitRepositoryRepoDir, "show-branch", "--no-name", "HEAD")
	if err != nil {
		return
	}

	if len(msg) > 100 {
		msg = msg[0:100]
	}

	return
}

func (r GitRepository) GetLastCommitHash() (hash string, err error) {
	r.Logger.Log("Get current commit hash")
	hash, err = r.output(GitRepositoryRepoDir, "rev-parse", "HEAD")
	return
}

func (r GitRepository) ValidateRepo() error {
	_, err := os.Stat(r.GetFullPath())
	return err
}

func (r GitRepository) GetFullPath() (path string) {
	path = r.Repository.GetFullPath(r.TemplateID)
	return
}

func (r GitRepository) GetLastRemoteCommitHash() (hash string, err error) {
	out, err := r.output(GitRepositoryTmpDir, "ls-remote", r.Repository.GetGitURL(), r.Repository.GitBranch)
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
