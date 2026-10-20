package mailer

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSMTP accepts one plain SMTP session and records the DATA payload.
type fakeSMTP struct {
	addr string
	mu   sync.Mutex
	data string
	from string
	to   string
}

func newFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return newFakeSMTPOn(t, ln)
}

// newFakeSMTPOn serves one session on an existing listener, so a test can
// wrap the listener in TLS.
func newFakeSMTPOn(t *testing.T, ln net.Listener) *fakeSMTP {
	t.Helper()
	t.Cleanup(func() { _ = ln.Close() })

	f := &fakeSMTP{addr: ln.Addr().String()}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck
		f.serve(conn)
	}()
	return f
}

func (f *fakeSMTP) serve(conn net.Conn) {
	rd := bufio.NewReader(conn)
	write := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
	write("220 fake ESMTP")
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		cmd := strings.ToUpper(line)
		f.mu.Lock()
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			write("250-fake")
			write("250 OK")
		case strings.HasPrefix(cmd, "HELO"):
			write("250 fake")
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			f.from = line[len("MAIL FROM:"):]
			write("250 OK")
		case strings.HasPrefix(cmd, "RCPT TO:"):
			f.to = line[len("RCPT TO:"):]
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
			f.data = b.String()
			write("250 queued")
		case cmd == "QUIT":
			write("221 bye")
			f.mu.Unlock()
			return
		default:
			write("500 what")
		}
		f.mu.Unlock()
	}
}

func (f *fakeSMTP) hostPort(t *testing.T) (string, string) {
	t.Helper()
	host, port, err := net.SplitHostPort(f.addr)
	require.NoError(t, err)
	return host, port
}

func TestSendMail_Anonymous(t *testing.T) {
	srv := newFakeSMTP(t)
	host, port := srv.hostPort(t)

	err := SendMail(context.Background(), Options{
		Host:    host,
		Port:    port,
		From:    "noreply@example.com",
		To:      "ops@example.com",
		Subject: "Task\r\nBcc: x@y.z",
		Body:    "<b>done</b>",
	})
	require.NoError(t, err)

	srv.mu.Lock()
	defer srv.mu.Unlock()
	assert.Equal(t, "<noreply@example.com>", srv.from)
	assert.Equal(t, "<ops@example.com>", srv.to)
	assert.Contains(t, srv.data, "Subject: TaskBcc: x@y.z\r\n")
	assert.Contains(t, srv.data, "<b>done</b>")
}

func TestSendMail_UsesCustomDialer(t *testing.T) {
	srv := newFakeSMTP(t)
	host, port := srv.hostPort(t)

	var dialed string
	err := SendMail(context.Background(), Options{
		Host: host,
		Port: port,
		From: "a@b.c",
		To:   "d@e.f",
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialed = addr
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	})
	require.NoError(t, err)
	assert.Equal(t, srv.addr, dialed)
}

func TestSendMail_DialerErrorAbortsBeforeAnyTraffic(t *testing.T) {
	denied := errors.New("host is not allowed")
	err := SendMail(context.Background(), Options{
		Host: "169.254.169.254",
		Port: "25",
		From: "a@b.c",
		To:   "d@e.f",
		Dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, denied
		},
	})
	assert.ErrorIs(t, err, denied)
}

func TestSendMail_EmptyHost(t *testing.T) {
	err := SendMail(context.Background(), Options{From: "a@b.c", To: "d@e.f"})
	assert.Error(t, err)
}

func TestSendMail_RespectsContextDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() //nolint:errcheck
	// Accept but never send a greeting.
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			defer conn.Close() //nolint:errcheck
			time.Sleep(2 * time.Second)
		}
	}()
	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = SendMail(ctx, Options{Host: host, Port: port, From: "a@b.c", To: "d@e.f"})
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 1500*time.Millisecond)
}

func TestParseTlsVersion(t *testing.T) {
	tests := []struct {
		in      string
		want    uint16
		wantErr bool
	}{
		{"", 0x0303, false},
		{"1.0", 0x0301, false},
		{"1.1", 0x0302, false},
		{"1.2", 0x0303, false},
		{"1.3", 0x0304, false},
		{"2.0", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseTlsVersion(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// selfSignedCert issues a certificate for 127.0.0.1 and the pool that
// trusts it.
func selfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "smtp.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	parsed, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pool
}

func TestSendMail_RootCAs(t *testing.T) {
	cert, pool := selfSignedCert(t)
	listen := func() *fakeSMTP {
		ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
		require.NoError(t, err)
		return newFakeSMTPOn(t, ln)
	}

	t.Run("custom pool verifies the server", func(t *testing.T) {
		srv := listen()
		host, port := srv.hostPort(t)
		err := SendMail(context.Background(), Options{
			Host: host, Port: port, Secure: true, TLS: true, RootCAs: pool,
			From: "a@b.c", To: "d@e.f", Body: "hello",
		})
		require.NoError(t, err)
		srv.mu.Lock()
		defer srv.mu.Unlock()
		assert.Contains(t, srv.data, "hello")
	})

	t.Run("system pool rejects the server", func(t *testing.T) {
		srv := listen()
		host, port := srv.hostPort(t)
		err := SendMail(context.Background(), Options{
			Host: host, Port: port, Secure: true, TLS: true,
			From: "a@b.c", To: "d@e.f",
		})
		assert.ErrorContains(t, err, "x509")
	})
}
