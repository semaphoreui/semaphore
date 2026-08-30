package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type AgentKey struct {
	Key        []byte
	Passphrase []byte
}

type Agent struct {
	Keys       []AgentKey
	Logger     task_logger.Logger
	listener   net.Listener
	SocketFile string
	done       chan struct{}
}

func NewAgent() Agent {
	return Agent{}
}

func (a *Agent) Listen() error {
	keyring := agent.NewKeyring()

	for _, k := range a.Keys {
		var (
			key any
			err error
		)

		if len(k.Passphrase) == 0 {
			key, err = ssh.ParseRawPrivateKey(k.Key)
		} else {
			key, err = ssh.ParseRawPrivateKeyWithPassphrase(k.Key, k.Passphrase)
		}

		if err != nil {
			return fmt.Errorf("parsing private key: %w", err)
		}

		if err := keyring.Add(agent.AddedKey{
			PrivateKey: key,
		}); err != nil {
			return fmt.Errorf("adding private key: %w", err)
		}
	}

	if err := os.MkdirAll(path.Dir(a.SocketFile), 0o755); err != nil {
		return fmt.Errorf("creating socket directory: %w", err)
	}

	l, err := net.ListenUnix(
		"unix",
		&net.UnixAddr{
			Net:  "unix",
			Name: a.SocketFile,
		},
	)
	if err != nil {
		return fmt.Errorf("listening on socket %q: %w", a.SocketFile, err)
	}

	l.SetUnlinkOnClose(true)
	a.listener = l
	a.done = make(chan struct{})

	go func() {
		for {
			conn, err := a.listener.Accept()
			if err != nil {
				select {
				case <-a.done:
					return
				default:
					a.Logger.Logf("error accepting socket connection: %w", err)
					return
				}
			}

			go func(conn net.Conn) {
				defer conn.Close() //nolint:errcheck

				if err := agent.ServeAgent(keyring, conn); err != nil && err != io.EOF {
					a.Logger.Logf("error serving SSH agent listener: %w", err)
				}
			}(conn)
		}
	}()

	return nil
}

func (a *Agent) Close() error {
	if a.done != nil {
		close(a.done)
	}
	if a.listener != nil {
		return a.listener.Close()
	}
	return nil
}

func StartSSHAgent(key db.AccessKey, logger task_logger.Logger) (Agent, error) {
	return StartSSHAgentWithKeys([]db.AccessKey{key}, logger)
}

// StartSSHAgentWithKeys starts an agent holding several keys, for a connection
// which authenticates against more than one host, such as a chain of proxies.
// ssh offers the keys in order and moves on to the next when a host rejects one.
func StartSSHAgentWithKeys(keys []db.AccessKey, logger task_logger.Logger) (Agent, error) {
	if len(keys) == 0 {
		return Agent{}, fmt.Errorf("no keys to start an ssh agent with")
	}

	first := keys[0]
	socketFilename := fmt.Sprintf("ssh-agent-%d-%s.sock", first.ID, random.String(10))

	var socketFile string

	if first.ProjectID == nil {
		socketFile = path.Join(util.Config.TmpPath, socketFilename)
	} else {
		socketFile = path.Join(util.Config.GetProjectTmpDir(*first.ProjectID), socketFilename)
	}

	agentKeys := make([]AgentKey, 0, len(keys))
	for _, key := range keys {
		agentKeys = append(agentKeys, AgentKey{
			Key:        []byte(key.SshKey.PrivateKey),
			Passphrase: []byte(key.SshKey.Passphrase),
		})
	}

	sshAgent := Agent{
		Logger:     logger,
		Keys:       agentKeys,
		SocketFile: socketFile,
	}

	return sshAgent, sshAgent.Listen()
}

type AccessKeyInstallation struct {
	SSHAgent *Agent
	Login    string
	Password string
	Script   string
}

// GetGitEnv returns the environment for git commands. Additional ssh options,
// for example the jump host of a proxy, are appended to GIT_SSH_COMMAND.
func (key *AccessKeyInstallation) GetGitEnv(sshOpts ...string) (env []string) {
	env = make([]string, 0)

	env = append(env, "GIT_TERMINAL_PROMPT=0")

	if key.SSHAgent != nil {
		env = append(env, fmt.Sprintf("SSH_AUTH_SOCK=%s", key.SSHAgent.SocketFile))
	}

	// GIT_SSH_COMMAND is also needed without an agent, otherwise a repository
	// which needs no key of its own loses the options of its proxy.
	if key.SSHAgent != nil || len(sshOpts) > 0 {
		sshCmd := "ssh " + gitHostKeyCheckingOpts()
		if util.Config.GetSshConfigPath() != "" {
			sshCmd += " -F " + util.Config.GetSshConfigPath()
		}
		for _, opt := range sshOpts {
			sshCmd += " " + opt
		}
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
	}

	return env
}

// gitHostKeyCheckingOpts returns the ssh host-key verification options used for
// git operations. Host-key checking is enabled so a network attacker cannot
// impersonate the git server. When an explicit known_hosts file is configured
// it is used with strict checking; otherwise a persistent trust-on-first-use
// file under TmpPath is used (accept-new): the first host key seen is pinned and
// any subsequent change is rejected.
func gitHostKeyCheckingOpts() string {
	switch util.Config.Ssh.StrictHostKeyChecking {
	case util.SshStrictHostKeyCheckingYes:
		return fmt.Sprintf("-o StrictHostKeyChecking=yes -o UserKnownHostsFile=%s", util.Config.Ssh.KnownHostsFile)
	case util.SshStrictHostKeyCheckingNo:
		// No leading "ssh": the caller prepends it, and a second one is taken by
		// ssh as the host to connect to.
		return "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
	case util.SshStrictHostKeyCheckingAcceptNew:
		return fmt.Sprintf("-o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=%s", util.Config.Ssh.KnownHostsFile)
	default:
		panic("Unknown SSH strict host key check option")
	}
}

func (key *AccessKeyInstallation) Destroy() error {
	if key.SSHAgent != nil {
		return key.SSHAgent.Close()
	}
	return nil
}

type KeyInstaller struct{}

// InstallAll installs several keys into a single agent. Only ssh keys can share
// an agent, so the usages which do not produce one are rejected.
func (i KeyInstaller) InstallAll(keys []db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation AccessKeyInstallation, err error) {
	if len(keys) == 1 {
		return i.Install(keys[0], usage, logger)
	}

	for _, key := range keys {
		if key.Type != db.AccessKeySSH {
			err = fmt.Errorf("only ssh keys can share an agent")
			return
		}
	}

	var agent Agent
	agent, err = StartSSHAgentWithKeys(keys, logger)
	installation.SSHAgent = &agent
	installation.Login = keys[0].SshKey.Login
	return
}

func (KeyInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation AccessKeyInstallation, err error) {

	switch usage {
	case db.AccessKeyRoleGit:
		switch key.Type {
		case db.AccessKeySSH:
			var agent Agent
			agent, err = StartSSHAgent(key, logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		}
	case db.AccessKeyRoleAnsiblePasswordVault:
		switch key.Type {
		case db.AccessKeyLoginPassword:
			installation.Password = key.LoginPassword.Password
		default:
			err = fmt.Errorf("access key type not supported for ansible password vault")
		}
	case db.AccessKeyRoleAnsibleBecomeUser:
		if key.Type != db.AccessKeyLoginPassword {
			err = fmt.Errorf("access key type not supported for ansible become user")
		}
		installation.Login = key.LoginPassword.Login
		installation.Password = key.LoginPassword.Password
	case db.AccessKeyRoleAnsibleUser:
		switch key.Type {
		case db.AccessKeySSH:
			var agent Agent
			agent, err = StartSSHAgent(key, logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		case db.AccessKeyLoginPassword:
			installation.Login = key.LoginPassword.Login
			installation.Password = key.LoginPassword.Password
		case db.AccessKeyNone:
			// No SSH agent or password needed for ansible user with no access key.
		default:
			err = fmt.Errorf("access key type not supported for ansible user")
		}
	}

	return
}
