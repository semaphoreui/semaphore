package ssh

import (
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/util"
	"io"
	"net"
	"path"

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

	socketFilename := fmt.Sprintf("ssh-agent-%d-%s.sock", key.ID, random.String(10))

	var socketFile string

	if key.ProjectID == nil {
		socketFile = path.Join(util.Config.TmpPath, socketFilename)
	} else {
		socketFile = path.Join(util.Config.GetProjectTmpDir(*key.ProjectID), socketFilename)
	}

	sshAgent := Agent{
		Logger: logger,
		Keys: []AgentKey{
			{
				Key:        []byte(key.SshKey.PrivateKey),
				Passphrase: []byte(key.SshKey.Passphrase),
			},
		},
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

func (key *AccessKeyInstallation) GetGitEnv() (env []string) {
	env = make([]string, 0)

	env = append(env, fmt.Sprintln("GIT_TERMINAL_PROMPT=0"))
	if key.SSHAgent != nil {
		env = append(env, fmt.Sprintf("SSH_AUTH_SOCK=%s", key.SSHAgent.SocketFile))
		sshCmd := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
		if util.Config.SshConfigPath != "" {
			sshCmd += " -F " + util.Config.SshConfigPath
		}
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
	}

	return env
}

func (key *AccessKeyInstallation) Destroy() error {
	if key.SSHAgent != nil {
		return key.SSHAgent.Close()
	}
	return nil
}

type KeyInstaller struct{}

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
