//go:build integration

package integration

import (
	"context"
	"errors"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/semaphoreui/semaphore/util/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	authUser = "user"
	authPass = "pass"
)

func send(t *testing.T, opts mailer.Options) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mailer.SendMail(ctx, opts)
}

// TestSendMail_PlainNoAuth: a relay on port 25 without authentication or
// TLS, the email_secure=false configuration.
func TestSendMail_PlainNoAuth(t *testing.T) {
	m := startMailpit(t, mailpitConfig{})

	t.Run("A1 anonymous", func(t *testing.T) {
		m.clear()
		require.NoError(t, send(t, m.options("localhost")))
		msg := m.waitForOne()
		assert.Equal(t, "noreply@example.com", msg.From.Address)
		assert.Equal(t, "ops@example.com", msg.To[0].Address)
		assert.Equal(t, "Task finished", msg.Subject)
		assert.Empty(t, msg.Username)
	})

	t.Run("A2 secure with credentials the server never asks for", func(t *testing.T) {
		m.clear()
		opts := m.options("smtp.test")
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		require.NoError(t, send(t, opts))
		assert.Empty(t, m.waitForOne().Username, "AUTH must not be sent when the server does not offer it")
	})

	t.Run("C1 line breaks in the subject can not inject headers", func(t *testing.T) {
		m.clear()
		opts := m.options("localhost")
		opts.Subject = "Task\r\nBcc: spy@example.com"
		require.NoError(t, send(t, opts))
		msg := m.waitForOne()
		raw := m.raw(msg.ID)
		assert.Contains(t, raw, "Subject: TaskBcc: spy@example.com\r\n")
		assert.NotContains(t, raw, "\nBcc:")
		assert.Equal(t, "TaskBcc: spy@example.com", msg.Subject)
	})

	t.Run("C2 line breaks in envelope addresses are stripped", func(t *testing.T) {
		m.clear()
		opts := m.options("localhost")
		opts.From = "\r\nsender@example.com"
		opts.To = "rcpt@example.com\n"
		require.NoError(t, send(t, opts))
		msg := m.waitForOne()
		assert.Equal(t, "sender@example.com", msg.From.Address)
		assert.Equal(t, "rcpt@example.com", msg.To[0].Address)
		raw := m.raw(msg.ID)
		assert.Contains(t, raw, "From: sender@example.com\r\n")
		assert.Contains(t, raw, "To: rcpt@example.com\r\n")
	})

	t.Run("C3 html body with utf-8", func(t *testing.T) {
		m.clear()
		opts := m.options("localhost")
		opts.Subject = "Задача выполнена ✓"
		opts.Body = "<p>Задача <b>выполнена</b> ✓</p>"
		require.NoError(t, send(t, opts))
		raw := m.raw(m.waitForOne().ID)
		assert.Contains(t, raw, "Content-Type: text/html; charset=UTF-8")
		assert.Contains(t, raw, "<p>Задача <b>выполнена</b> ✓</p>")
	})

	t.Run("C4 date header", func(t *testing.T) {
		m.clear()
		require.NoError(t, send(t, m.options("localhost")))
		raw := m.raw(m.waitForOne().ID)
		parsed, err := mail.ReadMessage(strings.NewReader(raw))
		require.NoError(t, err)
		date := parsed.Header.Get("Date")
		require.NotEmpty(t, date)
		sent, err := time.Parse(time.RFC1123, date)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), sent, 5*time.Minute)
	})

	t.Run("D2 cancelled context", func(t *testing.T) {
		m.clear()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := mailer.SendMail(ctx, m.options("localhost"))
		assert.ErrorIs(t, err, context.Canceled)
		m.assertNoMessages()
	})
}

// TestSendMail_InsecureAuth: AUTH advertised on a plain-text connection.
// The client accepts that for localhost only.
func TestSendMail_InsecureAuth(t *testing.T) {
	m := startMailpit(t, mailpitConfig{auth: authUser + ":" + authPass, allowInsecureAuth: true})

	t.Run("A3 plain-text auth to localhost", func(t *testing.T) {
		m.clear()
		opts := m.options("localhost")
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		require.NoError(t, send(t, opts))
		assert.Equal(t, authUser, m.waitForOne().Username)
	})

	t.Run("A4 plain-text auth to a remote host is refused", func(t *testing.T) {
		m.clear()
		opts := m.options("smtp.test")
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		err := send(t, opts)
		assert.ErrorContains(t, err, "unencrypted connection")
		m.assertNoMessages()
	})
}

// TestSendMail_StartTLS: email_secure=true, email_tls=false. AUTH is only
// offered after the upgrade, like real servers do.
func TestSendMail_StartTLS(t *testing.T) {
	certs := newTestCerts(t)
	m := startMailpit(t, mailpitConfig{auth: authUser + ":" + authPass, tls: &certs.server})

	secure := func(host string) mailer.Options {
		opts := m.options(host)
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		opts.RootCAs = certs.pool
		return opts
	}

	t.Run("A5 starttls then auth", func(t *testing.T) {
		m.clear()
		require.NoError(t, send(t, secure("smtp.test")))
		assert.Equal(t, authUser, m.waitForOne().Username,
			"AUTH is offered after STARTTLS only, so a username proves the upgrade")
	})

	t.Run("A9 implicit tls against a starttls server fails fast", func(t *testing.T) {
		m.clear()
		opts := secure("smtp.test")
		opts.TLS = true
		start := time.Now()
		err := send(t, opts)
		assert.Error(t, err)
		assert.Less(t, time.Since(start), 5*time.Second)
		m.assertNoMessages()
	})

	t.Run("A10 wrong password", func(t *testing.T) {
		m.clear()
		opts := secure("smtp.test")
		opts.Password = "nope"
		err := send(t, opts)
		assert.ErrorContains(t, err, "535")
		m.assertNoMessages()
	})

	t.Run("A11 empty username", func(t *testing.T) {
		m.clear()
		opts := secure("smtp.test")
		opts.Username = ""
		err := send(t, opts)
		assert.Error(t, err)
		m.assertNoMessages()
	})

	t.Run("B1 system roots do not trust the test CA", func(t *testing.T) {
		m.clear()
		opts := secure("smtp.test")
		opts.RootCAs = nil
		err := send(t, opts)
		assert.ErrorContains(t, err, "x509")
		m.assertNoMessages()
	})

	t.Run("B3 certificate name mismatch", func(t *testing.T) {
		m.clear()
		err := send(t, secure("other.test"))
		assert.ErrorContains(t, err, "x509")
		assert.ErrorContains(t, err, "other.test")
		m.assertNoMessages()
	})
}

// TestSendMail_RequireStartTLS: a server that refuses plain-text sessions.
func TestSendMail_RequireStartTLS(t *testing.T) {
	certs := newTestCerts(t)
	m := startMailpit(t, mailpitConfig{auth: authUser + ":" + authPass, tls: &certs.server, requireStartTLS: true})

	t.Run("A6 insecure client is rejected", func(t *testing.T) {
		m.clear()
		err := send(t, m.options("smtp.test"))
		assert.ErrorContains(t, err, "STARTTLS")
		m.assertNoMessages()
	})

	t.Run("secure client upgrades and delivers", func(t *testing.T) {
		m.clear()
		opts := m.options("smtp.test")
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		opts.RootCAs = certs.pool
		require.NoError(t, send(t, opts))
		assert.Equal(t, authUser, m.waitForOne().Username)
	})
}

// TestSendMail_ImplicitTLS: email_secure=true, email_tls=true.
func TestSendMail_ImplicitTLS(t *testing.T) {
	certs := newTestCerts(t)
	m := startMailpit(t, mailpitConfig{auth: authUser + ":" + authPass, tls: &certs.server, implicitTLS: true})

	implicit := func(host string) mailer.Options {
		opts := m.options(host)
		opts.Secure, opts.TLS = true, true
		opts.Username, opts.Password = authUser, authPass
		opts.RootCAs = certs.pool
		return opts
	}

	t.Run("A7 tls from the first byte then auth", func(t *testing.T) {
		m.clear()
		require.NoError(t, send(t, implicit("smtp.test")))
		assert.Equal(t, authUser, m.waitForOne().Username)
	})

	t.Run("A8 starttls client against an implicit tls server times out, not hangs", func(t *testing.T) {
		m.clear()
		opts := implicit("smtp.test")
		opts.TLS = false
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		start := time.Now()
		err := mailer.SendMail(ctx, opts)
		assert.Error(t, err)
		assert.Less(t, time.Since(start), 8*time.Second)
		m.assertNoMessages()
	})

	t.Run("B4 tls 1.3 minimum", func(t *testing.T) {
		m.clear()
		opts := implicit("smtp.test")
		opts.TLSMinVersion = "1.3"
		require.NoError(t, send(t, opts))
		m.waitForOne()
	})
}

// TestSendMail_UntrustedIssuer: the server presents a certificate from a CA
// the client does not know.
func TestSendMail_UntrustedIssuer(t *testing.T) {
	certs := newTestCerts(t)
	m := startMailpit(t, mailpitConfig{auth: authUser + ":" + authPass, tls: &certs.untrusted})

	t.Run("B2 starttls", func(t *testing.T) {
		m.clear()
		opts := m.options("smtp.test")
		opts.Secure = true
		opts.Username, opts.Password = authUser, authPass
		opts.RootCAs = certs.pool
		err := send(t, opts)
		assert.ErrorContains(t, err, "x509")
		m.assertNoMessages()
	})
}

// TestSendMail_UnsupportedTLSVersionFailsBeforeDial (B5) needs no server:
// the option is rejected before any connection is opened.
func TestSendMail_UnsupportedTLSVersionFailsBeforeDial(t *testing.T) {
	dialed := false
	err := mailer.SendMail(context.Background(), mailer.Options{
		Host:          "smtp.test",
		Port:          "25",
		From:          "a@b.c",
		To:            "d@e.f",
		TLSMinVersion: "9.9",
		Dial: func(context.Context, string, string) (net.Conn, error) {
			dialed = true
			return nil, errors.New("must not be called")
		},
	})
	assert.ErrorContains(t, err, "unsupported TLS version")
	assert.False(t, dialed)
}

// TestSendMail_ServerNeverGreets (D1): an accepted connection with no 220
// line must not hold the caller past the deadline.
func TestSendMail_ServerNeverGreets(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() //nolint:errcheck
	go func() {
		if conn, err := ln.Accept(); err == nil {
			defer conn.Close() //nolint:errcheck
			time.Sleep(5 * time.Second)
		}
	}()
	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	start := time.Now()
	err = mailer.SendMail(ctx, mailer.Options{Host: host, Port: port, From: "a@b.c", To: "d@e.f"})
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 3*time.Second)
}

// TestSendMail_PausedServer (D3): the daemon is frozen with docker pause, so
// the TCP handshake may still complete but nothing is ever answered.
func TestSendMail_PausedServer(t *testing.T) {
	m := startMailpit(t, mailpitConfig{})
	ctx := context.Background()

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer cli.Close() //nolint:errcheck

	id := m.container.GetContainerID()
	_, err = cli.ContainerPause(ctx, id, client.ContainerPauseOptions{})
	require.NoError(t, err)
	defer func() { _, _ = cli.ContainerUnpause(ctx, id, client.ContainerUnpauseOptions{}) }()

	sendCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	start := time.Now()
	err = mailer.SendMail(sendCtx, m.options("localhost"))
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 8*time.Second)
}
