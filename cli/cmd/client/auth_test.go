package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// authServer fakes the login, session and token endpoints of a Semaphore
// server closely enough for the auth commands.
type authServer struct {
	mu sync.Mutex

	passwordLogin bool
	totp          bool
	validToken    string // the one API token the server accepts
	password      string

	sessionVerified bool
	sessionOpen     bool
	created         []map[string]any // bodies of POST /api/user/tokens
	deleted         []string         // ids of DELETE /api/user/tokens/{id}
	loggedOut       int
}

func newAuthServer() *authServer {
	return &authServer{passwordLogin: true, validToken: "good-token", password: "pw"}
}

const createdToken = "freshly-created-token-value"

func (s *authServer) authorized(r *http.Request) bool {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth == "Bearer "+s.validToken || auth == "Bearer "+createdToken
	}
	c, err := r.Cookie("semaphore")
	return err == nil && c.Value == "sess" && s.sessionOpen && s.sessionVerified
}

func (s *authServer) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method + " " + r.URL.Path {
		case "GET /api/ping":
			_, _ = w.Write([]byte("pong"))
		case "GET /api/auth/login":
			m := apiclient.LoginMethods{LoginWithPassword: s.passwordLogin}
			if s.totp {
				m.AuthMethods.Totp = &struct{}{}
			}
			_ = json.NewEncoder(w).Encode(m)
		case "POST /api/auth/login":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["auth"] != "admin" || body["password"] != s.password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			s.sessionOpen, s.sessionVerified = true, !s.totp
			http.SetCookie(w, &http.Cookie{Name: "semaphore", Value: "sess", Path: "/"})
			w.WriteHeader(http.StatusNoContent)
		case "POST /api/auth/verify":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if !s.sessionOpen || body["passcode"] != "123456" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			s.sessionVerified = true
			w.WriteHeader(http.StatusNoContent)
		case "POST /api/auth/logout":
			s.loggedOut++
			s.sessionOpen = false
			w.WriteHeader(http.StatusNoContent)
		case "GET /api/user":
			if !s.authorized(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(apiclient.Me{User: db.User{ID: 1, Username: "admin", Name: "Admin", Admin: true}})
		case "GET /api/info":
			if !s.authorized(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(apiclient.SystemInfo{Version: "2.20.3"})
		case "POST /api/user/tokens":
			if !s.authorized(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			s.created = append(s.created, body)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(db.APIToken{ID: createdToken, Name: "cli"})
		default:
			if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/user/tokens/") {
				if !s.authorized(r) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				s.deleted = append(s.deleted, strings.TrimPrefix(r.URL.Path, "/api/user/tokens/"))
				w.WriteHeader(http.StatusNoContent)
				return
			}
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func loadCfg(t *testing.T, path string) *config.File {
	t.Helper()
	f, _, err := config.Load(path)
	require.NoError(t, err)
	return f
}

func TestAuthLogin_WithTokenFromStdin(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	env := map[string]string{config.EnvConfig: cfgPath}

	out, errOut, err := run(t, env, "good-token\n", "auth", "login", "--url", srv.URL)
	require.NoError(t, err)
	assert.Empty(t, out, "nothing on stdout")
	assert.Contains(t, errOut, "Logged in to "+srv.URL+" as admin.")
	assert.Contains(t, errOut, `Context "127.0.0.1" saved to `+cfgPath)

	f := loadCfg(t, cfgPath)
	assert.Equal(t, "127.0.0.1", f.CurrentContext)
	assert.Equal(t, config.Context{URL: srv.URL, Token: "good-token"}, f.Contexts["127.0.0.1"])
	if goruntime.GOOS != "windows" {
		info, err := os.Stat(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	// The saved context now drives every other command.
	out, _, err = run(t, env, "", "auth", "status", "-o", "json")
	require.NoError(t, err)
	var st map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &st))
	assert.Equal(t, "127.0.0.1", st["context"])
	assert.Equal(t, "admin", st["user"])
	assert.Equal(t, "context:token", st["token_source"])
	assert.Equal(t, "2.20.3", st["server_version"])
	assert.Equal(t, cfgPath, st["config_file"])

	out, _, err = run(t, env, "", "auth", "token")
	require.NoError(t, err)
	assert.Equal(t, "good-token\n", out)

	out, _, err = run(t, env, "", "context", "list")
	require.NoError(t, err)
	assert.Equal(t, "NAME\tCURRENT\tURL\tPROJECT\tTOKEN\n127.0.0.1\ttrue\t"+srv.URL+"\t\tinline\n", out)
	assert.NotContains(t, out, "good-token", "the token is never listed")
}

func TestAuthLogin_TokenErrors(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	env := map[string]string{config.EnvConfig: cfgPath}

	_, _, err := run(t, env, "", "auth", "login", "--url", srv.URL, "--with-token")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "no token on stdin")

	_, _, err = run(t, env, "bad-token\n", "auth", "login", "--url", srv.URL)
	assert.ErrorIs(t, err, apiclient.ErrUnauthorized)
	assert.Equal(t, ExitAuth, ExitCode(err))
	_, statErr := os.Stat(cfgPath)
	assert.True(t, os.IsNotExist(statErr), "nothing is saved after a rejected token")

	_, _, err = run(t, env, "x\n", "auth", "login")
	require.ErrorAs(t, err, &u, "no URL anywhere")

	// Not a Semaphore server.
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) }))
	defer other.Close()
	_, _, err = run(t, env, "x\n", "auth", "login", "--url", other.URL)
	assert.ErrorContains(t, err, "does not look like a Semaphore server")
}

func TestAuthLogin_Password(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	env := map[string]string{config.EnvConfig: cfgPath}

	_, errOut, err := runWith(t, env, "admin\npw\n", true, "auth", "login", "--url", srv.URL, "--name", "laptop", "--ttl", "720h", "--context", "work")
	require.NoError(t, err)
	assert.Contains(t, errOut, "Login: ")
	assert.Contains(t, errOut, "Password: ")
	assert.Contains(t, errOut, "Logged in to "+srv.URL+" as admin.")
	assert.NotContains(t, errOut, "pw", "the password is never echoed")

	require.Len(t, s.created, 1)
	assert.Equal(t, "laptop", s.created[0]["name"])
	exp, err := time.Parse(time.RFC3339, s.created[0]["expires_at"].(string))
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(720*time.Hour), exp, time.Minute)
	assert.Equal(t, 1, s.loggedOut, "the web session is closed")

	f := loadCfg(t, cfgPath)
	assert.Equal(t, "work", f.CurrentContext)
	assert.Equal(t, createdToken, f.Contexts["work"].Token)

	// --login skips the first prompt; a default token name is used.
	_, errOut, err = runWith(t, env, "pw\n", true, "auth", "login", "--url", srv.URL, "--login", "admin")
	require.NoError(t, err)
	assert.NotContains(t, errOut, "Login: ")
	require.Len(t, s.created, 2)
	assert.True(t, strings.HasPrefix(s.created[1]["name"].(string), "semaphore-cli@"), s.created[1]["name"])
	assert.Nil(t, s.created[1]["expires_at"])
}

func TestAuthLogin_PasswordErrors(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "client.yml")}

	_, _, err := runWith(t, env, "admin\nwrong\n", true, "auth", "login", "--url", srv.URL)
	assert.ErrorIs(t, err, apiclient.ErrUnauthorized)
	assert.ErrorContains(t, err, "wrong login or password")
	assert.Equal(t, ExitAuth, ExitCode(err))
	assert.Empty(t, s.created)

	_, _, err = runWith(t, env, "admin\n\n", true, "auth", "login", "--url", srv.URL)
	var u *UsageError
	require.ErrorAs(t, err, &u)

	// --no-input in a terminal: the password flow cannot run, stdin is read as a token.
	_, _, err = runWith(t, env, "", true, "auth", "login", "--url", srv.URL, "--no-input")
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "no token on stdin")

	// SSO-only server.
	s.passwordLogin = false
	_, _, err = runWith(t, env, "admin\npw\n", true, "auth", "login", "--url", srv.URL)
	assert.ErrorContains(t, err, srv.URL+"/tokens")
	assert.ErrorContains(t, err, "--with-token")
}

func TestAuthLogin_TOTP(t *testing.T) {
	s := newAuthServer()
	s.totp = true
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "client.yml")}

	_, errOut, err := runWith(t, env, "admin\npw\n123456\n", true, "auth", "login", "--url", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, errOut, "One-time code: ")
	assert.Len(t, s.created, 1)

	s.sessionOpen, s.sessionVerified = false, false
	_, _, err = runWith(t, env, "admin\npw\n000000\n", true, "auth", "login", "--url", srv.URL)
	assert.ErrorContains(t, err, "two-factor verification failed")
	assert.Equal(t, ExitAuth, ExitCode(err))
}

func TestAuthLogin_KeepsProjectAndTLSOfExistingContext(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, config.Save(cfgPath, &config.File{
		CurrentContext: "old",
		Contexts:       map[string]config.Context{"old": {URL: srv.URL, Token: "stale", Project: "7", TokenFile: ""}},
	}))
	env := map[string]string{config.EnvConfig: cfgPath}

	// Re-login into the current context without --url: the context's URL is used.
	_, _, err := run(t, env, "good-token\n", "auth", "login", "--insecure")
	require.NoError(t, err)
	f := loadCfg(t, cfgPath)
	assert.Equal(t, "old", f.CurrentContext)
	assert.Equal(t, config.Context{URL: srv.URL, Token: "good-token", Project: "7", Insecure: true}, f.Contexts["old"])

	// A new named context must get its URL from --url, not from the current one.
	_, _, err = run(t, env, "good-token\n", "auth", "login", "--context", "fresh")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "server URL is not set")

	_, _, err = run(t, env, "good-token\n", "auth", "login", "--context", "fresh", "--url", srv.URL)
	require.NoError(t, err)
	f = loadCfg(t, cfgPath)
	assert.Equal(t, "fresh", f.CurrentContext)
	assert.Equal(t, []string{"fresh", "old"}, f.ContextNames())
}

func TestAuthLogout(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, config.Save(cfgPath, &config.File{
		CurrentContext: "prod",
		Contexts: map[string]config.Context{
			"prod":    {URL: srv.URL, Token: "good-token", Project: "3"},
			"staging": {URL: srv.URL, Token: "good-token"},
		},
	}))
	env := map[string]string{config.EnvConfig: cfgPath}

	_, errOut, err := run(t, env, "", "auth", "logout", "--revoke")
	require.NoError(t, err)
	assert.Equal(t, []string{"good-tok"}, s.deleted, "revoked by the 8-character prefix the server lists")
	assert.Contains(t, errOut, "Token revoked")
	assert.Contains(t, errOut, `Token removed from context "prod"`)
	f := loadCfg(t, cfgPath)
	assert.Equal(t, config.Context{URL: srv.URL, Project: "3"}, f.Contexts["prod"], "URL and project survive")
	assert.Equal(t, "good-token", f.Contexts["staging"].Token, "other contexts untouched")

	// Logging out again is a no-op.
	_, errOut, err = run(t, env, "", "auth", "logout")
	require.NoError(t, err)
	assert.Contains(t, errOut, "holds no token")

	// Without --revoke nothing is deleted on the server.
	_, _, err = run(t, env, "", "auth", "logout", "--context", "staging")
	require.NoError(t, err)
	assert.Len(t, s.deleted, 1)
	assert.Empty(t, loadCfg(t, cfgPath).Contexts["staging"].Token)

	// --revoke with the token coming from the environment is refused.
	require.NoError(t, config.Save(cfgPath, &config.File{CurrentContext: "prod", Contexts: map[string]config.Context{"prod": {URL: srv.URL, Token: "good-token"}}}))
	_, _, err = run(t, map[string]string{config.EnvConfig: cfgPath, config.EnvURL: srv.URL, config.EnvToken: "good-token"}, "", "auth", "logout", "--revoke")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "env:"+config.EnvToken)

	// No context at all.
	_, _, err = run(t, map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "none.yml")}, "", "auth", "logout")
	require.ErrorAs(t, err, &u)
}

func TestAuthStatus_RejectedToken(t *testing.T) {
	s := newAuthServer()
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, errOut, err := run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "stale"}, "", "auth", "status")
	assert.ErrorIs(t, err, apiclient.ErrUnauthorized)
	assert.Equal(t, ExitAuth, ExitCode(err))
	assert.Contains(t, errOut, "rejected the token (source: env:"+config.EnvToken+")")

	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL}, "", "auth", "status")
	var u *UsageError
	require.ErrorAs(t, err, &u)

	_, _, err = run(t, nil, "", "auth", "token")
	require.ErrorAs(t, err, &u)
}

func TestTokenPrefixAndHostOf(t *testing.T) {
	assert.Equal(t, "abcdefgh", tokenPrefix("abcdefghijkl"))
	assert.Equal(t, "short", tokenPrefix("short"))
	assert.Equal(t, "semaphore.example.com", hostOf("https://semaphore.example.com:8443/semaphore"))
	assert.Equal(t, "default", hostOf("::bad"))
}
