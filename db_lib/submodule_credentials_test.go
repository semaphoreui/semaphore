package db_lib

import (
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runGit runs a git command in dir, failing the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

// initBareSubmoduleRepo creates a bare repo at dir and populates it with one
// commit, by pushing from a scratch working clone -- the same shape a real
// submodule host would have.
func initBareSubmoduleRepo(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0755))
	runGit(t, dir, "init", "-q", "--bare", "-b", "main")

	work := t.TempDir()
	runGit(t, work, "clone", "-q", dir, ".")
	runGit(t, work, "config", "user.email", "t@t")
	runGit(t, work, "config", "user.name", "t")
	require.NoError(t, os.WriteFile(filepath.Join(work, "submodule-file.txt"), []byte("submodule content"), 0644))
	runGit(t, work, "add", "submodule-file.txt")
	runGit(t, work, "commit", "-qm", "init submodule")
	runGit(t, work, "push", "-q", "origin", "main")
}

// lastCommitHash returns the current HEAD hash of the repo at dir.
func lastCommitHash(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	return string(out[:40])
}

// addGitlinkSubmodule registers a submodule entry (.gitmodules + gitlink) in
// the repo at mainDir without cloning it -- avoiding the chicken-and-egg
// problem of `git submodule add` needing working credentials at setup time.
func addGitlinkSubmodule(t *testing.T, mainDir, subPath, subURL, subSHA string) {
	t.Helper()

	gitmodulesPath := filepath.Join(mainDir, ".gitmodules")
	entry := "[submodule \"" + subPath + "\"]\n\tpath = " + subPath + "\n\turl = " + subURL + "\n"

	existing, err := os.ReadFile(gitmodulesPath)
	if err == nil {
		entry = string(existing) + entry
	}
	require.NoError(t, os.WriteFile(gitmodulesPath, []byte(entry), 0644))

	runGit(t, mainDir, "add", ".gitmodules")
	runGit(t, mainDir, "update-index", "--add", "--cacheinfo", "160000,"+subSHA+","+subPath)
	runGit(t, mainDir, "commit", "-qm", "add submodule "+subPath)
}

// startAuthedGitHTTPServer serves every bare repo under reposRoot over smart
// HTTP via the real `git http-backend` CGI, rejecting any request that
// doesn't carry exactly login/password as HTTP Basic Auth -- standing in for
// a git host that needs different credentials than the main repository.
func startAuthedGitHTTPServer(t *testing.T, reposRoot, login, password string) *httptest.Server {
	t.Helper()

	gitPath, err := exec.LookPath("git")
	require.NoError(t, err)

	backend := &cgi.Handler{
		Path: gitPath,
		Args: []string{"http-backend"},
		Dir:  reposRoot,
		Env: []string{
			"GIT_PROJECT_ROOT=" + reposRoot,
			"GIT_HTTP_EXPORT_ALL=1",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != login || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="git"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	return server
}

// submoduleCredentialFixture wires up the exact shape of issue #4134: a main
// repository (no credentials needed, served over the filesystem) whose
// .gitmodules references a submodule on a different host that requires
// login/password credentials the main repository doesn't have.
type submoduleCredentialFixture struct {
	MainDir      string
	SubmoduleURL string
	Login        string
	Password     string
}

func newSubmoduleCredentialFixture(t *testing.T) submoduleCredentialFixture {
	t.Helper()

	const login = "subuser"
	const password = "subpass" //nolint:gosec

	reposRoot := t.TempDir()
	subDir := filepath.Join(reposRoot, "submodule.git")
	initBareSubmoduleRepo(t, subDir)
	subSHA := lastCommitHash(t, subDir)

	server := startAuthedGitHTTPServer(t, reposRoot, login, password)
	submoduleURL, err := url.JoinPath(server.URL, "submodule.git")
	require.NoError(t, err)

	mainDir := t.TempDir()
	gitInit(t, mainDir)
	addGitlinkSubmodule(t, mainDir, "scripts/submodule1", submoduleURL, subSHA)

	return submoduleCredentialFixture{
		MainDir:      mainDir,
		SubmoduleURL: submoduleURL,
		Login:        login,
		Password:     password,
	}
}

func (f submoduleCredentialFixture) submoduleHost(t *testing.T) string {
	t.Helper()
	u, err := url.Parse(f.SubmoduleURL)
	require.NoError(t, err)
	return u.Host
}

// TestCmdGitClient_SubmoduleCredentials_Unmatched reproduces GitHub issue
// #4134: cloning a repository whose submodule needs different credentials
// than the main repository fails on the submodule, exactly as reported,
// when no submodule credential is configured for that host.
func TestCmdGitClient_SubmoduleCredentials_Unmatched(t *testing.T) {
	setupGitClientTest(t)
	fixture := newSubmoduleCredentialFixture(t)

	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, fixture.MainDir, "main")
	gitRepo := GitRepository{
		Repository: repo.Repository,
		TmpDirName: "clone",
		Logger:     task_logger.NopLogger{},
		Client:     client,
	}

	err := gitRepo.Clone()
	assert.Error(t, err, "submodule clone must fail without matching credentials")
}

// TestCmdGitClient_SubmoduleCredentials_Matched proves the fix: once a
// RepositorySubmoduleCredential mapping the submodule's host to its own
// access key is configured, the clone succeeds end-to-end even though the
// main repository uses no credentials at all.
func TestCmdGitClient_SubmoduleCredentials_Matched(t *testing.T) {
	setupGitClientTest(t)
	fixture := newSubmoduleCredentialFixture(t)

	client := CreateCmdGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, fixture.MainDir, "main")
	gitRepo := GitRepository{
		Repository: repo.Repository,
		TmpDirName: "clone",
		SubmoduleCredentials: []db.RepositorySubmoduleCredential{
			{
				Host: fixture.submoduleHost(t),
				AccessKey: db.AccessKey{
					Type: db.AccessKeyLoginPassword,
					LoginPassword: db.LoginPassword{
						Login:    fixture.Login,
						Password: fixture.Password,
					},
				},
			},
		},
		Logger: task_logger.NopLogger{},
		Client: client,
	}

	err := gitRepo.Clone()
	require.NoError(t, err, "submodule clone must succeed once its host has a matching credential")

	content, err := os.ReadFile(filepath.Join(gitRepo.GetFullPath(), "scripts", "submodule1", "submodule-file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "submodule content", string(content))
}

// TestGoGitClient_SubmoduleCredentials_Unmatched is the go-git backend's
// counterpart of TestCmdGitClient_SubmoduleCredentials_Unmatched, proving
// switching SEMAPHORE_GIT_CLIENT to go_git does not sidestep issue #4134.
func TestGoGitClient_SubmoduleCredentials_Unmatched(t *testing.T) {
	setupGitClientTest(t)
	fixture := newSubmoduleCredentialFixture(t)

	client := CreateGoGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, fixture.MainDir, "main")
	gitRepo := GitRepository{
		Repository: repo.Repository,
		TmpDirName: "clone",
		Logger:     task_logger.NopLogger{},
		Client:     client,
	}

	err := gitRepo.Clone()
	assert.Error(t, err, "submodule clone must fail without matching credentials")
}

// TestGoGitClient_SubmoduleCredentials_Matched is the go-git backend's
// counterpart of TestCmdGitClient_SubmoduleCredentials_Matched.
func TestGoGitClient_SubmoduleCredentials_Matched(t *testing.T) {
	setupGitClientTest(t)
	fixture := newSubmoduleCredentialFixture(t)

	client := CreateGoGitClient(nopKeyInstaller{})
	repo := newTestGitRepo(t, fixture.MainDir, "main")
	gitRepo := GitRepository{
		Repository: repo.Repository,
		TmpDirName: "clone",
		SubmoduleCredentials: []db.RepositorySubmoduleCredential{
			{
				Host: fixture.submoduleHost(t),
				AccessKey: db.AccessKey{
					Type: db.AccessKeyLoginPassword,
					LoginPassword: db.LoginPassword{
						Login:    fixture.Login,
						Password: fixture.Password,
					},
				},
			},
		},
		Logger: task_logger.NopLogger{},
		Client: client,
	}

	err := gitRepo.Clone()
	require.NoError(t, err, "submodule clone must succeed once its host has a matching credential")

	content, err := os.ReadFile(filepath.Join(gitRepo.GetFullPath(), "scripts", "submodule1", "submodule-file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "submodule content", string(content))
}
