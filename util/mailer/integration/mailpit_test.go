//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/util/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mailpitImage = "axllent/mailpit:v1.28"
	smtpPort     = "1025/tcp"
	apiPort      = "8025/tcp"
)

// mailpitConfig selects how the SMTP daemon is configured. Every field maps
// to one Mailpit environment variable.
type mailpitConfig struct {
	// auth is "user:pass"; empty disables AUTH altogether.
	auth string
	// allowInsecureAuth advertises AUTH on a plain-text connection too.
	// Without it Mailpit offers AUTH only after TLS, like real servers.
	allowInsecureAuth bool
	// tls enables STARTTLS with this certificate.
	tls *certFiles
	// requireStartTLS rejects MAIL FROM until the client upgraded.
	requireStartTLS bool
	// implicitTLS serves TLS from the first byte (port 465 style).
	implicitTLS bool
	// allowedRecipients is a regular expression; empty allows every RCPT.
	allowedRecipients string
}

// mailpit is one running container with its SMTP and API endpoints.
type mailpit struct {
	t         *testing.T
	container testcontainers.Container
	// host is the Docker host the mapped ports listen on ("localhost").
	host     string
	smtpPort string
	apiURL   string
	// ip is the container address on the Docker bridge network. Routable
	// from the test process on Linux only.
	ip string
}

func startMailpit(t *testing.T, cfg mailpitConfig) *mailpit {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	env := map[string]string{"MP_SMTP_DISABLE_RDNS": "true"}
	if cfg.auth != "" {
		env["MP_SMTP_AUTH"] = cfg.auth
	}
	if cfg.allowInsecureAuth {
		env["MP_SMTP_AUTH_ALLOW_INSECURE"] = "true"
	}
	if cfg.requireStartTLS {
		env["MP_SMTP_REQUIRE_STARTTLS"] = "true"
	}
	if cfg.implicitTLS {
		env["MP_SMTP_REQUIRE_TLS"] = "true"
	}
	if cfg.allowedRecipients != "" {
		env["MP_SMTP_ALLOWED_RECIPIENTS"] = cfg.allowedRecipients
	}

	var files []testcontainers.ContainerFile
	if cfg.tls != nil {
		env["MP_SMTP_TLS_CERT"] = "/certs/server.crt"
		env["MP_SMTP_TLS_KEY"] = "/certs/server.key"
		files = []testcontainers.ContainerFile{
			{HostFilePath: cfg.tls.certPath, ContainerFilePath: "/certs/server.crt", FileMode: 0o644},
			{HostFilePath: cfg.tls.keyPath, ContainerFilePath: "/certs/server.key", FileMode: 0o644},
		}
	}

	req := testcontainers.ContainerRequest{
		Image:        mailpitImage,
		ExposedPorts: []string{smtpPort, apiPort},
		Env:          env,
		Files:        files,
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(smtpPort),
			wait.ForHTTP("/api/v1/info").WithPort(apiPort),
		),
	}
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	smtp, err := ctr.MappedPort(ctx, smtpPort)
	require.NoError(t, err)
	api, err := ctr.MappedPort(ctx, apiPort)
	require.NoError(t, err)
	ip, err := ctr.ContainerIP(ctx)
	require.NoError(t, err)

	return &mailpit{
		t:         t,
		container: ctr,
		host:      host,
		smtpPort:  smtp.Port(),
		apiURL:    "http://" + net.JoinHostPort(host, api.Port()),
		ip:        ip,
	}
}

// dial routes any address to the container's mapped SMTP port, so a test
// can set Options.Host freely (certificate name, localhost exception)
// without DNS.
func (m *mailpit) dial(ctx context.Context, network, _ string) (net.Conn, error) {
	return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(m.host, m.smtpPort))
}

// options is a valid message for this server; tests override what they
// test.
func (m *mailpit) options(host string) mailer.Options {
	return mailer.Options{
		Host:    host,
		Port:    "25",
		From:    "noreply@example.com",
		To:      "ops@example.com",
		Subject: "Task finished",
		Body:    "<p>ok</p>",
		Dial:    m.dial,
	}
}

type mailpitAddress struct {
	Name    string
	Address string
}

type mailpitMessage struct {
	ID       string
	From     mailpitAddress
	To       []mailpitAddress
	Subject  string
	Username string
}

func (m *mailpit) api(method, path string, body string) []byte {
	m.t.Helper()
	req, err := http.NewRequest(method, m.apiURL+path, strings.NewReader(body))
	require.NoError(m.t, err)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(m.t, err)
	defer resp.Body.Close() //nolint:errcheck
	data, err := io.ReadAll(resp.Body)
	require.NoError(m.t, err)
	require.Equal(m.t, http.StatusOK, resp.StatusCode, "%s %s: %s", method, path, data)
	return data
}

// clear deletes every stored message.
func (m *mailpit) clear() {
	m.t.Helper()
	m.api(http.MethodDelete, "/api/v1/messages", "{}")
}

func (m *mailpit) messages() []mailpitMessage {
	m.t.Helper()
	var list struct {
		Messages []mailpitMessage `json:"messages"`
	}
	require.NoError(m.t, json.Unmarshal(m.api(http.MethodGet, "/api/v1/messages?limit=200", ""), &list))
	return list.Messages
}

// raw returns the stored message with headers, as the server received it.
func (m *mailpit) raw(id string) string {
	m.t.Helper()
	return string(m.api(http.MethodGet, "/api/v1/message/"+id+"/raw", ""))
}

// waitForMessages polls until n messages are stored.
func (m *mailpit) waitForMessages(n int) []mailpitMessage {
	m.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		msgs := m.messages()
		if len(msgs) >= n || time.Now().After(deadline) {
			require.Len(m.t, msgs, n)
			return msgs
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// assertNoMessages gives a late delivery a moment to show up, then checks
// nothing did.
func (m *mailpit) assertNoMessages() {
	m.t.Helper()
	time.Sleep(300 * time.Millisecond)
	assert.Empty(m.t, m.messages())
}

// waitForOne is the common "exactly one delivered message" assertion.
func (m *mailpit) waitForOne() mailpitMessage {
	m.t.Helper()
	return m.waitForMessages(1)[0]
}
