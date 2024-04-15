package ssh

import (
	"fmt"
	"io"
	"net"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
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
			key interface{}
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
				defer conn.Close()

				if err := agent.ServeAgent(keyring, conn); err != nil && err != io.EOF {
					a.Logger.Logf("error serving SSH agent listener: %w", err)
				}
			}(conn)
		}
	}()

	return nil
}

func (a *Agent) Close() error {
	close(a.done)
	return a.listener.Close()
}
