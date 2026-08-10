package db_lib

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"

	ssh2 "golang.org/x/crypto/ssh"
)

type GoGitClient struct {
	keyInstaller AccessKeyInstaller
}

type ProgressWrapper struct {
	Logger task_logger.Logger
}

func (t ProgressWrapper) Write(p []byte) (n int, err error) {
	s := string(p)

	if strings.HasPrefix(s, "Counting objects:") || strings.HasPrefix(s, "Compressing objects:") {
		return len(p), nil
	}

	t.Logger.Log(string(p))
	return len(p), nil
}

func (c GoGitClient) getAuthMethod(r GitRepository) (transport.AuthMethod, error) {
	return c.getAuthMethodForKey(r, r.Repository.SSHKey)
}

// getAuthMethodForKey builds a go-git auth method for an arbitrary access
// key, rather than always the repository's own SSHKey -- used to resolve a
// distinct credential per submodule.
func (c GoGitClient) getAuthMethodForKey(r GitRepository, key db.AccessKey) (transport.AuthMethod, error) {
	switch key.Type {
	case db.AccessKeySSH:

		install, err := c.keyInstaller.Install(key, db.AccessKeyRoleGit, r.Logger)
		if err != nil {
			return nil, err
		}

		defer install.Destroy()

		var sshKeyBuff = key.SshKey.PrivateKey

		if key.SshKey.Login == "" {
			key.SshKey.Login = "git"
		}

		publicKey, sshErr := ssh.NewPublicKeys(key.SshKey.Login, []byte(sshKeyBuff), key.SshKey.Passphrase)

		if sshErr != nil {
			return nil, sshErr
		}
		publicKey.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

		return publicKey, sshErr
	case db.AccessKeyLoginPassword:
		password := &http.BasicAuth{
			Username: key.LoginPassword.Login,
			Password: key.LoginPassword.Password,
		}

		return password, nil
	case db.AccessKeyNone:
		return nil, nil
	default:
		return nil, errors.New("unsupported auth method")
	}
}

func openRepository(r GitRepository, targetDir GitRepositoryDirType) (*git.Repository, error) {

	var dir string

	switch targetDir {
	case GitRepositoryTmpPath:
		dir = util.Config.GetProjectTmpDir(r.Repository.ProjectID)
	case GitRepositoryFullPath:
		dir = r.GetFullPath()
	default:
		panic("unknown Repository directory type")
	}

	return git.PlainOpen(dir)
}

func (c GoGitClient) Clone(r GitRepository) error {
	r.Logger.Log("Cloning Repository " + r.Repository.GitURL)

	authMethod, authErr := c.getAuthMethod(r)

	if authErr != nil {
		return authErr
	}

	cloneOpt := &git.CloneOptions{
		URL:               r.Repository.GetGitURL(true),
		Progress:          ProgressWrapper{r.Logger},
		RecurseSubmodules: git.NoRecurseSubmodules,
		ReferenceName:     plumbing.NewBranchReferenceName(r.Repository.GitBranch),
		Auth:              authMethod,
	}

	repo, err := git.PlainClone(r.GetFullPath(), false, cloneOpt)
	if err != nil {
		r.Logger.Log("Unable to clone repository: " + err.Error())
		return err
	}

	return c.updateSubmodules(r, repo)
}

func (c GoGitClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Repository " + r.Repository.GitURL)

	rep, err := openRepository(r, GitRepositoryFullPath)
	if err != nil {
		return err
	}

	wt, err := rep.Worktree()
	if err != nil {
		return err
	}

	authMethod, authErr := c.getAuthMethod(r)
	if authErr != nil {
		return authErr
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = wt.Pull(&git.PullOptions{RemoteName: "origin",
		Auth:              authMethod,
		RecurseSubmodules: git.NoRecurseSubmodules})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		r.Logger.Log("Unable to pull latest changes")
		return err
	}

	return c.updateSubmodules(r, rep)
}

// updateSubmodules recursively initializes and clones/updates every submodule
// of repo, resolving credentials per submodule host from
// r.SubmoduleCredentials (falling back to the repository's own SSHKey when a
// submodule's host has no explicit mapping -- today's behavior, unchanged).
//
// Unlike CloneOptions/PullOptions' built-in RecurseSubmodules (which reuses a
// single Auth for the entire submodule tree), each submodule is updated with
// its own resolved credential, so a submodule on a different host than the
// main repository is no longer forced to reuse the main repository's
// credentials.
func (c GoGitClient) updateSubmodules(r GitRepository, repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	submodules, err := wt.Submodules()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, sm := range submodules {
		key := resolveSubmoduleAccessKey(r.Repository.SSHKey, r.SubmoduleCredentials, sm.Config().URL)

		auth, err := c.getAuthMethodForKey(r, key)
		if err != nil {
			return err
		}

		err = sm.UpdateContext(ctx, &git.SubmoduleUpdateOptions{
			Init:              true,
			RecurseSubmodules: git.NoRecurseSubmodules,
			Auth:              auth,
		})
		if err != nil {
			return err
		}

		subRepo, err := sm.Repository()
		if err != nil {
			return err
		}

		if err := c.updateSubmodules(r, subRepo); err != nil {
			return err
		}
	}

	return nil
}

func (c GoGitClient) Checkout(r GitRepository, target string) error {
	r.Logger.Log("Checkout repository to " + target)

	rep, err := openRepository(r, GitRepositoryFullPath)
	if err != nil {
		return err
	}

	wt, err := rep.Worktree()

	if err != nil {
		return err
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(target),
	})

	return err
}

func (c GoGitClient) CanBePulled(r GitRepository) bool {

	rep, err := openRepository(r, GitRepositoryFullPath)
	if err != nil {
		return false
	}

	authMethod, err := c.getAuthMethod(r)
	if err != nil {
		return false
	}

	err = rep.Fetch(&git.FetchOptions{
		Auth: authMethod,
	})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return false
	}

	head, err := rep.Head()
	if err != nil {
		return false
	}

	headCommit, err := rep.CommitObject(head.Hash())
	if err != nil {
		return false
	}

	hash, err := rep.ResolveRevision(plumbing.Revision("origin/" + r.Repository.GitBranch))
	if err != nil {
		return false
	}

	lastCommit, err := rep.CommitObject(*hash)
	if err != nil {
		return false
	}

	isAncestor, err := headCommit.IsAncestor(lastCommit)
	return isAncestor && err == nil
}

func (c GoGitClient) GetLastCommitMessage(r GitRepository) (msg string, err error) {
	r.Logger.Log("Get current commit message")

	rep, err := openRepository(r, GitRepositoryFullPath)
	if err != nil {
		return
	}

	headRef, err := rep.Head()
	if err != nil {
		return
	}
	headCommit, err := rep.CommitObject(headRef.Hash())
	if err != nil {
		return
	}

	msg = truncateCommitMessage(headCommit.Message)

	r.Logger.Log("Message: " + msg)

	return
}

func (c GoGitClient) GetLastCommitHash(r GitRepository) (hash string, err error) {
	r.Logger.Log("Get current commit hash")

	rep, err := openRepository(r, GitRepositoryFullPath)
	if err != nil {
		return
	}

	headRef, err := rep.Head()
	if err != nil {
		return
	}
	hash = headRef.Hash().String()
	return
}

func (c GoGitClient) GetLastRemoteCommitHash(r GitRepository) (hash string, err error) {

	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{r.Repository.GitURL},
	})

	auth, err := c.getAuthMethod(r)
	if err != nil {
		return
	}

	refs, err := rem.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return
	}

	var lastRemoteRef *plumbing.Reference

	for _, rf := range refs {

		if rf.Name().Short() == r.Repository.GitBranch {
			lastRemoteRef = rf
		}
	}

	if lastRemoteRef != nil {
		hash = lastRemoteRef.Hash().String()
	}

	return
}

func (c GoGitClient) GetRemoteBranches(r GitRepository) ([]string, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{r.Repository.GitURL},
	})

	auth, err := c.getAuthMethod(r)

	if err != nil {
		return nil, fmt.Errorf("failed to create SSH auth method: %w", err)
	}

	listOptions := &git.ListOptions{}
	if auth != nil {
		listOptions.Auth = auth
	}

	refs, err := remote.List(listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote references: %w", err)
	}

	branches := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
	}
	return branches, nil
}
