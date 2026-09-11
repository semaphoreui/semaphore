package db_lib

import (
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepositoryBranchNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "simple branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"def456\trefs/heads/develop",
			},
			expected: []string{"main", "develop"},
		},
		{
			name: "branch names with slashes",
			input: []string{
				"abc123\trefs/heads/env/test",
				"def456\trefs/heads/feature/my-feature",
			},
			expected: []string{"env/test", "feature/my-feature"},
		},
		{
			name: "mixed branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"def456\trefs/heads/env/test",
				"ghi789\trefs/heads/release/v1.0",
			},
			expected: []string{"main", "env/test", "release/v1.0"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "skip lines without tab",
			input: []string{
				"invalid line",
				"abc123\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "skip non-heads refs",
			input: []string{
				"abc123\trefs/tags/v1.0",
				"def456\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "trailing whitespace in ref path",
			input: []string{
				"abc123\trefs/heads/env/test\n",
			},
			expected: []string{"env/test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRepositoryBranchNames(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d branches, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, branch := range result {
				if branch != tt.expected[i] {
					t.Errorf("branch[%d]: expected %q, got %q", i, tt.expected[i], branch)
				}
			}
		})
	}
}

func TestCmdGitClient_SpecialCharAuthAndProxyBypass(t *testing.T) {
	// 1. Create a bare git repo
	repoDir := t.TempDir()
	gitExec := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	srcDir := filepath.Join(repoDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	gitExec(srcDir, "init", "-q", "-b", "main")
	gitExec(srcDir, "config", "user.email", "test@test.com")
	gitExec(srcDir, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("# Test Repo"), 0644))
	gitExec(srcDir, "add", "README.md")
	gitExec(srcDir, "commit", "-qm", "initial commit")

	bareDir := filepath.Join(repoDir, "repo.git")
	gitExec(repoDir, "clone", "--bare", srcDir, bareDir)
	gitExec(bareDir, "config", "http.receivepack", "true")

	// 2. Setup Git HTTP server with Basic Auth containing special characters
	expectedUser := "user@domain.com"
	expectedPass := "p#ss%word@123"

	gitBackendPath, err := exec.LookPath("git-http-backend")
	if err != nil {
		gitExecPathCmd := exec.Command("git", "--exec-path")
		out, _ := gitExecPathCmd.Output()
		gitBackendPath = filepath.Join(strings.TrimSpace(string(out)), "git-http-backend")
	}

	if _, err := os.Stat(gitBackendPath); err != nil {
		t.Skip("git-http-backend not found, skipping HTTP auth integration test")
	}

	cgiHandler := &cgi.Handler{
		Path: gitBackendPath,
		Env: []string{
			"GIT_PROJECT_ROOT=" + repoDir,
			"GIT_HTTP_EXPORT_ALL=1",
		},
		InheritEnv: []string{"PATH", "USER", "SYSTEMROOT"},
	}

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != expectedUser || p != expectedPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Git"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		cgiHandler.ServeHTTP(w, r)
	})

	gitServer := httptest.NewTLSServer(authHandler)
	defer gitServer.Close()

	// 3. Setup dummy proxy server that should NOT be contacted when NO_PROXY is active
	var proxyHitCount int32
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&proxyHitCount, 1)
		http.Error(w, "Proxy should not be used", http.StatusBadGateway)
	}))
	defer proxyServer.Close()

	setupGitClientTest(t)

	// 4. Configure environment with proxy and NO_PROXY bypass
	t.Setenv("HTTP_PROXY", proxyServer.URL)
	t.Setenv("HTTPS_PROXY", proxyServer.URL)
	t.Setenv("NO_PROXY", "127.0.0.1,localhost")
	t.Setenv("GIT_SSL_NO_VERIFY", "true")

	repo := db.Repository{
		ProjectID: 1,
		GitURL:    gitServer.URL + "/repo.git",
		GitBranch: "main",
		SSHKey: db.AccessKey{
			Type: db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{
				Login:    expectedUser,
				Password: expectedPass,
			},
		},
	}

	client := CmdGitClient{
		keyInstaller: nopKeyInstaller{},
	}

	gitRepo := GitRepository{
		Repository: repo,
		Client:     client,
		Logger:     task_logger.NopLogger{},
	}

	// 4. Control phase: forward proxy variables without NO_PROXY.
	// Git MUST attempt to contact the proxy server, hit it, and fail.
	util.Config.ForwardedEnvVars = []string{"GIT_SSL_NO_VERIFY", "HTTP_PROXY", "HTTPS_PROXY"}
	_, err = client.GetRemoteBranches(gitRepo)
	require.Error(t, err, "Traffic without NO_PROXY forwarded should attempt to use the proxy and fail")
	controlHits := atomic.LoadInt32(&proxyHitCount)
	assert.Greater(t, controlHits, int32(0), "Proxy server should have been contacted during control run")

	// 5. Bypass phase: forward NO_PROXY as well.
	// Git MUST recognize NO_PROXY, bypass the proxy server, and succeed.
	util.Config.ForwardedEnvVars = []string{"GIT_SSL_NO_VERIFY", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"}
	branches, err := client.GetRemoteBranches(gitRepo)
	require.NoError(t, err)
	assert.Contains(t, branches, "main")

	// 6. Test Clone with special chars and NO_PROXY bypass
	err = client.Clone(gitRepo)
	require.NoError(t, err)

	// 7. Verify dummy proxy was bypassed (no additional hits beyond control run)
	assert.Equal(t, controlHits, atomic.LoadInt32(&proxyHitCount), "Traffic should have bypassed the proxy due to NO_PROXY")

	// 8. Test Pull with non-positive GitSubmoduleJobs (0 and negative)
	util.Config.GitSubmoduleJobs = 0
	err = client.Pull(gitRepo)
	require.NoError(t, err)

	util.Config.GitSubmoduleJobs = -1
	err = client.Pull(gitRepo)
	require.NoError(t, err)
}

func TestCmdGitClient_MakeCmd_HomePrecedence(t *testing.T) {
	setupGitClientTest(t)

	// Set ambient HOME
	t.Setenv("HOME", "/ambient/home")

	client := CmdGitClient{keyInstaller: nopKeyInstaller{}}
	repo := db.Repository{ProjectID: 1}
	gitRepo := GitRepository{Repository: repo, Client: client}

	// 1. Default case: ambient HOME is used when not overridden
	util.Config.EnvVars = map[string]string{}
	cmd := client.makeCmd(gitRepo, GitRepositoryTmpPath, ssh.AccessKeyInstallation{})
	assert.True(t, containsPrefix(cmd.Env, "HOME=/ambient/home"))

	// 2. Explicit Config.EnvVars overrides ambient HOME and avoids duplicates
	util.Config.EnvVars = map[string]string{
		"HOME": "/custom/config/home",
	}
	cmdExplicit := client.makeCmd(gitRepo, GitRepositoryTmpPath, ssh.AccessKeyInstallation{})
	assert.True(t, containsPrefix(cmdExplicit.Env, "HOME=/custom/config/home"))
	assert.False(t, containsPrefix(cmdExplicit.Env, "HOME=/ambient/home"))

	homeCount := 0
	for _, env := range cmdExplicit.Env {
		k, _, _ := strings.Cut(env, "=")
		if strings.EqualFold(k, "HOME") {
			homeCount++
		}
	}
	assert.Equal(t, 1, homeCount, "HOME should appear exactly once in cmd.Env")
}
