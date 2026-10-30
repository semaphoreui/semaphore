package apiclient

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient builds a client against srv with a fixed token.
func newTestClient(t *testing.T, srv *httptest.Server, mod func(*Options)) *Client {
	t.Helper()
	opts := Options{BaseURL: srv.URL, Token: "secret-token", Timeout: 5 * time.Second}
	if mod != nil {
		mod(&opts)
	}
	c, err := New(opts)
	require.NoError(t, err)
	return c
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"plain", "https://semaphore.example.com", "https://semaphore.example.com", false},
		{"trailing slash", "https://semaphore.example.com/", "https://semaphore.example.com", false},
		{"trailing api", "https://semaphore.example.com/api", "https://semaphore.example.com", false},
		{"trailing api slash", "https://semaphore.example.com/api/", "https://semaphore.example.com", false},
		{"sub path", "https://host/semaphore/", "https://host/semaphore", false},
		{"query dropped", "https://host/?x=1#frag", "https://host", false},
		{"empty", "", "", true},
		{"no scheme", "semaphore.example.com", "", true},
		{"bad scheme", "ftp://host", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := normalizeBaseURL(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, u.String())
		})
	}
}

func TestClient_URL(t *testing.T) {
	c, err := New(Options{BaseURL: "https://host/semaphore/"})
	require.NoError(t, err)
	assert.Equal(t, "https://host/semaphore/api/project/1/tasks", c.URL("/project/1/tasks"))
	assert.Equal(t, "https://host/semaphore/api/ping", c.URL("ping"))
	assert.Equal(t, "https://host/semaphore/api/project/1/tasks/last?before=4&count=2", c.URL("/project/1/tasks/last?before=4&count=2"))
}

func TestDo_SendsBearerAndDecodes(t *testing.T) {
	var gotAuth, gotUA, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		assert.Equal(t, "/api/projects", r.URL.Path)
		_ = json.NewEncoder(w).Encode([]db.Project{{ID: 1, Name: "one"}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv, func(o *Options) { o.UserAgent = "semaphore-cli/test" })
	projects, err := c.ListProjects(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", gotAuth)
	assert.Equal(t, "semaphore-cli/test", gotUA)
	assert.Equal(t, "application/json", gotAccept)
	assert.Equal(t, []db.Project{{ID: 1, Name: "one"}}, projects)
}

func TestDo_NoTokenNoHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, has := r.Header["Authorization"]
		assert.False(t, has)
		_, _ = w.Write([]byte("pong"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, func(o *Options) { o.Token = "" })
	assert.NoError(t, c.Ping(context.Background()))
}

func TestDo_PostEncodesJSON(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(db.Task{ID: 42, Status: "waiting"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv, nil)
	task, err := c.RunTask(context.Background(), 7, db.Task{TemplateName: "deploy"})
	require.NoError(t, err)
	assert.Equal(t, 42, task.ID)
	assert.Equal(t, "deploy", got["template_name"])
	assert.EqualValues(t, 7, got["project_id"])
}

func TestDo_EmptyBodyWithOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, nil)
	var out db.Project
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/x", nil, &out))
	assert.Equal(t, db.Project{}, out)
}

func TestErrors_StatusMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		sentinel error
		message  string
	}{
		{"401 unauthorized", http.StatusUnauthorized, "", ErrUnauthorized, ""},
		{"403 forbidden json", http.StatusForbidden, `{"error":"You have no subscription."}`, ErrUnauthorized, "You have no subscription."},
		{"404", http.StatusNotFound, "", ErrNotFound, ""},
		{"400 plain text", http.StatusBadRequest, "bad cron", nil, "bad cron"},
		{"502 html hidden", http.StatusBadGateway, "<html><body>nginx</body></html>", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := newTestClient(t, srv, nil)
			err := c.Do(context.Background(), http.MethodGet, "/project/1/templates/5", nil, nil)
			require.Error(t, err)

			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.status, apiErr.Status)
			assert.Equal(t, tt.message, apiErr.Message)
			assert.Equal(t, "/api/project/1/templates/5", apiErr.Path)
			if tt.sentinel != nil {
				assert.ErrorIs(t, err, tt.sentinel)
			} else {
				assert.False(t, errors.Is(err, ErrUnauthorized))
				assert.False(t, errors.Is(err, ErrNotFound))
			}
		})
	}
}

func TestErrors_Transport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // connection refused from now on

	c := newTestClient(t, srv, nil)
	err := c.Ping(context.Background())
	var tErr *TransportError
	assert.ErrorAs(t, err, &tErr)
}

func TestRedirectsAreRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example.com/api/ping", http.StatusFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, nil)
	err := c.Ping(context.Background())
	var tErr *TransportError
	require.ErrorAs(t, err, &tErr)
	assert.Contains(t, err.Error(), "redirects are not followed")
}

func TestDebug_MasksSecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "semaphore", Value: "session-cookie-value"})
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "tok-id-value", "name": "cli"})
	}))
	defer srv.Close()

	var trace bytes.Buffer
	c := newTestClient(t, srv, func(o *Options) { o.Debug = &trace })
	err := c.Do(context.Background(), http.MethodPost, "/auth/login", map[string]string{"auth": "admin", "password": "p@ss"}, &map[string]any{})
	require.NoError(t, err)

	out := trace.String()
	assert.Contains(t, out, "> POST "+srv.URL+"/api/auth/login")
	assert.Contains(t, out, "< 200 OK")
	assert.Contains(t, out, `"auth":"admin"`)
	assert.NotContains(t, out, "secret-token")
	assert.NotContains(t, out, "p@ss")
	assert.NotContains(t, out, "session-cookie-value")
	assert.Contains(t, out, "Authorization: ***")
	assert.Contains(t, out, "Set-Cookie: ***")
}

func TestRedactJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"password", `{"auth":"a","password":"x"}`, `{"auth":"a","password":"***"}`},
		{"token nested", `{"runner":{"token":"abc"},"name":"r"}`, `{"runner":{"token":"***"},"name":"r"}`},
		{"escaped quote", `{"secret":"a\"b"}`, `{"secret":"***"}`},
		{"array untouched", `[{"id":1}]`, `[{"id":1}]`},
		{"not json", "pong\n", "pong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, redactJSON([]byte(tt.in)))
		})
	}
}

func TestTLS_CustomCAAndInsecure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}))
	defer srv.Close()

	// Without trust the handshake fails.
	c := newTestClient(t, srv, nil)
	var tErr *TransportError
	assert.ErrorAs(t, c.Ping(context.Background()), &tErr)

	// Insecure skips verification.
	c = newTestClient(t, srv, func(o *Options) { o.Insecure = true })
	assert.NoError(t, c.Ping(context.Background()))

	// A CA file with the server certificate verifies it properly.
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	require.NoError(t, os.WriteFile(caFile, certPEM, 0o600))
	c = newTestClient(t, srv, func(o *Options) { o.CACertFile = caFile })
	assert.NoError(t, c.Ping(context.Background()))

	// A file without certificates is rejected at construction.
	bad := filepath.Join(t.TempDir(), "bad.pem")
	require.NoError(t, os.WriteFile(bad, []byte("not pem"), 0o600))
	_, err := New(Options{BaseURL: srv.URL, CACertFile: bad})
	assert.ErrorContains(t, err, "no PEM certificates")

	_, err = New(Options{BaseURL: srv.URL, CACertFile: filepath.Join(t.TempDir(), "missing.pem")})
	assert.ErrorContains(t, err, "read CA certificate")

	_ = x509.NewCertPool() // keep the import honest if the pool helper changes
}

func TestPing_NotSemaphore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>hello</html>"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, nil)
	assert.ErrorContains(t, c.Ping(context.Background()), "does not look like a Semaphore server")
}
