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

// TestInstallHostConfigs_PrefixMatchesOnStringBoundary pins how git matches a
// URL mapping: insteadOf is a plain string prefix, with no notion of a path
// segment. A prefix that does not end in "/" therefore also captures its
// siblings, which is why the interface tells the user to end one with "/".
func TestInstallHostConfigs_PrefixMatchesOnStringBoundary(t *testing.T) {
	setupHostConfig(t)

	install := func(t *testing.T, mapped string) *HostConfigInstallation {
		t.Helper()
		installation, err := InstallHostConfigs(1, []db.HostConfig{{
			ID: 1, ProjectID: 1, Type: db.HostConfigURL,
			Name: mapped, SSHKey: sshKey(t, 1, "k"),
		}}, task_logger.NopLogger{})
		require.NoError(t, err)
		t.Cleanup(installation.Destroy)
		return installation
	}

	rewritten := func(t *testing.T, installation *HostConfigInstallation, url string) string {
		t.Helper()
		dir := t.TempDir()
		cmd := exec.Command("git", "ls-remote", "--get-url", url)
		cmd.Dir = dir
		cmd.Env = []string{
			"GIT_CONFIG_PARAMETERS=" + installation.GitConfigParameters(),
			"HOME=" + dir, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		}
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", out)
		return strings.TrimSpace(string(out))
	}

	t.Run("a trailing slash keeps the mapping inside the group", func(t *testing.T) {
		installation := install(t, "https://github.com/acme/private/")

		assert.Equal(t, "git@semaphore-mapping-1:acme/private/repo.git",
			rewritten(t, installation, "https://github.com/acme/private/repo.git"))

		// A sibling group must keep its own URL.
		assert.Equal(t, "https://github.com/acme/private-other/repo.git",
			rewritten(t, installation, "https://github.com/acme/private-other/repo.git"))
	})

	t.Run("without a trailing slash a sibling is captured too", func(t *testing.T) {
		installation := install(t, "https://github.com/acme/private")

		assert.Equal(t, "git@semaphore-mapping-1:acme/private-other.git",
			rewritten(t, installation, "https://github.com/acme/private-other.git"),
			"git matches insteadOf as a plain prefix, so a sibling is captured")
	})
}

func loginPasswordKey(id int, login string, password string) db.AccessKey {
	return db.AccessKey{
		ID: id, Type: db.AccessKeyLoginPassword,
		LoginPassword: db.LoginPassword{Login: login, Password: password},
	}
}

// A URL mapping may authenticate over https with a login and a password, in
// which case no ssh identity is involved at all.
func TestInstallHostConfigs_LoginPasswordURLMapping(t *testing.T) {
	setupHostConfig(t)

	installation, err := InstallHostConfigs(1, []db.HostConfig{{
		ID: 1, ProjectID: 1, Type: db.HostConfigURL,
		Name: "https://test.asdf.ru/", SSHKey: loginPasswordKey(1, "bob", "s3cr3t"),
	}}, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	assert.Empty(t, installation.Agents, "a login/password needs no ssh agent")

	params := installation.GitConfigParameters()
	assert.Contains(t, params, "url.https://bob:s3cr3t@test.asdf.ru/")
	assert.Contains(t, params, ".insteadOf=https://test.asdf.ru/")
}

// git is the only thing which decides whether the rewrite installs, and it
// refuses a parameter whose key holds an unescaped "=" — common in tokens.
func TestInstallHostConfigs_LoginPasswordParsedByGit(t *testing.T) {
	setupHostConfig(t)

	tests := []struct {
		name     string
		login    string
		password string
		expected string
	}{
		{"login and password", "bob", "s3cr3t", "https://bob:s3cr3t@test.asdf.ru/"},
		{"token only", "", "ghp_abcdef", "https://ghp_abcdef@test.asdf.ru/"},
		{"equals in the password", "bob", "tok=en", "https://bob:tok%3Den@test.asdf.ru/"},
		{"characters needing escaping", "user@corp", "p@ss/w:rd",
			"https://user%40corp:p%40ss%2Fw%3Ard@test.asdf.ru/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installation, err := InstallHostConfigs(1, []db.HostConfig{{
				ID: 1, ProjectID: 1, Type: db.HostConfigURL,
				Name: "https://test.asdf.ru/", SSHKey: loginPasswordKey(1, tt.login, tt.password),
			}}, task_logger.NopLogger{})
			require.NoError(t, err)
			defer installation.Destroy()

			dir := t.TempDir()
			cmd := exec.Command("git", "config", "--get-regexp", "^url\\.")
			cmd.Dir = dir
			cmd.Env = []string{
				"GIT_CONFIG_PARAMETERS=" + installation.GitConfigParameters(),
				"HOME=" + dir, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			}

			out, err := cmd.CombinedOutput()

			require.NoError(t, err, "git could not parse the rewrite: %s", out)
			assert.Contains(t, string(out), "url."+tt.expected+".insteadof https://test.asdf.ru/")
		})
	}
}

// The mapping types must not interfere: an ssh URL mapping still gets an agent
// and an alias, a login/password one gets neither.
func TestInstallHostConfigs_MixedCredentialTypes(t *testing.T) {
	setupHostConfig(t)

	installation, err := InstallHostConfigs(1, []db.HostConfig{
		{ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "github.com", SSHKey: sshKey(t, 1, "gh")},
		{ID: 2, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://test.asdf.ru/", SSHKey: loginPasswordKey(2, "bob", "pw")},
		{ID: 3, ProjectID: 1, Type: db.HostConfigURL,
			Name: "https://github.com/acme/", SSHKey: sshKey(t, 3, "acme")},
	}, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	// Only the two ssh mappings hold a key.
	assert.Len(t, installation.Agents, 2)

	cfg := installation.SSHConfigPath()
	assert.Equal(t, installation.Agents[0].SocketFile,
		resolveOption(t, cfg, "github.com", "identityagent"))
	assert.Equal(t, installation.Agents[1].SocketFile,
		resolveOption(t, cfg, "semaphore-mapping-3", "identityagent"))

	params := installation.GitConfigParameters()
	assert.Contains(t, params, "url.https://bob:pw@test.asdf.ru/")
	assert.Contains(t, params, "url.git@semaphore-mapping-3:acme/")
}

// The login of an access key is user supplied and nothing else validates it, so
// a newline in it would inject further directives into the generated config.
func TestInstallHostConfigs_RejectsUnsafeKeyLogin(t *testing.T) {
	setupHostConfig(t)

	key := sshKey(t, 1, "evil")
	key.SshKey.Login = "git\n  ProxyCommand /bin/sh -c id"

	installation, err := InstallHostConfigs(1, []db.HostConfig{{
		ID: 1, ProjectID: 1, Type: db.HostConfigURL,
		Name: "https://github.com/acme/", SSHKey: key,
	}}, task_logger.NopLogger{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "login")
	assert.Nil(t, installation, "nothing may be left installed when generation fails")
}

// A login which is simply unusual must still work.
func TestInstallHostConfigs_AcceptsRealKeyLogins(t *testing.T) {
	setupHostConfig(t)

	for _, login := range []string{"git", "ec2-user", "first.last", "build_agent", "u2"} {
		t.Run(login, func(t *testing.T) {
			key := sshKey(t, 1, "k")
			key.SshKey.Login = login

			installation, err := InstallHostConfigs(1, []db.HostConfig{{
				ID: 1, ProjectID: 1, Type: db.HostConfigURL,
				Name: "https://github.com/acme/", SSHKey: key,
			}}, task_logger.NopLogger{})
			require.NoError(t, err)
			defer installation.Destroy()

			assert.Equal(t, login,
				resolveOption(t, installation.SSHConfigPath(), "semaphore-mapping-1", "user"))
		})
	}
}

// An https-only project must keep the ssh configuration it already has: an
// empty file given to -F would take it away rather than add nothing to it.
func TestInstallHostConfigs_NoConfigFileWithoutSSHMappings(t *testing.T) {
	setupHostConfig(t)

	installation, err := InstallHostConfigs(1, []db.HostConfig{{
		ID: 1, ProjectID: 1, Type: db.HostConfigURL,
		Name: "https://test.asdf.ru/", SSHKey: loginPasswordKey(1, "bob", "s3cr3t"),
	}}, task_logger.NopLogger{})
	require.NoError(t, err)
	defer installation.Destroy()

	assert.Empty(t, installation.SSHConfigPath(), "no ssh mapping, no generated config")
	assert.NotEmpty(t, installation.GitConfigParameters(), "the rewrite is still installed")
}

// The credential is put in the replacement URL, so http would send it in clear.
func TestInstallHostConfigs_RejectsCleartextCredentialURL(t *testing.T) {
	setupHostConfig(t)

	_, err := InstallHostConfigs(1, []db.HostConfig{{
		ID: 1, ProjectID: 1, Type: db.HostConfigURL,
		Name: "http://test.asdf.ru/", SSHKey: loginPasswordKey(1, "bob", "s3cr3t"),
	}}, task_logger.NopLogger{})
	assert.ErrorContains(t, err, "invalid URL")
}

// strict_host_key_checking defaults to "no", whose options once carried the
// executable as well, so the caller's own "ssh " produced "ssh ssh -o ..." and
// git took the second one for the host.
func TestGetGitEnvWithHostConfigs_SingleSSHExecutable(t *testing.T) {
	prev := util.Config
	t.Cleanup(func() { util.Config = prev })

	for _, mode := range []util.SshStrictHostKeyChecking{
		util.SshStrictHostKeyCheckingNo,
		util.SshStrictHostKeyCheckingYes,
		util.SshStrictHostKeyCheckingAcceptNew,
	} {
		t.Run(string(mode), func(t *testing.T) {
			util.Config = &util.ConfigType{Ssh: &util.SshConfig{
				StrictHostKeyChecking: mode,
				KnownHostsFile:        "/tmp/known_hosts",
			}}

			var key AccessKeyInstallation
			env := key.GetGitEnvWithHostConfigs(&HostConfigInstallation{ConfigFile: "/tmp/c.conf"})

			var sshCmd string
			for _, v := range env {
				if strings.HasPrefix(v, "GIT_SSH_COMMAND=") {
					sshCmd = strings.TrimPrefix(v, "GIT_SSH_COMMAND=")
				}
			}

			require.NotEmpty(t, sshCmd)
			assert.Equal(t, "ssh", strings.Fields(sshCmd)[0])
			assert.NotEqual(t, "ssh", strings.Fields(sshCmd)[1], "the executable must appear once")
		})
	}
}

// A host mapping also covers the hosts of an inventory, so the login of the key
// is written out when it has one and left to ssh when it does not.
func TestInstallHostConfigs_HostMappingLogin(t *testing.T) {
	tests := []struct {
		name  string
		login string
		want  string
	}{
		{"key with a login", "deploy", "  User deploy\n"},
		{"key without a login", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupHostConfig(t)

			key := sshKey(t, 1, "k")
			key.SshKey.Login = tt.login

			installation, err := InstallHostConfigs(1, []db.HostConfig{{
				ID: 1, ProjectID: 1, Type: db.HostConfigHost, Name: "example.com", SSHKey: key,
			}}, task_logger.NopLogger{})
			require.NoError(t, err)
			defer installation.Destroy()

			content, err := os.ReadFile(installation.SSHConfigPath())
			require.NoError(t, err)

			if tt.want == "" {
				assert.NotContains(t, string(content), "User ")
			} else {
				assert.Contains(t, string(content), tt.want)
			}
		})
	}
}
