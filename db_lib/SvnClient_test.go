package db_lib

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireSvn(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"svn", "svnadmin"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skip(bin + " is not installed")
		}
	}
}

func svnRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("svn", append([]string{"--non-interactive"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

// svnFixture is a Subversion repository in the conventional layout, and a
// working copy of its trunk to commit to it with. Revision 1 creates trunk,
// branches and tags, so the first commit is revision 2.
type svnFixture struct {
	root string // repository directory, for svnserve
	url  string // file:// URL of the repository root
	wc   string
}

func newSvnFixture(t *testing.T) svnFixture {
	t.Helper()
	requireSvn(t)

	dir := t.TempDir()
	root := filepath.Join(dir, "repo")
	out, err := exec.Command("svnadmin", "create", root).CombinedOutput()
	require.NoError(t, err, string(out))

	f := svnFixture{
		root: root,
		url:  "file://" + filepath.ToSlash(root),
		wc:   filepath.Join(dir, "wc"),
	}

	svnRun(t, dir, "mkdir", "-m", "layout", f.url+"/trunk", f.url+"/branches", f.url+"/tags")
	svnRun(t, dir, "checkout", f.url+"/trunk", f.wc)

	return f
}

// commit adds or changes a file of the trunk.
func (f svnFixture) commit(t *testing.T, file, content, msg string) {
	t.Helper()
	path := filepath.Join(f.wc, file)
	_, statErr := os.Stat(path)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	if os.IsNotExist(statErr) {
		svnRun(t, f.wc, "add", file)
	}
	svnRun(t, f.wc, "commit", "-m", msg)
}

// branch copies the trunk to the path, as svn copy does for a branch or a tag.
func (f svnFixture) branch(t *testing.T, path string) {
	t.Helper()
	svnRun(t, f.wc, "copy", "-m", "branch "+path, f.url+"/trunk", f.url+"/"+path+"@")
}

func newTestSvnRepo(t *testing.T, url, branch string) GitRepository {
	t.Helper()
	return GitRepository{
		Repository: db.Repository{
			ProjectID: 1,
			GitURL:    url,
			GitBranch: branch,
			SSHKey:    db.AccessKey{Type: db.AccessKeyNone},
		},
		Logger: task_logger.NopLogger{},
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func TestSvnClient_CloneUpdateCheckout(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second revision\n\nwith a body")

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "trunk")

	require.NoError(t, client.Clone(r))
	assert.Equal(t, "v2", readFile(t, filepath.Join(r.GetFullPath(), "site.yml")))

	hash, err := client.GetLastCommitHash(r)
	require.NoError(t, err)
	assert.Equal(t, "2", hash)

	msg, err := client.GetLastCommitMessage(r)
	require.NoError(t, err)
	assert.Equal(t, "second revision", msg)

	assert.True(t, client.CanBePulled(r))

	f.commit(t, "site.yml", "v3", "third revision")

	remote, err := client.GetLastRemoteCommitHash(r)
	require.NoError(t, err)
	assert.Equal(t, "3", remote)

	require.NoError(t, client.Pull(r))
	hash, err = client.GetLastCommitHash(r)
	require.NoError(t, err)
	assert.Equal(t, "3", hash)

	// A task re-run at the revision it recorded.
	require.NoError(t, client.Checkout(r, "2"))
	assert.Equal(t, "v2", readFile(t, filepath.Join(r.GetFullPath(), "site.yml")))
}

// svn refuses non-ASCII file names under the C locale on Linux; macOS takes
// file names as UTF-8 whatever the locale.
func TestSvnClient_NonASCIIFileName(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "déploiement.yml", "v2", "accentué")

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "trunk")

	require.NoError(t, client.Clone(r))
	assert.Equal(t, "v2", readFile(t, filepath.Join(r.GetFullPath(), "déploiement.yml")))

	msg, err := client.GetLastCommitMessage(r)
	require.NoError(t, err)
	assert.Equal(t, "accentué", msg)
}

func TestSvnClient_Locale(t *testing.T) {
	setupGitClientTest(t)
	r := newTestSvnRepo(t, "svn://svn.example.com/repo", "trunk")

	t.Run("utf-8 when no locale is set", func(t *testing.T) {
		cmd := SvnClient{}.makeCmd(r, GitRepositoryTmpPath, ssh.AccessKeyInstallation{}, "info")
		assert.Contains(t, cmd.Env, "LC_CTYPE=C.UTF-8")
	})

	t.Run("the administrator locale wins", func(t *testing.T) {
		util.Config.EnvVars = map[string]string{"LANG": "de_CH.UTF-8"}
		cmd := SvnClient{}.makeCmd(r, GitRepositoryTmpPath, ssh.AccessKeyInstallation{}, "info")
		assert.NotContains(t, cmd.Env, "LC_CTYPE=C.UTF-8")
		assert.Contains(t, cmd.Env, "LANG=de_CH.UTF-8")
	})
}

func TestSvnClient_Branch(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second")
	f.branch(t, "branches/release")
	f.commit(t, "site.yml", "v4", "trunk moves on")

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "branches/release")

	require.NoError(t, client.Clone(r))
	assert.Equal(t, "v2", readFile(t, filepath.Join(r.GetFullPath(), "site.yml")))

	remote, err := client.GetLastRemoteCommitHash(r)
	require.NoError(t, err)
	assert.Equal(t, "3", remote, "the commit to trunk is not a change of the branch")
}

// svn reads the text after the last "@" of a URL as a peg revision.
func TestSvnClient_BranchWithAt(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second")
	f.branch(t, "tags/v1@2")

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "tags/v1@2")

	require.NoError(t, client.Clone(r))
	assert.Equal(t, "v2", readFile(t, filepath.Join(r.GetFullPath(), "site.yml")))
	assert.True(t, client.CanBePulled(r))
}

func TestSvnClient_CommitOutsidePathIsNotAChange(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second")

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "trunk")

	f.branch(t, "branches/unrelated")

	remote, err := client.GetLastRemoteCommitHash(r)
	require.NoError(t, err)
	assert.Equal(t, "2", remote)
}

func TestSvnClient_GetRemoteBranches(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.branch(t, "branches/release")
	f.branch(t, "tags/v1")

	client := CreateSvnClient(nopKeyInstaller{})

	t.Run("conventional layout", func(t *testing.T) {
		branches, err := client.GetRemoteBranches(newTestSvnRepo(t, f.url, "trunk"))
		require.NoError(t, err)
		assert.Equal(t, []string{"trunk", "branches/release", "tags/v1"}, branches)
	})

	t.Run("other layout", func(t *testing.T) {
		branches, err := client.GetRemoteBranches(newTestSvnRepo(t, f.url+"/branches", "release"))
		require.NoError(t, err)
		assert.Equal(t, []string{"release"}, branches)
	})
}

func TestSvnClient_CanBePulled(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second")
	f.branch(t, "branches/release")

	client := CreateSvnClient(nopKeyInstaller{})

	t.Run("missing working copy", func(t *testing.T) {
		assert.False(t, client.CanBePulled(newTestSvnRepo(t, f.url, "trunk")))
	})

	t.Run("local modification", func(t *testing.T) {
		r := newTestSvnRepo(t, f.url, "trunk")
		r.TmpDirName = "modified"
		require.NoError(t, client.Clone(r))
		require.NoError(t, os.WriteFile(filepath.Join(r.GetFullPath(), "site.yml"), []byte("local"), 0644))
		assert.False(t, client.CanBePulled(r))
	})

	t.Run("other branch", func(t *testing.T) {
		r := newTestSvnRepo(t, f.url, "trunk")
		r.TmpDirName = "other"
		require.NoError(t, client.Clone(r))
		r.Repository.GitBranch = "branches/release"
		assert.False(t, client.CanBePulled(r))
	})

	t.Run("unversioned file", func(t *testing.T) {
		r := newTestSvnRepo(t, f.url, "trunk")
		r.TmpDirName = "unversioned"
		require.NoError(t, client.Clone(r))
		require.NoError(t, os.WriteFile(filepath.Join(r.GetFullPath(), "generated.retry"), []byte("x"), 0644))
		assert.True(t, client.CanBePulled(r))
	})
}

func TestSvnClient_RejectsInvalidBranch(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)

	client := CreateSvnClient(nopKeyInstaller{})

	for _, branch := range []string{"../escape", "trunk/../../escape", "-r1", "--config-dir=/tmp", "/abs", "a b"} {
		t.Run(branch, func(t *testing.T) {
			r := newTestSvnRepo(t, f.url, branch)
			assert.Error(t, client.Clone(r))
			_, err := client.GetLastRemoteCommitHash(r)
			assert.Error(t, err)
		})
	}
}

func TestSvnClient_RejectsInvalidCommit(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)

	client := CreateSvnClient(nopKeyInstaller{})
	r := newTestSvnRepo(t, f.url, "trunk")
	require.NoError(t, client.Clone(r))

	for _, target := range []string{"", "HEAD", "abcdef1", "-1", "--config-dir=/tmp", "1:2"} {
		t.Run(target, func(t *testing.T) {
			assert.Error(t, client.Checkout(r, target))
		})
	}
}

// startSvnserve serves the fixture with svnserve, readable only with the given
// credentials, and returns the svn:// URL of the repository root.
func startSvnserve(t *testing.T, f svnFixture, login, password string) string {
	t.Helper()
	if _, err := exec.LookPath("svnserve"); err != nil {
		t.Skip("svnserve is not installed")
	}

	conf := "[general]\nanon-access = none\nauth-access = write\npassword-db = passwd\n"
	require.NoError(t, os.WriteFile(filepath.Join(f.root, "conf", "svnserve.conf"), []byte(conf), 0644))
	passwd := fmt.Sprintf("[users]\n%s = %s\n", login, password)
	require.NoError(t, os.WriteFile(filepath.Join(f.root, "conf", "passwd"), []byte(passwd), 0644))

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	cmd := exec.Command("svnserve", "--daemon", "--foreground",
		"--listen-host", "127.0.0.1", "--listen-port", fmt.Sprint(port), "--root", f.root)
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		c, err := net.Dial("tcp", addr)
		if err == nil {
			_ = c.Close()
		}
		return err == nil
	}, 5*time.Second, 50*time.Millisecond)

	return "svn://" + addr
}

func TestSvnClient_LoginPassword(t *testing.T) {
	setupGitClientTest(t)
	f := newSvnFixture(t)
	f.commit(t, "site.yml", "v2", "second")

	const password = "s3cret pass"
	url := startSvnserve(t, f, "deploy", password)

	client := CreateSvnClient(nopKeyInstaller{})

	withKey := func(login, password string) GitRepository {
		r := newTestSvnRepo(t, url, "trunk")
		r.Repository.SSHKey = db.AccessKey{
			Type:          db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{Login: login, Password: password},
		}
		return r
	}

	t.Run("valid credentials", func(t *testing.T) {
		r := withKey("deploy", password)
		require.Equal(t, db.RepositorySVN, r.Repository.GetType())
		require.NoError(t, client.Clone(r))
		hash, err := client.GetLastCommitHash(r)
		require.NoError(t, err)
		assert.Equal(t, "2", hash)
		assert.True(t, client.CanBePulled(r), "svn prints the working copy URL canonicalized")
	})

	t.Run("password is not passed as an argument", func(t *testing.T) {
		r := withKey("deploy", password)
		cmd := SvnClient{}.makeCmd(r, GitRepositoryTmpPath, ssh.AccessKeyInstallation{}, "info")
		assert.NotContains(t, cmd.Args, password)
		assert.Contains(t, cmd.Args, "--password-from-stdin")
		assert.Contains(t, cmd.Args, "--no-auth-cache")
		assert.Contains(t, cmd.Args, "--non-interactive")
	})

	t.Run("wrong password", func(t *testing.T) {
		r := withKey("deploy", "wrong")
		r.TmpDirName = "wrong"
		assert.Error(t, client.Clone(r))
	})

	t.Run("no credentials", func(t *testing.T) {
		r := newTestSvnRepo(t, url, "trunk")
		_, err := client.GetRemoteBranches(r)
		assert.Error(t, err)
	})
}

func TestSvnClient_SshTunnel(t *testing.T) {
	setupGitClientTest(t)
	util.Config.Ssh = &util.SshConfig{StrictHostKeyChecking: util.SshStrictHostKeyCheckingAcceptNew, KnownHostsFile: "/tmp/known_hosts"}

	r := newTestSvnRepo(t, "svn+ssh://svn.example.com/repo/trunk", "HEAD")
	installation := ssh.AccessKeyInstallation{SSHAgent: &ssh.Agent{SocketFile: "/tmp/agent.sock"}}

	cmd := SvnClient{}.makeCmd(r, GitRepositoryTmpPath, installation, "info")

	assert.Contains(t, cmd.Env, "SSH_AUTH_SOCK=/tmp/agent.sock")
	assert.Contains(t, cmd.Env, "SVN_SSH=ssh -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=/tmp/known_hosts")
	for _, v := range cmd.Env {
		assert.NotContains(t, v, "GIT_SSH_COMMAND=")
	}
}

func TestCreateDefaultGitClient_RoutesSubversion(t *testing.T) {
	setupGitClientTest(t)

	tests := []struct {
		url      string
		expected GitClient
	}{
		{"svn://svn.example.com/repo/trunk", SvnClient{}},
		{"svn+ssh://svn.example.com/repo/trunk", SvnClient{}},
		{"https://github.com/semaphoreui/semaphore.git", CmdGitClient{}},
		{"git@github.com:semaphoreui/semaphore.git", CmdGitClient{}},
	}

	client, ok := CreateDefaultGitClient(nopKeyInstaller{}).(repositoryTypeClient)
	require.True(t, ok)

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			r := GitRepository{Repository: db.Repository{GitURL: tt.url}}
			assert.IsType(t, tt.expected, client.client(r))
		})
	}
}
