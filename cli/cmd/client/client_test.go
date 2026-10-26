package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// run executes `semaphore client <args>` against an in-memory environment and
// returns stdout, stderr and the error cobra reported.
func run(t *testing.T, env map[string]string, stdin string, args ...string) (string, string, error) {
	t.Helper()
	if env == nil {
		env = map[string]string{}
	}
	if _, ok := env[config.EnvConfig]; !ok {
		// Never read the developer's real ~/.semaphore/client.yml.
		env[config.EnvConfig] = filepath.Join(t.TempDir(), "client.yml")
	}
	var out, errOut bytes.Buffer
	cmd := NewCommand(IO{
		In:      strings.NewReader(stdin),
		Out:     &out,
		Err:     &errOut,
		Getenv:  func(k string) string { return env[k] },
		Version: "2.20.0-test",
	})
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

func infoServer(t *testing.T, version string, wantToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/info" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+wantToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(apiclient.SystemInfo{Version: version, WebHost: "https://web", ScheduleTimezone: "UTC"})
	}))
}

func TestInfo_EnvOnly(t *testing.T) {
	srv := infoServer(t, "2.20.3", "tok")
	defer srv.Close()

	out, errOut, err := run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "info")
	require.NoError(t, err)
	// A single record is shown vertically, every field, id-less records alphabetically.
	assert.Contains(t, out, "URL\t"+srv.URL+"\n")
	assert.Contains(t, out, "SERVER VERSION\t2.20.3\n")
	assert.Contains(t, out, "CLIENT VERSION\t2.20.0-test\n")
	assert.Contains(t, out, "VERSIONS COMPATIBLE\ttrue\n")
	assert.NotContains(t, errOut, "plain http", "loopback over http does not warn")
	assert.NotContains(t, errOut, "differ in major or minor")
}

func TestInfo_JSONAndVersionWarning(t *testing.T) {
	srv := infoServer(t, "v2.21.0", "tok")
	defer srv.Close()

	out, errOut, err := run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "info", "-o", "json")
	require.NoError(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	assert.Equal(t, "v2.21.0", res["server_version"])
	assert.Equal(t, "2.20.0-test", res["client_version"])
	assert.Equal(t, false, res["versions_compatible"])
	assert.Contains(t, errOut, "differ in major or minor version")
}

func TestInfo_ContextFile(t *testing.T) {
	srv := infoServer(t, "2.20.3", "ctx-token")
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, config.Save(cfgPath, &config.File{
		CurrentContext: "lab",
		Contexts:       map[string]config.Context{"lab": {URL: srv.URL, Token: "ctx-token"}, "other": {URL: "http://127.0.0.1:1", Token: "x"}},
		Defaults:       config.Defaults{Output: "yaml"},
	}))

	out, _, err := run(t, map[string]string{config.EnvConfig: cfgPath}, "", "info")
	require.NoError(t, err)
	assert.Contains(t, out, "server_version: 2.20.3\n", "defaults.output from the file applies")
	assert.Contains(t, out, "context: lab\n")

	// Flag beats file.
	out, _, err = run(t, map[string]string{config.EnvConfig: cfgPath}, "", "info", "-o", "json", "--fields", "context")
	require.NoError(t, err)
	assert.JSONEq(t, `{"context":"lab"}`, out)

	// An env token must not ride along to a context's URL.
	_, _, err = run(t, map[string]string{config.EnvConfig: cfgPath, config.EnvToken: "ci"}, "", "info")
	assert.ErrorIs(t, err, config.ErrTokenWithoutURL)
	assert.Equal(t, ExitUsage, ExitCode(err))

	// Unknown context.
	_, _, err = run(t, map[string]string{config.EnvConfig: cfgPath}, "", "info", "--context", "nope")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
}

func TestInfo_Errors(t *testing.T) {
	srv := infoServer(t, "2.20.3", "tok")
	defer srv.Close()

	// No URL anywhere.
	_, _, err := run(t, nil, "", "info")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), config.EnvURL)
	assert.Equal(t, ExitUsage, ExitCode(err))

	// URL but no token.
	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL}, "", "info")
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "no API token")

	// Wrong token → 4.
	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "wrong"}, "", "info")
	assert.ErrorIs(t, err, apiclient.ErrUnauthorized)
	assert.Equal(t, ExitAuth, ExitCode(err))

	// Unknown flag → usage error with hint, no cobra usage dump on stdout.
	out, errOut, err := run(t, nil, "", "info", "--bogus")
	require.ErrorAs(t, err, &u)
	assert.Equal(t, ExitUsage, ExitCode(err))
	assert.Contains(t, err.Error(), "unknown flag: --bogus")
	assert.Contains(t, err.Error(), "Run 'client info --help' for usage.", "the test tree has no `semaphore` root")
	assert.Empty(t, out)
	assert.Empty(t, errOut, "cobra's own error/usage printing is silenced")

	// Unexpected argument.
	_, _, err = run(t, nil, "", "info", "extra")
	assert.Error(t, err)

	// Bad output format.
	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "info", "-o", "xml")
	assert.ErrorIs(t, err, config.ErrBadOutput)
}

func TestInsecureWarningAndConfigPermissions(t *testing.T) {
	srv := infoServer(t, "2.20.3", "tok")
	defer srv.Close()

	_, errOut, err := run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok", config.EnvInsecure: "1"}, "", "info")
	require.NoError(t, err)
	assert.Contains(t, errOut, "TLS certificate verification is disabled")

	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("contexts: {}\n"), 0o644))
	_, errOut, err = run(t, map[string]string{config.EnvConfig: cfgPath, config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "info")
	require.NoError(t, err)
	assert.Contains(t, errOut, "warning: "+cfgPath)
	assert.Contains(t, errOut, "chmod 600")

	require.NoError(t, os.Chmod(cfgPath, 0o666))
	_, _, err = run(t, map[string]string{config.EnvConfig: cfgPath, config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "info")
	assert.ErrorIs(t, err, config.ErrInsecurePermissions)
}

func TestClientHelpListsFlags(t *testing.T) {
	out, _, err := run(t, nil, "", "--help")
	require.NoError(t, err)
	for _, flag := range []string{"--url", "--context", "-p, --project", "-o, --output", "--fields", "-q, --quiet", "--no-headers", "--no-color", "-y, --yes", "--no-input", "--insecure", "--ca-cert", "--timeout", "--debug"} {
		assert.Contains(t, out, flag)
	}
	assert.NotContains(t, out, "--token", "there is no --token flag on purpose")
}

func TestSameMinor(t *testing.T) {
	assert.True(t, sameMinor("v2.20.3", "2.20.0-abc-date"))
	assert.False(t, sameMinor("2.21.0", "2.20.9"))
	assert.False(t, sameMinor("3.0.0", "2.20.9"))
	assert.True(t, sameMinor("development", "2.20.0"), "unknown versions never warn")
	assert.True(t, sameMinor("", "2.20.0"))
}

func TestIsLoopback(t *testing.T) {
	for _, u := range []string{"http://localhost:3000", "http://127.0.0.1", "http://127.1.2.3:8080/x", "http://[::1]:3000"} {
		assert.True(t, isLoopback(u), u)
	}
	for _, u := range []string{"http://semaphore.internal", "http://10.0.0.5:3000", "https://example.com"} {
		assert.False(t, isLoopback(u), u)
	}
}
