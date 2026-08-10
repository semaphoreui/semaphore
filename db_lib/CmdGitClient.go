package db_lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
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
	args ...string,
) *exec.Cmd {
	cmd := exec.Command("git") //nolint: gas

	cmd.Env = append(getEnvironmentVars(), installation.GetGitEnv()...)

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

func (c CmdGitClient) run(r GitRepository, targetDir GitRepositoryDirType, args ...string) error {
	return c.runWithKey(r, targetDir, r.Repository.SSHKey, args...)
}

func (c CmdGitClient) runWithKey(r GitRepository, targetDir GitRepositoryDirType, key db.AccessKey, args ...string) error {
	keyInstallation, err := c.keyInstaller.Install(key, db.AccessKeyRoleGit, r.Logger)

	if err != nil {
		return err
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	cmd := c.makeCmd(r, targetDir, keyInstallation, args...)

	r.Logger.LogCmd(cmd)

	return cmd.Run()
}

func (c CmdGitClient) output(r GitRepository, targetDir GitRepositoryDirType, args ...string) (out string, err error) {
	keyInstallation, err := c.keyInstaller.Install(r.Repository.SSHKey, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	bytes, err := c.makeCmd(r, targetDir, keyInstallation, args...).Output()
	if err != nil {
		return
	}
	out = strings.Trim(string(bytes), " \n")
	return
}

// makeCmdInDir builds a git command rooted at an arbitrary directory, used to
// drive submodule operations scoped to a specific submodule's own working
// directory (which GitRepositoryDirType's two cases can't express).
func (c CmdGitClient) makeCmdInDir(dir string, installation ssh.AccessKeyInstallation, args ...string) *exec.Cmd {
	cmd := exec.Command("git") //nolint: gas

	cmd.Env = append(getEnvironmentVars(), installation.GetGitEnv()...)
	cmd.Dir = dir
	cmd.Args = append(cmd.Args, args...)
	cmd.SysProcAttr = util.Config.GetSysProcAttr()

	return cmd
}

func (c CmdGitClient) runInDir(r GitRepository, dir string, key db.AccessKey, args ...string) error {
	keyInstallation, err := c.keyInstaller.Install(key, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return err
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	cmd := c.makeCmdInDir(dir, keyInstallation, args...)

	r.Logger.LogCmd(cmd)

	return cmd.Run()
}

func (c CmdGitClient) outputInDir(r GitRepository, dir string, key db.AccessKey, args ...string) (out string, err error) {
	keyInstallation, err := c.keyInstaller.Install(key, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	bytes, err := c.makeCmdInDir(dir, keyInstallation, args...).Output()
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

	err := c.run(r, GitRepositoryTmpPath,
		"clone",
		"--branch",
		r.Repository.GitBranch,
		"--end-of-options",
		r.Repository.GetGitURL(false),
		dirName)
	if err != nil {
		return err
	}

	return c.updateSubmodules(r, r.GetFullPath())
}

func (c CmdGitClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Repository " + r.Repository.GitURL)

	err := c.run(r, GitRepositoryFullPath, "pull", "origin", "--end-of-options", r.Repository.GitBranch)
	if err != nil {
		return err
	}
	return c.updateSubmodules(r, r.GetFullPath())
}

// gitSubmodule describes one entry read from a repository's tracked
// .gitmodules file: its configured name (the "[submodule "name"]" header,
// used as the local .git/config key), its checkout path, and its remote URL.
type gitSubmodule struct {
	Name string
	Path string
	URL  string
}

var submoduleConfigKeyRe = regexp.MustCompile(`^submodule\.(.+)\.(path|url)$`)

// listSubmodules reads submodule name/path/url triples straight from the
// tracked .gitmodules file in dir (not from local .git/config, which only
// gains a submodule's url once it has been initialized -- and initialization
// is exactly the step this package needs to intercept per-submodule in order
// to inject the right credentials before update runs).
func (c CmdGitClient) listSubmodules(r GitRepository, dir string) ([]gitSubmodule, error) {
	if _, statErr := os.Stat(path.Join(dir, ".gitmodules")); statErr != nil {
		return nil, nil
	}

	out, err := c.outputInDir(r, dir, r.Repository.SSHKey,
		"config", "--file", ".gitmodules", "--get-regexp", `^submodule\..*\.(path|url)$`)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// No submodule entries matched -- an empty/malformed .gitmodules.
			return nil, nil
		}
		return nil, err
	}

	byName := map[string]*gitSubmodule{}
	var order []string

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		m := submoduleConfigKeyRe.FindStringSubmatch(parts[0])
		if m == nil {
			continue
		}

		name, field, value := m[1], m[2], parts[1]

		sm, ok := byName[name]
		if !ok {
			sm = &gitSubmodule{Name: name}
			byName[name] = sm
			order = append(order, name)
		}

		if field == "path" {
			sm.Path = value
		} else {
			sm.URL = value
		}
	}

	submodules := make([]gitSubmodule, 0, len(order))
	for _, name := range order {
		sm := *byName[name]
		if sm.Path == "" {
			sm.Path = sm.Name
		}
		if sm.URL == "" {
			continue
		}
		submodules = append(submodules, sm)
	}

	return submodules, nil
}

// updateSubmodules recursively initializes and clones/updates every submodule
// found under dir, resolving credentials per submodule host from
// r.SubmoduleCredentials (falling back to the repository's own SSHKey when a
// submodule's host has no explicit mapping -- today's behavior, unchanged).
//
// Unlike `git clone --recursive` / `git submodule update --recursive`, each
// submodule is cloned with its own dedicated credential installation (its own
// ssh-agent, or its own URL-embedded HTTPS credentials), so a submodule on a
// different host than the main repository is no longer forced to reuse the
// main repository's credentials.
func (c CmdGitClient) updateSubmodules(r GitRepository, dir string) error {
	submodules, err := c.listSubmodules(r, dir)
	if err != nil {
		return err
	}

	for _, sm := range submodules {
		key := resolveSubmoduleAccessKey(r.Repository.SSHKey, r.SubmoduleCredentials, sm.URL)

		// Override the submodule's local URL before init/update runs, so an
		// HTTPS credential is embedded exactly like the main repository's
		// GetGitURL(false) does -- never touching the tracked .gitmodules file.
		effectiveURL := gitURLWithCredentials(sm.URL, key)
		if err := c.runInDir(r, dir, r.Repository.SSHKey,
			"config", "submodule."+sm.Name+".url", effectiveURL); err != nil {
			return err
		}

		if err := c.updateSubmoduleWithRetry(r, dir, key, sm.Path); err != nil {
			return err
		}

		submoduleDir := path.Join(dir, sm.Path)
		if err := c.updateSubmodules(r, submoduleDir); err != nil {
			return err
		}
	}

	return nil
}

// updateSubmoduleWithRetry clones/updates a single submodule at subPath
// (relative to dir), retrying once on failure. This preserves the resilience
// git's own `--recursive` clone provides (it retries a failed submodule clone
// once before aborting).
func (c CmdGitClient) updateSubmoduleWithRetry(r GitRepository, dir string, key db.AccessKey, subPath string) error {
	err := c.runInDir(r, dir, key, "submodule", "update", "--init", "--checkout", "--", subPath)
	if err == nil {
		return nil
	}

	r.Logger.Log(fmt.Sprintf("Failed to clone '%s'. Retry scheduled", subPath))

	err = c.runInDir(r, dir, key, "submodule", "update", "--init", "--checkout", "--", subPath)
	if err != nil {
		r.Logger.Log(fmt.Sprintf("Failed to clone '%s' a second time, aborting", subPath))
	}

	return err
}

// gitURLWithCredentials embeds login:password@ into an http(s) URL for key,
// mirroring db.Repository.GetGitURL(false) but generalized to an arbitrary
// URL and access key so it can be applied per-submodule. Non-HTTP(S) URLs and
// non-login/password keys are returned unchanged.
func gitURLWithCredentials(rawURL string, key db.AccessKey) string {
	if key.Type != db.AccessKeyLoginPassword {
		return rawURL
	}

	m := httpsURLPrefixRe.FindStringSubmatch(rawURL)
	if m == nil {
		return rawURL
	}

	auth := ""
	if key.LoginPassword.Login == "" {
		auth = key.LoginPassword.Password
	} else {
		auth = key.LoginPassword.Login + ":" + key.LoginPassword.Password
	}

	if auth == "" {
		return rawURL
	}

	protocol := m[1]
	return protocol + "://" + auth + "@" + rawURL[len(protocol)+3:]
}

var httpsURLPrefixRe = regexp.MustCompile(`^(https?)://`)

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
