package alerting

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSMTPServer answers one minimal SMTP session on the loopback interface
// and reports whether anybody connected.
type fakeSMTPServer struct {
	host, port string
	connected  chan struct{}
	data       chan string
}

func newFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)

	f := &fakeSMTPServer{host: host, port: port, connected: make(chan struct{}, 1), data: make(chan string, 1)}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		f.connected <- struct{}{}
		defer conn.Close() //nolint:errcheck

		rd := bufio.NewReader(conn)
		write := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		write("220 fake ESMTP")
		for {
			line, err := rd.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimRight(line, "\r\n"))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				write("250-fake")
				write("250 OK")
			case strings.HasPrefix(cmd, "HELO"), strings.HasPrefix(cmd, "MAIL FROM:"), strings.HasPrefix(cmd, "RCPT TO:"):
				write("250 OK")
			case cmd == "DATA":
				write("354 go")
				var b strings.Builder
				for {
					l, err := rd.ReadString('\n')
					if err != nil || l == ".\r\n" {
						break
					}
					b.WriteString(l)
				}
				f.data <- b.String()
				write("250 queued")
			case cmd == "QUIT":
				write("221 bye")
				return
			default:
				write("500 what")
			}
		}
	}()
	return f
}

func (f *fakeSMTPServer) wasConnected() bool {
	select {
	case <-f.connected:
		return true
	case <-time.After(300 * time.Millisecond):
		return false
	}
}

func emailChannelForTest(t *testing.T) *emailChannel {
	t.Helper()
	ch, err := NewRegistry().Get(TypeEmail)
	require.NoError(t, err)
	email, ok := ch.(*emailChannel)
	require.True(t, ok)
	return email
}

func TestEmailChannel_Validate_OwnHostPolicy(t *testing.T) {
	email := emailChannelForTest(t)
	cfg := &util.ConfigType{EmailHost: "localhost", EmailPort: "25", EmailSender: "srv@example.com"}

	tests := []struct {
		name    string
		dest    Destination
		wantErr string
	}{
		{"own public host", Destination{Params: map[string]any{"smtp_host": "smtp.example.com", "smtp_sender": "a@b.c"}}, ""},
		{"own private host", Destination{Params: map[string]any{"smtp_host": "10.1.2.3", "smtp_sender": "a@b.c"}}, ""},
		{"own loopback", Destination{Params: map[string]any{"smtp_host": "127.0.0.1", "smtp_sender": "a@b.c"}}, "smtp_host is not allowed"},
		{"own localhost", Destination{Params: map[string]any{"smtp_host": "localhost", "smtp_sender": "a@b.c"}}, "smtp_host is not allowed"},
		{"own metadata", Destination{Params: map[string]any{"smtp_host": "169.254.169.254", "smtp_port": "80", "smtp_sender": "a@b.c"}}, "smtp_host is not allowed"},
		{"own host with port", Destination{Params: map[string]any{"smtp_host": "smtp.example.com:25", "smtp_sender": "a@b.c"}}, "smtp_host is not allowed"},
		{"own bad port", Destination{Params: map[string]any{"smtp_host": "smtp.example.com", "smtp_port": "99999", "smtp_sender": "a@b.c"}}, "smtp_port"},
		{"own zero port", Destination{Params: map[string]any{"smtp_host": "smtp.example.com", "smtp_port": "0", "smtp_sender": "a@b.c"}}, "smtp_port"},
		// The server-wide host is the admin's choice and may be local.
		{"server host on loopback", Destination{Recipients: []string{"a@b.c"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := email.Validate(cfg, tt.dest)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestEmailChannel_Send_OwnLoopbackHostIsNeverDialed(t *testing.T) {
	email := emailChannelForTest(t)
	srv := newFakeSMTPServer(t)

	dest := Destination{
		Recipients: []string{"ops@example.com"},
		Params:     map[string]any{"smtp_host": srv.host, "smtp_port": srv.port, "smtp_sender": "a@b.c"},
	}
	err := email.Send(context.Background(), &util.ConfigType{}, dest, Message{Subject: "s", Body: "b"})
	assert.ErrorContains(t, err, "smtp_host is not allowed")
	assert.False(t, srv.wasConnected(), "a project SMTP override must never reach the loopback interface")
}

func TestEmailChannel_Send_ServerHostOnLoopbackWorks(t *testing.T) {
	email := emailChannelForTest(t)
	srv := newFakeSMTPServer(t)
	cfg := &util.ConfigType{EmailHost: srv.host, EmailPort: srv.port, EmailSender: "srv@example.com"}

	dest := Destination{Recipients: []string{"ops@example.com"}}
	err := email.Send(context.Background(), cfg, dest, Message{Subject: "Task done", Body: "<p>ok</p>"})
	require.NoError(t, err)
	assert.True(t, srv.wasConnected())

	select {
	case data := <-srv.data:
		assert.Contains(t, data, "Subject: Task done")
		assert.Contains(t, data, "<p>ok</p>")
	case <-time.After(time.Second):
		t.Fatal("no message received")
	}
}

func TestEmailChannel_Send_TrustedDestinationBypassesPolicy(t *testing.T) {
	email := emailChannelForTest(t)
	srv := newFakeSMTPServer(t)

	dest := Destination{
		Trusted:    true,
		Recipients: []string{"ops@example.com"},
		Params:     map[string]any{"smtp_host": srv.host, "smtp_port": srv.port, "smtp_sender": "a@b.c"},
	}
	err := email.Send(context.Background(), &util.ConfigType{}, dest, Message{Subject: "s", Body: "b"})
	require.NoError(t, err)
	assert.True(t, srv.wasConnected())
}

func TestEmailChannel_Validate_OwnHostKeepsServerDefaults(t *testing.T) {
	email := emailChannelForTest(t)
	cfg := &util.ConfigType{
		EmailHost:   "smtp.server",
		EmailPort:   "587",
		EmailSender: "srv@example.com",
		EmailSecure: true,
		EmailTls:    true,
	}

	settings, err := email.settings(cfg, Destination{
		Params: map[string]any{"smtp_host": "smtp.project.example"},
	})
	require.NoError(t, err)
	assert.Equal(t, "smtp.project.example", settings.host)
	assert.Equal(t, "587", settings.port)
	assert.Equal(t, "srv@example.com", settings.sender)
	assert.True(t, settings.secure)
	assert.True(t, settings.tls)
}

func TestEmailChannel_Send_NoRecipients(t *testing.T) {
	email := emailChannelForTest(t)
	cfg := &util.ConfigType{EmailHost: "smtp.example.com", EmailSender: "srv@example.com"}
	err := email.Send(context.Background(), cfg, Destination{}, Message{})
	assert.ErrorContains(t, err, "no e-mail recipients")
}
