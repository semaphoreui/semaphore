package lib

import (
	"io"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SshAgentKey struct {
	Key        []byte
	Passphrase []byte
}

type SshAgent struct {
	Keys       []SshAgentKey
	Logger     Logger
	listener   net.Listener
	SocketFile string
	done       chan struct{}
}

func (a *SshAgent) Listen() error {

	var err error

	keyring := agent.NewKeyring()

	for _, k := range a.Keys {
		var key interface{}

		if len(k.Passphrase) == 0 {
			key, err = ssh.ParseRawPrivateKey(k.Key)
		} else {
			key, err = ssh.ParseRawPrivateKeyWithPassphrase(k.Key, k.Passphrase)
		}

		if err != nil {
			return errors.Wrap(err, "parsing private key")
		}
		err = keyring.Add(agent.AddedKey{PrivateKey: key})

		if err != nil {
			return err
		}
	}

	l, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: a.SocketFile})
	if err != nil {
		return errors.Wrapf(err, "listening on socket %q", a.SocketFile)
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
