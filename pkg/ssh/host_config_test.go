package ssh

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupHostConfig points the package globals at a temp dir. Not t.TempDir():
// agent sockets are unix sockets, whose path is limited to about 104 bytes.
func setupHostConfig(t *testing.T) string {
	t.Helper()

	tmp, err := os.MkdirTemp("/tmp", "hc")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmp) }) //nolint:errcheck

	original := util.Config
	t.Cleanup(func() { util.Config = original })

	util.Config = &util.ConfigType{
		TmpPath: tmp,
		Ssh:     &util.SshConfig{StrictHostKeyChecking: util.SshStrictHostKeyCheckingNo},
	}

	return tmp
}

func sshKey(t *testing.T, id int, name string) db.AccessKey {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "",
		"-f", filepath.Join(dir, "k")).Run())

	private, err := os.ReadFile(filepath.Join(dir, "k"))
	require.NoError(t, err)

	return db.AccessKey{
		ID: id, Name: name, Type: db.AccessKeySSH,
		SshKey: db.SshKey{PrivateKey: string(private)},
	}
}

// A project with no mappings must behave exactly as it does today.
func TestInstallHostConfigs_NoMappings(t *testing.T) {
	setupHostConfig(t)

	installation, err := InstallHostConfigs(1, nil, task_logger.NopLogger{})

	require.NoError(t, err)
	assert.Nil(t, installation)
	assert.Empty(t, installation.SSHConfigPath())
	assert.Empty(t, installation.GitConfigParameters())

	// A nil installation must not change the environment either.
	key := AccessKeyInstallation{}
	assert.Equal(t, key.GetGitEnv(), key.GetGitEnvWithHostConfigs(nil))
}

// The generated config must bind each mapped host to its own agent, which is
// what ssh itself has to agree with.
func TestInstallHostConfigs_HostMappingResolvedBySSH(t *testing.T) {
	setupHostConfig(t)

	mappings := []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "github.com", SSHKey: sshKey(t, 1, "gh")},
		{ID: 2, ProjectID: 1, Type: db.HostConfigHost, Name: "gitlab.com", SSHKey: sshKey(t, 2, "gl")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	require.FileExists(t, installation.SSHConfigPath())

	githubAgent := installation.Agents[0].SocketFile
	gitlabAgent := installation.Agents[1].SocketFile
	require.NotEqual(t, githubAgent, gitlabAgent)

	assert.Equal(t, githubAgent, resolveOption(t, installation.SSHConfigPath(), "github.com", "identityagent"))
	assert.Equal(t, gitlabAgent, resolveOption(t, installation.SSHConfigPath(), "gitlab.com", "identityagent"))
}

// A URL mapping reaches its credential through an alias, so two mappings on one
// host keep their own key.
func TestInstallHostConfigs_URLMappingUsesItsOwnAlias(t *testing.T) {
	setupHostConfig(t)

	mappings := []db.HostConfig{
		{ID: 7, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/private/", SSHKey: sshKey(t, 1, "a")},
		{ID: 8, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/other/", SSHKey: sshKey(t, 2, "b")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	cfg := installation.SSHConfigPath()

	// Both aliases point at the real host, each with its own agent.
	assert.Equal(t, "github.com", resolveOption(t, cfg, "semaphore-mapping-7", "hostname"))
	assert.Equal(t, "github.com", resolveOption(t, cfg, "semaphore-mapping-8", "hostname"))
	assert.Equal(t, "git", resolveOption(t, cfg, "semaphore-mapping-7", "user"))

	first := resolveOption(t, cfg, "semaphore-mapping-7", "identityagent")
	second := resolveOption(t, cfg, "semaphore-mapping-8", "identityagent")
	assert.NotEqual(t, first, second, "two mappings on one host must not share a credential")
}

// The administrator's config must keep applying to everything the project does
// not map, which is what "existing projects behave unchanged" rests on.
func TestInstallHostConfigs_IncludesAdminConfig(t *testing.T) {
	tmp := setupHostConfig(t)

	adminConfig := filepath.Join(tmp, "admin_ssh_config")
	require.NoError(t, os.WriteFile(adminConfig,
		[]byte("Host *\n  ServerAliveInterval 60\n"), 0600))
	util.Config.SshConfigPath = adminConfig

	mappings := []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "github.com", SSHKey: sshKey(t, 1, "gh")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	cfg := installation.SSHConfigPath()

	t.Run("a mapped host keeps its mapping and gains the admin settings", func(t *testing.T) {
		assert.Equal(t, installation.Agents[0].SocketFile,
			resolveOption(t, cfg, "github.com", "identityagent"))
		assert.Equal(t, "60", resolveOption(t, cfg, "github.com", "serveraliveinterval"))
	})

	t.Run("an unmapped host still gets the admin settings", func(t *testing.T) {
		assert.Equal(t, "60", resolveOption(t, cfg, "gitlab.com", "serveraliveinterval"))
	})
}

// git is the only thing which decides whether a rewrite installs, and it
// refuses a parameter whose key holds an unescaped "=".
func TestInstallHostConfigs_RewritesParsedByGit(t *testing.T) {
	setupHostConfig(t)

	mappings := []db.HostConfig{
		{ID: 7, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/private/", SSHKey: sshKey(t, 1, "a")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	params := installation.GitConfigParameters()
	require.NotEmpty(t, params)

	dir := t.TempDir()
	cmd := exec.Command("git", "config", "--get-regexp", "^url\\.")
	cmd.Dir = dir
	// An isolated HOME: a developer's own gitconfig also holds url.* rewrites.
	cmd.Env = []string{
		"GIT_CONFIG_PARAMETERS=" + params,
		"HOME=" + dir,
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	}

	out, err := cmd.CombinedOutput()

	require.NoError(t, err, "git could not parse the rewrite: %s", out)
	assert.Contains(t, string(out),
		"url.git@semaphore-mapping-7:acme/private/.insteadof https://github.com/acme/private/")
}

// A more specific URL must win over a broader one. git resolves this by longest
// match, so the test pins the behaviour rather than an implementation.
func TestInstallHostConfigs_MoreSpecificURLWins(t *testing.T) {
	setupHostConfig(t)

	mappings := []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/", SSHKey: sshKey(t, 1, "broad")},
		{ID: 2, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/private/", SSHKey: sshKey(t, 2, "specific")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	dir := t.TempDir()
	cmd := exec.Command("git", "ls-remote", "--get-url",
		"https://github.com/acme/private/repo.git")
	cmd.Dir = dir
	cmd.Env = []string{
		"GIT_CONFIG_PARAMETERS=" + installation.GitConfigParameters(),
		"HOME=" + dir,
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	}

	out, err := cmd.CombinedOutput()

	require.NoError(t, err, "%s", out)
	assert.Equal(t, "git@semaphore-mapping-2:acme/private/repo.git", strings.TrimSpace(string(out)),
		"the mapping of the repository must win over the mapping of the group")
}

// Nothing may outlive the task: the config holds the paths of live agents.
func TestInstallHostConfigs_DestroyRemovesEverything(t *testing.T) {
	setupHostConfig(t)

	mappings := []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "github.com", SSHKey: sshKey(t, 1, "gh")},
	}

	installation, err := InstallHostConfigs(1, mappings, task_logger.NopLogger{})
	require.NoError(t, err)

	configFile := installation.SSHConfigPath()
	socket := installation.Agents[0].SocketFile
	require.FileExists(t, configFile)

	installation.Destroy()

	assert.NoFileExists(t, configFile)
	assert.NoFileExists(t, socket)
}

// resolveOption asks ssh itself what a host resolves to, rather than matching
// the generated text.
func resolveOption(t *testing.T, configFile string, host string, option string) string {
	t.Helper()

	out, err := exec.Command("ssh", "-F", configFile, "-G", host).Output()
	require.NoError(t, err)

	for _, line := range strings.Split(string(out), "\n") {
		name, value, found := strings.Cut(line, " ")
		if found && name == option {
			return value
		}
	}

	return ""
}
