package client

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seededContexts(t *testing.T) (string, map[string]string) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, config.Save(cfgPath, &config.File{
		Contexts: map[string]config.Context{
			"default": {URL: "https://prod.example.com", Token: "t1", Project: "3"},
			"staging": {URL: "https://staging.example.com", TokenFile: "/tmp/does-not-matter-here", CACert: "/etc/ca.pem"},
			"empty":   {URL: "https://empty.example.com"},
		},
	}))
	// "default" is the implicit one, so its inline token is what Resolve
	// reads: no file access.
	return cfgPath, map[string]string{config.EnvConfig: cfgPath}
}

func TestContextList(t *testing.T) {
	_, env := seededContexts(t)

	out, _, err := run(t, env, "", "context", "list")
	require.NoError(t, err)
	assert.Equal(t, "NAME\tURL\tPROJECT\tTOKEN\n"+
		"default\thttps://prod.example.com\t3\tinline\n"+
		"empty\thttps://empty.example.com\t\tnone\n"+
		"staging\thttps://staging.example.com\t\tfile\n", out)

	out, _, err = run(t, env, "", "ctx", "ls", "-q")
	require.NoError(t, err)
	assert.Equal(t, "\n\n\n", out, "contexts have no id; -q prints empty lines rather than names")

	out, _, err = run(t, env, "", "context", "list", "-o", "json")
	require.NoError(t, err)
	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &rows))
	require.Len(t, rows, 3)
	assert.Equal(t, "/tmp/does-not-matter-here", rows[2]["token_file"])
	assert.NotContains(t, out, "t1", "tokens never appear")

	// Empty file.
	out, _, err = run(t, map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "none.yml")}, "", "context", "list", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out)
}

func TestContextGet(t *testing.T) {
	_, env := seededContexts(t)

	out, _, err := run(t, env, "", "context", "get")
	require.NoError(t, err)
	assert.Contains(t, out, "NAME\tdefault\n", "no name given → the context named default")

	out, _, err = run(t, env, "", "context", "show", "staging", "-o", "json")
	require.NoError(t, err)
	var row map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &row))
	assert.Equal(t, "staging", row["name"])
	assert.Equal(t, "/etc/ca.pem", row["ca_cert"])
	assert.NotContains(t, row, "current", "there is no current context any more")

	_, _, err = run(t, env, "", "context", "get", "nope")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
	assert.Equal(t, ExitUsage, ExitCode(err))
}

// There is no `context use`: a saved choice of current context is exactly what
// let `auth login --url other` overwrite a context the user trusted.
func TestContextUse_IsGone(t *testing.T) {
	_, env := seededContexts(t)

	_, _, err := run(t, env, "", "context", "use", "empty")
	assert.Equal(t, ExitUsage, ExitCode(err))
	assert.ErrorContains(t, err, "unknown command")
}

func TestContextSet(t *testing.T) {
	cfgPath, env := seededContexts(t)

	// Create.
	_, errOut, err := run(t, env, "", "context", "set", "lab", "--url", "https://lab.example.com", "--token-file", "~/.semaphore/lab.token", "--insecure", "-p", "9")
	require.NoError(t, err)
	assert.Contains(t, errOut, `Context "lab" created.`)
	f := loadCfg(t, cfgPath)
	assert.Equal(t, config.Context{URL: "https://lab.example.com", TokenFile: "~/.semaphore/lab.token", Insecure: true, Project: "9"}, f.Contexts["lab"])
	assert.Equal(t, config.Context{URL: "https://prod.example.com", Token: "t1", Project: "3"}, f.Contexts["default"], "the implicit context is untouched")

	// Change one field; the inline token is replaced by a file reference.
	_, _, err = run(t, env, "", "context", "set", "default", "--token-file", "/tmp/prod.token", "--no-insecure")
	require.NoError(t, err)
	f = loadCfg(t, cfgPath)
	assert.Equal(t, config.Context{URL: "https://prod.example.com", TokenFile: "/tmp/prod.token", Project: "3"}, f.Contexts["default"])

	_, _, err = run(t, env, "", "context", "set", "default", "--no-project", "--ca-cert", "/etc/new.pem")
	require.NoError(t, err)
	f = loadCfg(t, cfgPath)
	assert.Equal(t, "", f.Contexts["default"].Project)
	assert.Equal(t, "/etc/new.pem", f.Contexts["default"].CACert)

	// Mistakes.
	_, _, err = run(t, env, "", "context", "set", "new-without-url")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "--url")

	_, _, err = run(t, env, "", "context", "set", "empty")
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "nothing to change")

	// A context created under another name stays an alternative: nothing in
	// the file marks it as the one to use.
	env2 := map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "client.yml")}
	_, _, err = run(t, env2, "", "context", "set", "first", "--url", "https://first.example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"first"}, loadCfg(t, env2[config.EnvConfig]).ContextNames())
}

func TestContextDelete(t *testing.T) {
	cfgPath, env := seededContexts(t)

	_, _, err := run(t, env, "", "context", "delete", "empty")
	var u *UsageError
	require.ErrorAs(t, err, &u, "no TTY and no --yes")
	assert.Contains(t, err.Error(), "--yes")
	assert.Len(t, loadCfg(t, cfgPath).Contexts, 3)

	_, errOut, err := run(t, env, "", "context", "rm", "empty", "-y")
	require.NoError(t, err)
	assert.Contains(t, errOut, `Context "empty" deleted.`)
	assert.Equal(t, []string{"default", "staging"}, loadCfg(t, cfgPath).ContextNames())

	// Interactive decline, then accept.
	_, _, err = runWith(t, env, "n\n", true, "context", "delete", "default")
	assert.ErrorIs(t, err, ErrAborted)

	_, _, err = runWith(t, env, "y\n", true, "context", "delete", "default")
	require.NoError(t, err)
	assert.Equal(t, []string{"staging"}, loadCfg(t, cfgPath).ContextNames())

	_, _, err = run(t, env, "", "context", "delete", "nope", "-y")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
}
