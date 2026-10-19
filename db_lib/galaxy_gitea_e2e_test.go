package db_lib

import (
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

// These tests need a git server that really refuses anonymous access, because a
// public host proves nothing: the clone succeeds with or without the credential.
// They skip unless one answers on giteaAddr with:
//
//	semuser / sempass123
//	a private repository semuser/private-role, holding an ansible role
//
// A loopback address is deliberate. A developer's ~/.gitconfig commonly rewrites
// or supplies credentials for github.com, which silently passes the test for the
// wrong reason; nothing matches 127.0.0.1.
//
//	docker run -d --name gitea -p 127.0.0.1:3300:3000 //	  -e GITEA__security__INSTALL_LOCK=true gitea/gitea:1.24
const giteaAddr = "127.0.0.1:3300"

func requireGitea(t *testing.T) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", giteaAddr, 2*time.Second)
	if err != nil {
		t.Skip("gitea fixture is not running, see sem122-gitea/docker-compose.yml")
	}
	_ = conn.Close()

	if _, err := exec.LookPath("ansible-galaxy"); err != nil {
		t.Skip("ansible-galaxy is not available")
	}
}

// runGalaxyInstall runs the real ansible-galaxy against the private role with the
// given environment, on top of the ambient one exactly as a task does. The
// fixture host is what keeps the developer's own credentials out of it: no
// rewrite rule and no credential helper in ~/.gitconfig matches 127.0.0.1.
func runGalaxyInstall(t *testing.T, env []string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	requirements := filepath.Join(dir, "requirements.yml")
	require.NoError(t, os.WriteFile(requirements, []byte(
		"---\nroles:\n  - name: private_role\n    src: http://"+giteaAddr+
			"/semuser/private-role.git\n    scm: git\n    version: main\n"), 0600))

	cmd := exec.Command("ansible-galaxy", "role", "install",
		"-r", requirements, "--force", "-p", filepath.Join(dir, "roles"))
	cmd.Env = append(os.Environ(), env...)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func giteaRepository() db.Repository {
	return db.Repository{
		GitURL: "http://" + giteaAddr + "/semuser/main-repo.git",
		SSHKey: db.AccessKey{
			Type:          db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{Login: "semuser", Password: "sempass123"},
		},
	}
}

// mappingEnv is what the task pipeline hands the galaxy step once the project has
// credential mappings: its own GIT_CONFIG_PARAMETERS, for an unrelated host.
func mappingEnv(t *testing.T) []string {
	t.Helper()

	installation, err := ssh.InstallHostConfigs(1, []db.HostConfig{{
		ID: 1, ProjectID: 1, Type: db.HostConfigURL,
		Name: "https://github.com/acme/",
		SSHKey: db.AccessKey{
			ID: 2, Type: db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{Login: "alice", Password: "t0ken"},
		},
	}}, task_logger.NopLogger{})
	require.NoError(t, err)
	t.Cleanup(installation.Destroy)

	var noKey ssh.AccessKeyInstallation
	return noKey.GetGitEnvWithHostConfigs(installation)
}

// The whole point of SEM-122: ansible-galaxy authenticates to the private server
// with the credential of the repository, even though a mapping for an unrelated
// host also wants GIT_CONFIG_PARAMETERS.
func TestGalaxyInstallsPrivateRoleAlongsideHostConfigs(t *testing.T) {
	requireGitea(t)
	setupGalaxyConfig(t)
	util.Config.Ssh = &util.SshConfig{StrictHostKeyChecking: util.SshStrictHostKeyCheckingNo}

	app := &AnsibleApp{Repository: giteaRepository()}

	env, err := app.galaxyEnv(mappingEnv(t))
	require.NoError(t, err)

	out, err := runGalaxyInstall(t, env)

	require.NoError(t, err, out)
	assert.Contains(t, out, "private_role")
}

// The same run with the two sets of rewrites simply concatenated, which is what
// happens without mergeGitConfigParameters: git keeps the last variable only and
// the credential of the repository is gone.
func TestGalaxyWithoutMergeLosesTheRepositoryCredential(t *testing.T) {
	requireGitea(t)
	setupGalaxyConfig(t)
	util.Config.Ssh = &util.SshConfig{StrictHostKeyChecking: util.SshStrictHostKeyCheckingNo}

	app := &AnsibleApp{Repository: giteaRepository()}

	gitEnv, err := app.galaxyGitEnvForRun()
	require.NoError(t, err)

	out, err := runGalaxyInstall(t, append(gitEnv, mappingEnv(t)...))

	require.Error(t, err, out)
	assert.Contains(t, out, "could not read Username")
}
