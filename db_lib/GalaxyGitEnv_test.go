package db_lib

import (
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func httpRepo(gitURL, login, password string) db.Repository {
	return db.Repository{
		GitURL: gitURL,
		SSHKey: db.AccessKey{
			Type:          db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{Login: login, Password: password},
		},
	}
}

func TestSqQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain", "abc", "'abc'"},
		{"empty", "", "''"},
		{"embedded quote", "a'b", `'a'\''b'`},
		{"only quote", "'", `''\'''`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sqQuote(tt.input))
		})
	}
}

func TestGalaxyGitEnv_AlwaysDisablesTerminalPrompt(t *testing.T) {
	for _, repo := range []db.Repository{
		{GitURL: "git@github.com:acme/roles.git"},
		{GitURL: "https://git.private.repo/acme/roles.git"},
		httpRepo("https://git.private.repo/acme/roles.git", "u", "p"),
	} {
		assert.Contains(t, galaxyGitEnv(repo), "GIT_TERMINAL_PROMPT=0")
	}
}

func TestGalaxyGitEnv_HTTPSWithLoginPassword(t *testing.T) {
	env := galaxyGitEnv(httpRepo("https://git.private.repo/acme/roles.git", "semuser", "sempass"))

	require.Len(t, env, 2)
	assert.Equal(t,
		`GIT_CONFIG_PARAMETERS='url.https://semuser:sempass@git.private.repo/.insteadOf=https://git.private.repo/'`,
		env[1])
}

// The rewrite must match host and port, or git will not apply it.
func TestGalaxyGitEnv_KeepsPort(t *testing.T) {
	env := galaxyGitEnv(httpRepo("http://127.0.0.1:3300/semuser/main-repo.git", "semuser", "sempass"))

	require.Len(t, env, 2)
	assert.Contains(t, env[1], "url.http://semuser:sempass@127.0.0.1:3300/.insteadOf=http://127.0.0.1:3300/")
}

// Matches GetGitURL: an empty login means the password is the whole credential.
func TestGalaxyGitEnv_TokenOnlyKeyUsesPasswordAsUser(t *testing.T) {
	env := galaxyGitEnv(httpRepo("https://git.private.repo/acme/roles.git", "", "gho_token"))

	require.Len(t, env, 2)
	assert.Contains(t, env[1], "url.https://gho_token@git.private.repo/.insteadOf=")
}

// A password containing '@' or '/' would otherwise corrupt the URL.
func TestGalaxyGitEnv_EncodesCredentials(t *testing.T) {
	env := galaxyGitEnv(httpRepo("https://git.private.repo/acme/roles.git", "user@corp", "p@ss/w:rd"))

	require.Len(t, env, 2)
	assert.Contains(t, env[1], "user%40corp:p%40ss%2Fw%3Ard@git.private.repo")
	assert.NotContains(t, env[1], "p@ss/w:rd")
}

func TestGalaxyGitEnv_NoCredentialsForOtherRepoTypes(t *testing.T) {
	tests := []struct {
		name string
		repo db.Repository
	}{
		{"ssh url", db.Repository{
			GitURL: "git@github.com:acme/roles.git",
			SSHKey: db.AccessKey{Type: db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{Login: "u", Password: "p"}},
		}},
		{"https url but ssh key", db.Repository{
			GitURL: "https://git.private.repo/acme/roles.git",
			SSHKey: db.AccessKey{Type: db.AccessKeySSH},
		}},
		{"https url but no key", db.Repository{
			GitURL: "https://git.private.repo/acme/roles.git",
			SSHKey: db.AccessKey{Type: db.AccessKeyNone},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := galaxyGitEnv(tt.repo)
			assert.Equal(t, []string{"GIT_TERMINAL_PROMPT=0"}, env)
		})
	}
}

// requirements.yml may name several servers; only the repository's own host
// may ever be offered its credential.
func TestGalaxyGitEnv_ScopesCredentialToOneHost(t *testing.T) {
	env := galaxyGitEnv(httpRepo("https://git.private.repo/acme/roles.git", "semuser", "sempass"))

	require.Len(t, env, 2)
	assert.Contains(t, env[1], ".insteadOf=https://git.private.repo/")
	assert.NotContains(t, env[1], "acme/roles.git")
}

type fakeInstaller struct {
	key   db.AccessKey
	usage db.AccessKeyRole
	env   []string
	err   error
	calls int
}

func (f *fakeInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, _ task_logger.Logger) (ssh.AccessKeyInstallation, error) {
	f.key, f.usage = key, usage
	f.calls++
	return ssh.AccessKeyInstallation{}, f.err
}

// setupGalaxyConfig gives the package-level util.Config a temp dir to resolve
// repository paths against, and restores it afterwards.
func setupGalaxyConfig(t *testing.T) {
	original := util.Config
	t.Cleanup(func() { util.Config = original })
	util.Config = &util.ConfigType{TmpPath: t.TempDir(), Process: &util.ConfigProcess{}}
}

// stubGalaxy puts a succeeding ansible-galaxy first on PATH and returns a
// function reporting how many times it ran. Without it the real binary runs and
// fails, so InstallRequirements returns before the second requirements file and
// nothing is reused.
func stubGalaxy(t *testing.T) func() int {
	t.Helper()

	dir := t.TempDir()
	runLog := path.Join(dir, "runs")
	script := "#!/bin/sh\necho run >> " + runLog + "\n"
	require.NoError(t, os.WriteFile(path.Join(dir, "ansible-galaxy"), []byte(script), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return func() int {
		content, err := os.ReadFile(runLog)
		if err != nil {
			return 0
		}
		return strings.Count(string(content), "run")
	}
}

// newGalaxyApp builds the app the way AppFactory does, so that runGalaxy has a
// Playbook to run.
func newGalaxyApp(repo db.Repository) *AnsibleApp {
	logger := task_logger.NopLogger{}

	return &AnsibleApp{
		Logger:     logger,
		Repository: repo,
		Playbook:   &AnsiblePlaybook{Repository: repo, Logger: logger},
	}
}

// writeRequirements puts a requirements.yml where the app looks for it, so that
// galaxy actually runs. Without one every install is skipped.
func writeRequirements(t *testing.T, app *AnsibleApp) {
	t.Helper()

	dir := app.getRepoPath()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(path.Join(dir, "requirements.yml"), []byte("collections: []\n"), 0o644))
}

// The repository's own key must be the one galaxy gets, under the git role.
func TestInstallRequirements_InstallsRepositoryKey(t *testing.T) {
	setupGalaxyConfig(t)

	inst := &fakeInstaller{}
	app := newGalaxyApp(db.Repository{SSHKey: db.AccessKey{ID: 42, Type: db.AccessKeySSH}})
	writeRequirements(t, app)
	galaxyRuns := stubGalaxy(t)

	require.NoError(t, app.InstallRequirements(LocalAppInstallingArgs{Installer: inst}))

	require.Greater(t, galaxyRuns(), 1, "reuse is only meaningful across more than one galaxy run")
	assert.Equal(t, 42, inst.key.ID)
	assert.Equal(t, db.AccessKeyRole(db.AccessKeyRoleGit), inst.usage)
	assert.Equal(t, 1, inst.calls, "one installation must be reused across requirements files")
}

// Nothing for galaxy to install means no key is decrypted and no agent started.
func TestInstallRequirements_NoRequirementsFileInstallsNoKey(t *testing.T) {
	setupGalaxyConfig(t)

	inst := &fakeInstaller{}
	app := newGalaxyApp(db.Repository{SSHKey: db.AccessKey{ID: 42, Type: db.AccessKeySSH}})

	require.NoError(t, app.InstallRequirements(LocalAppInstallingArgs{Installer: inst}))

	assert.Zero(t, inst.calls)
}

func TestInstallRequirements_FailsWhenKeyInstallFails(t *testing.T) {
	setupGalaxyConfig(t)

	app := newGalaxyApp(db.Repository{})
	writeRequirements(t, app)

	err := app.InstallRequirements(LocalAppInstallingArgs{
		Installer: &fakeInstaller{err: errors.New("agent unavailable")},
	})

	assert.ErrorContains(t, err, "agent unavailable")
}

// A nil installer is the remote-runner path; it must not panic.
func TestInstallRequirements_NilInstaller(t *testing.T) {
	setupGalaxyConfig(t)

	app := newGalaxyApp(db.Repository{})

	assert.NoError(t, app.InstallRequirements(LocalAppInstallingArgs{}))
}
