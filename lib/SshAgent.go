package lib

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"net"
)

type SshAgent struct {
	Key        []byte
	Passphrase []byte
	Logger     Logger
	listener   net.Listener
	done       chan struct{}
}

func (a *SshAgent) Listen(sock string) error {
	key, err := ssh.ParseRawPrivateKeyWithPassphrase(a.Key, a.Passphrase)
	if err != nil {
		return errors.Wrap(err, "parsing private key")
	}
	keyring := agent.NewKeyring()
	err = keyring.Add(agent.AddedKey{
		PrivateKey: key,
	})
	if err != nil {
		return err
	}
	l, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: sock})
	if err != nil {
		return errors.Wrapf(err, "listening on socket %q", sock)
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
					a.Logger.Log("error accepting socket connection")
					return
				}
			}

			go func(conn net.Conn) {
				defer conn.Close()

				if err := agent.ServeAgent(keyring, conn); err != nil && err != io.EOF {
					a.Logger.Log("error serving SSH agent")
				}
			}(conn)
		}
	}()
	return nil
}

func (a *SshAgent) Close() error {
	close(a.done)
	return a.listener.Close()
}
