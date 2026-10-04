package db_lib

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/git"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/svn"
	"github.com/semaphoreui/semaphore/util"

	log "github.com/sirupsen/logrus"
)

// SvnClient serves Subversion repositories through the svn binary.
//
// The repository URL is the repository root, and the branch is a path under it
// (trunk, branches/x, tags/y), so a template or a task selects a branch the way
// it does for git. The last revision which changed that path stands in for the
// commit hash.
type SvnClient struct {
	keyInstaller AccessKeyInstaller
}

func CreateSvnClient(keyInstaller AccessKeyInstaller) GitClient {
	return SvnClient{
		keyInstaller: keyInstaller,
	}
}

// svnBranchURL returns the URL of the branch path, without a peg revision.
//
// The branch is validated as a git branch name when the repository, template
// or task is saved, which also keeps it from climbing out of the root ("..")
// or carrying options; it is checked again here because the playbook browser
// takes it from the query string.
func svnBranchURL(r GitRepository) (string, error) {
	if err := git.ValidateGitBranch(r.Repository.GitBranch, "repository"); err != nil {
		return "", err
	}

	root := strings.TrimRight(r.Repository.GetGitURL(false), "/")
	if r.Repository.GitBranch == "" {
		return root, nil
	}
	return root + "/" + r.Repository.GitBranch, nil
}

// svnPeg ends a URL with an empty peg revision. svn reads the text after the
// last "@" of a URL as its peg revision, so a path containing "@" would
// otherwise be cut there.
func svnPeg(u string) string {
	return u + "@"
}

// svnGlobalArgs are given to every svn command. Prompting would hang the task,
// and a cached password would outlive the access key it came from.
func svnGlobalArgs(r GitRepository) []string {
	args := []string{"--non-interactive", "--no-auth-cache"}

	if r.Repository.SSHKey.Type == db.AccessKeyLoginPassword {
		if login := r.Repository.SSHKey.LoginPassword.Login; login != "" {
			args = append(args, "--username", login)
		}
		if r.Repository.SSHKey.LoginPassword.Password != "" {
			// From stdin, not --password: the arguments of a process are
			// visible to every user of the host.
			args = append(args, "--password-from-stdin")
		}
	}

	return args
}

func (c SvnClient) makeCmd(
	r GitRepository,
	targetDir GitRepositoryDirType,
	installation ssh.AccessKeyInstallation,
	args ...string,
) *exec.Cmd {
	cmd := exec.Command("svn") //nolint:gosec

	cmd.Env = getEnvironmentVars()

	// svn+ssh:// runs the command in SVN_SSH, which gets the agent, the host
	// key checking and the project's host configs git gets.
	for _, v := range installation.GetGitEnvWithHostConfigs(r.HostConfigs) {
		if sshCmd, ok := strings.CutPrefix(v, "GIT_SSH_COMMAND="); ok {
			cmd.Env = append(cmd.Env, "SVN_SSH="+sshCmd)
		} else if strings.HasPrefix(v, "SSH_AUTH_SOCK=") {
			cmd.Env = append(cmd.Env, v)
		}
	}

	// Same rule as git: never an empty HOME, which would hide the svn
	// configuration of the user Semaphore runs as.
	if !hasNonEmptyEnvVar(cmd.Env, "HOME") {
		if homeDir := getHomeDir(r.Repository, r.TemplateID); homeDir != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", homeDir))
		} else if h := os.Getenv("HOME"); h != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", h))
		}
	}
	appendPlatformEnv(&cmd.Env)

	// The environment carries no locale of its own. Without one, svn writes
	// translated messages in a legacy encoding the task log shows as garbage,
	// and on Linux it refuses to check out a file whose name is not ASCII
	// ("Can't convert string from 'UTF-8' to native encoding"). A locale set
	// by the administrator wins.
	if !hasNonEmptyEnvVar(cmd.Env, "LC_ALL") && !hasNonEmptyEnvVar(cmd.Env, "LC_CTYPE") &&
		!hasNonEmptyEnvVar(cmd.Env, "LANG") {
		cmd.Env = append(cmd.Env, "LC_CTYPE=C.UTF-8")
	}

	switch targetDir {
	case GitRepositoryTmpPath:
		cmd.Dir = util.Config.GetProjectTmpDir(r.Repository.ProjectID)
		if err := os.MkdirAll(cmd.Dir, 0755); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "svn",
			}).Error("failed to create project temp directory")
		}
	case GitRepositoryFullPath:
		cmd.Dir = r.GetFullPath()
	default:
		panic("unknown Repository directory type")
	}

	cmd.Args = append(cmd.Args, svnGlobalArgs(r)...)
	cmd.Args = append(cmd.Args, args...)

	if r.Repository.SSHKey.Type == db.AccessKeyLoginPassword && r.Repository.SSHKey.LoginPassword.Password != "" {
		cmd.Stdin = strings.NewReader(r.Repository.SSHKey.LoginPassword.Password)
	}

	cmd.SysProcAttr = util.Config.GetSysProcAttr()

	return cmd
}

func (c SvnClient) run(r GitRepository, targetDir GitRepositoryDirType, args ...string) error {
	keyInstallation, err := c.keyInstaller.Install(r.Repository.SSHKey, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return err
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	cmd := c.makeCmd(r, targetDir, keyInstallation, args...)

	finishLog := r.Logger.LogCmd(cmd)
	defer finishLog()

	// The task log gets stderr through LogCmd; the error keeps its end too,
	// for callers whose logger discards it, such as the repository API.
	stderr := &svnStderrTail{}
	if cmd.Stderr != nil {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, stderr)
	} else {
		cmd.Stderr = stderr
	}

	return svnError(cmd.Run(), stderr.buf)
}

func (c SvnClient) output(r GitRepository, targetDir GitRepositoryDirType, args ...string) (out string, err error) {
	keyInstallation, err := c.keyInstaller.Install(r.Repository.SSHKey, db.AccessKeyRoleGit, r.Logger)
	if err != nil {
		return
	}

	defer keyInstallation.Destroy() //nolint: errcheck

	bytes, err := c.makeCmd(r, targetDir, keyInstallation, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			err = svnError(err, exitErr.Stderr)
		}
		return
	}
	out = strings.TrimSpace(string(bytes))
	return
}

// svnStderrMax bounds the stderr kept for an error. svn writes a few short
// lines when it fails; a long checkout can write much more before that.
const svnStderrMax = 1024

// svnStderrTail keeps the end of what svn writes to stderr.
type svnStderrTail struct {
	buf []byte
}

func (t *svnStderrTail) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)
	if len(t.buf) > svnStderrMax {
		t.buf = t.buf[len(t.buf)-svnStderrMax:]
	}
	return len(p), nil
}

// svnUserinfo matches the userinfo of a URL quoted in an svn message.
var svnUserinfo = regexp.MustCompile(`(\w[\w+.-]*://)[^\s/@']*@`)

// svnError adds what svn said to a failed command's error, which is only an
// exit status, with the userinfo of quoted URLs removed. svn never prints a
// password, but a URL may carry one typed into it.
func svnError(err error, stderr []byte) error {
	if err == nil {
		return nil
	}

	lines := []string{}
	for _, line := range strings.Split(string(stderr), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, svnUserinfo.ReplaceAllString(line, "$1"))
		}
	}

	if len(lines) == 0 {
		return err
	}

	return fmt.Errorf("%w: %s", err, strings.Join(lines, " / "))
}

func (c SvnClient) Clone(r GitRepository) error {
	r.Logger.Log("Checking out Subversion repository " + r.Repository.GetRedactedGitURL())

	branchURL, err := svnBranchURL(r)
	if err != nil {
		return err
	}

	var dirName string
	if r.TmpDirName == "" {
		dirName = r.Repository.GetCheckoutDirName(r.TemplateID)
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

	return c.run(r, GitRepositoryTmpPath, "checkout", "--", svnPeg(branchURL), dirName)
}

func (c SvnClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Subversion repository " + r.Repository.GetRedactedGitURL())

	return c.run(r, GitRepositoryFullPath, "update")
}

func (c SvnClient) Checkout(r GitRepository, target string) error {
	r.Logger.Log("Updating Subversion repository to revision " + target)

	if target == "" {
		return fmt.Errorf("task commit hash is empty")
	}
	if err := svn.ValidateRevision(target, "task"); err != nil {
		return err
	}

	return c.run(r, GitRepositoryFullPath, "update", "--revision", target)
}

// CanBePulled reports whether the working copy can be updated in place: it is
// a checkout of the same branch and has no local modifications. Unlike git
// pull, svn update merges into local modifications and reports a conflict as
// success, so a modified working copy is checked out again instead.
func (c SvnClient) CanBePulled(r GitRepository) bool {
	branchURL, err := svnBranchURL(r)
	if err != nil {
		return false
	}

	// Repository edits clear the cache of this server only; a runner may
	// still hold a working copy of the URL the repository had before.
	wcURL, err := c.output(r, GitRepositoryFullPath, "info", "--show-item", "url")
	if err != nil || !svn.SameURL(wcURL, branchURL) {
		return false
	}

	status, err := c.output(r, GitRepositoryFullPath, "status", "--quiet")
	return err == nil && status == ""
}

type svnLog struct {
	Entries []struct {
		Revision string `xml:"revision,attr"`
		Msg      string `xml:"msg"`
	} `xml:"logentry"`
}

func (c SvnClient) GetLastCommitMessage(r GitRepository) (msg string, err error) {
	r.Logger.Log("Get current commit message")

	revision, err := c.GetLastCommitHash(r)
	if err != nil {
		return
	}

	out, err := c.output(r, GitRepositoryFullPath, "log", "--xml", "--limit", "1", "--revision", revision)
	if err != nil {
		return
	}

	var entries svnLog
	if err = xml.Unmarshal([]byte(out), &entries); err != nil {
		return
	}

	if len(entries.Entries) == 0 {
		return
	}

	// show-branch prints the subject line only; do the same.
	msg, _, _ = strings.Cut(strings.TrimSpace(entries.Entries[0].Msg), "\n")
	msg = truncateCommitMessage(msg)

	return
}

// GetLastCommitHash returns the last revision which changed the checked out
// path, not the revision of the whole repository, so a commit elsewhere in the
// repository does not look like a change to this one.
func (c SvnClient) GetLastCommitHash(r GitRepository) (hash string, err error) {
	r.Logger.Log("Get current commit hash")
	return c.output(r, GitRepositoryFullPath, "info", "--show-item", "last-changed-revision")
}

func (c SvnClient) GetLastRemoteCommitHash(r GitRepository) (hash string, err error) {
	branchURL, err := svnBranchURL(r)
	if err != nil {
		return
	}

	return c.output(r, GitRepositoryTmpPath,
		"info", "--show-item", "last-changed-revision", "--", svnPeg(branchURL))
}

// GetRemoteBranches lists the branch paths of the repository root. In the
// conventional layout these are trunk and the children of branches and tags;
// otherwise every directory of the root is a branch.
func (c SvnClient) GetRemoteBranches(r GitRepository) ([]string, error) {
	root := strings.TrimRight(r.Repository.GetGitURL(false), "/")

	top, err := c.listDirs(r, root)
	if err != nil {
		return nil, err
	}

	// trunk first: it is the branch most repositories are used with.
	branches := []string{}
	standard := false

	for _, dir := range top {
		switch dir {
		case "trunk":
			branches = append([]string{dir}, branches...)
			standard = true
		case "branches", "tags":
			children, err := c.listDirs(r, root+"/"+dir)
			if err != nil {
				return nil, err
			}
			for _, child := range children {
				branches = append(branches, dir+"/"+child)
			}
			standard = true
		}
	}

	if !standard {
		return top, nil
	}

	return branches, nil
}

// listDirs returns the names of the directories directly under the URL.
func (c SvnClient) listDirs(r GitRepository, dirURL string) ([]string, error) {
	out, err := c.output(r, GitRepositoryTmpPath, "list", "--", svnPeg(dirURL))
	if err != nil {
		return nil, err
	}

	dirs := []string{}
	for _, line := range strings.Split(out, "\n") {
		if name, ok := strings.CutSuffix(strings.TrimSpace(line), "/"); ok && name != "" {
			dirs = append(dirs, name)
		}
	}
	return dirs, nil
}
