package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/ssh"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"

	log "github.com/sirupsen/logrus"
)

type CmdGitClient struct {
	keyInstaller AccessKeyInstaller
}

func (c CmdGitClient) makeCmd(
	r GitRepository,
	targetDir GitRepositoryDirType,
	installation ssh.AccessKeyInstallation,
	proxyInstallation ssh.AccessKeyInstallation,
	args ...string,
) *exec.Cmd {
	cmd := exec.Command("git") //nolint: gas

	cmd.Env = append(getEnvironmentVars(), installation.GetGitEnv(gitProxyOpts(r.Repository, proxyInstallation)...)...)
	cmd.Env = append(cmd.Env, gitProxyEnv(r.Repository)...)

	switch targetDir {
	case GitRepositoryTmpPath:
		cmd.Dir = util.Config.GetProjectTmpDir(r.Repository.ProjectID)
		_, err := os.Stat(cmd.Dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(cmd.Dir, 0755)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"context": "git",
					}).Error("failed to create project temp directory")
				}
			} else {
				log.WithError(err).WithFields(log.Fields{
					"context": "git",
				}).Error("failed to check existing project temp directory")
			}
		}
	case GitRepositoryFullPath:
		cmd.Dir = r.GetFullPath()
	default:
		panic("unknown Repository directory type")
	}

	cmd.Args = append(cmd.Args, args...)

	cmd.SysProcAttr = util.Config.GetSysProcAttr()

	return cmd
}

// installProxyKey installs the keys of the repository proxy chain into their own
// agent. The proxies are usually reached with different keys than the git server,
// and that agent must not be the git one: ssh stops at the first key the git host
// accepts and would never try the other one.
func (c CmdGitClient) installProxyKey(r GitRepository) (installation ssh.AccessKeyInstallation, err error) {
	proxy := r.Repository.Proxy

	if proxy == nil || proxy.Type != db.ProxySSH {
		return
	}

	keys := ProxyChainKeys(*proxy)
	if len(keys) == 0 {
		return
	}

	return c.keyInstaller.InstallAll(keys, db.AccessKeyRoleGit, r.Logger)
}

func (c CmdGitClient) run(r GitRepository, targetDir GitRepositoryDirType, args ...string) error {
	var err error
	keyInstallation, err := c.keyInstaller.Install(r.Repository.SSHKey, db.AccessKeyRoleGit, r.Logger)

	if err != nil {
		return err
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	proxyKeyInstallation, err := c.installProxyKey(r)
	if err != nil {
		return err
	}

	defer proxyKeyInstallation.Destroy() //nolint: errcheck

	cmd := c.makeCmd(r, targetDir, keyInstallation, proxyKeyInstallation, args...)

	finishLog := r.Logger.LogCmd(cmd)
	defer finishLog()

	return cmd.Run()
}

func (c CmdGitClient) output(r GitRepository, targetDir GitRepositoryDirType, args ...string) (out string, err error) {
	keyInstallation, err := c.keyInstaller.Install(r.Repository.SSHKey, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	proxyKeyInstallation, err := c.installProxyKey(r)
	if err != nil {
		return
	}

	defer proxyKeyInstallation.Destroy() //nolint: errcheck

	bytes, err := c.makeCmd(r, targetDir, keyInstallation, proxyKeyInstallation, args...).Output()
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

	targetPath := r.GetFullPath()
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return err
	}
	if err := util.ChownDir(targetPath); err != nil {
		return err
	}

	return c.run(r, GitRepositoryTmpPath,
		"clone",
		"--recursive",
		"--branch",
		r.Repository.GitBranch,
		"--end-of-options",
		r.Repository.GetGitURL(false),
		dirName)
}

func (c CmdGitClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Repository " + r.Repository.GitURL)

	err := c.run(r, GitRepositoryFullPath, "pull", "origin", "--end-of-options", r.Repository.GitBranch)
	if err != nil {
		return err
	}
	return c.run(r, GitRepositoryFullPath, "submodule", "update", "--init", "--recursive")
}

func (c CmdGitClient) Checkout(r GitRepository, target string) error {
	r.Logger.Log("Checkout repository to " + target)

	return c.run(r, GitRepositoryFullPath, "checkout", "--end-of-options", target)
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

	msg = truncateCommitMessage(msg)

	return
}

func (c CmdGitClient) GetLastCommitHash(r GitRepository) (hash string, err error) {
	r.Logger.Log("Get current commit hash")
	hash, err = c.output(r, GitRepositoryFullPath, "rev-parse", "HEAD")
	return
}

func (c CmdGitClient) GetLastRemoteCommitHash(r GitRepository) (hash string, err error) {
	out, err := c.output(r, GitRepositoryTmpPath, "ls-remote", "--end-of-options", r.Repository.GetGitURL(false), r.Repository.GitBranch)
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

func (c CmdGitClient) GetRemoteBranches(r GitRepository) ([]string, error) {
	out, err := c.output(r, GitRepositoryTmpPath, "ls-remote", "--heads", "--end-of-options", r.Repository.GetGitURL(false))
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return []string{}, nil
	}

	branches := strings.Split(out, "\n")
	branchNames := getRepositoryBranchNames(branches)
	return branchNames, nil
}

func getRepositoryBranchNames(branches []string) []string {
	branchNames := make([]string, 0, len(branches))

	for _, branch := range branches {
		parts := strings.Split(branch, "\t")
		if len(parts) < 2 {
			continue
		}

		refPath := strings.TrimSpace(parts[1])

		const refsHeadsPrefix = "refs/heads/"
		if strings.HasPrefix(refPath, refsHeadsPrefix) {
			branchName := strings.TrimPrefix(refPath, refsHeadsPrefix)
			branchNames = append(branchNames, branchName)
		}
	}

	return branchNames
}

// gitProxyOpts returns the ssh options needed to reach the git server of the
// repository through its proxy chain.
func gitProxyOpts(repo db.Repository, proxyInstallation ssh.AccessKeyInstallation) []string {
	if repo.Proxy == nil {
		return nil
	}

	// An http(s) repository is proxied by git itself, see gitProxyEnv.
	if repo.GetType() == db.RepositoryHTTP {
		return nil
	}

	var socket string
	if proxyInstallation.SSHAgent != nil {
		socket = proxyInstallation.SSHAgent.SocketFile
	}

	return []string{"-o", strconv.Quote(ProxyCommandOption(*repo.Proxy, socket))}
}

// gitProxyEnv returns the environment the proxy of a repository needs. git
// speaks SOCKS and HTTP proxies itself for http(s) remotes, so those are
// configured with the variables it reads rather than with a ProxyCommand.
func gitProxyEnv(repo db.Repository) []string {
	if repo.Proxy == nil || repo.Proxy.Type.IsSSH() {
		return nil
	}

	if repo.GetType() != db.RepositoryHTTP {
		// An ssh remote goes through the connector, which reads its credentials
		// from the environment.
		return ProxyEnv(repo.Proxy)
	}

	proxyURL := ProxyURLWithCredentials(*repo.Proxy)

	return []string{
		"ALL_PROXY=" + proxyURL,
		"HTTP_PROXY=" + proxyURL,
		"HTTPS_PROXY=" + proxyURL,
	}
}
