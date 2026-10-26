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
		CurrentContext: "prod",
		Contexts: map[string]config.Context{
			"prod":    {URL: "https://prod.example.com", Token: "t1", Project: "3"},
			"staging": {URL: "https://staging.example.com", TokenFile: "/tmp/does-not-matter-here", CACert: "/etc/ca.pem"},
			"empty":   {URL: "https://empty.example.com"},
		},
	}))
	// prod is current, so its inline token is what Resolve reads: no file access.
	return cfgPath, map[string]string{config.EnvConfig: cfgPath}
}

func TestContextList(t *testing.T) {
	_, env := seededContexts(t)

	out, _, err := run(t, env, "", "context", "list")
	require.NoError(t, err)
	assert.Equal(t, "NAME\tCURRENT\tURL\tPROJECT\tTOKEN\n"+
		"empty\tfalse\thttps://empty.example.com\t\tnone\n"+
		"prod\ttrue\thttps://prod.example.com\t3\tinline\n"+
		"staging\tfalse\thttps://staging.example.com\t\tfile\n", out)

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
	assert.Contains(t, out, "NAME\tprod\n")
	assert.Contains(t, out, "CURRENT\ttrue\n")

	out, _, err = run(t, env, "", "context", "show", "staging", "-o", "json")
	require.NoError(t, err)
	var row map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &row))
	assert.Equal(t, "staging", row["name"])
	assert.Equal(t, "/etc/ca.pem", row["ca_cert"])
	assert.Equal(t, false, row["current"])

	_, _, err = run(t, env, "", "context", "get", "nope")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
	assert.Equal(t, ExitUsage, ExitCode(err))
}

func TestContextUse(t *testing.T) {
	cfgPath, env := seededContexts(t)

	_, errOut, err := run(t, env, "", "context", "use", "empty")
	require.NoError(t, err)
	assert.Contains(t, errOut, `Current context is now "empty".`)
	assert.Equal(t, "empty", loadCfg(t, cfgPath).CurrentContext)

	_, _, err = run(t, env, "", "context", "use", "nope")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
}

func TestContextSet(t *testing.T) {
	cfgPath, env := seededContexts(t)

	// Create.
	_, errOut, err := run(t, env, "", "context", "set", "lab", "--url", "https://lab.example.com", "--token-file", "~/.semaphore/lab.token", "--insecure", "-p", "9")
	require.NoError(t, err)
	assert.Contains(t, errOut, `Context "lab" created.`)
	f := loadCfg(t, cfgPath)
	assert.Equal(t, config.Context{URL: "https://lab.example.com", TokenFile: "~/.semaphore/lab.token", Insecure: true, Project: "9"}, f.Contexts["lab"])
	assert.Equal(t, "prod", f.CurrentContext, "creating does not switch the current context")

	// Change one field; the inline token is replaced by a file reference.
	_, _, err = run(t, env, "", "context", "set", "prod", "--token-file", "/tmp/prod.token", "--no-insecure")
	require.NoError(t, err)
	f = loadCfg(t, cfgPath)
	assert.Equal(t, config.Context{URL: "https://prod.example.com", TokenFile: "/tmp/prod.token", Project: "3"}, f.Contexts["prod"])

	_, _, err = run(t, env, "", "context", "set", "prod", "--no-project", "--ca-cert", "/etc/new.pem")
	require.NoError(t, err)
	f = loadCfg(t, cfgPath)
	assert.Equal(t, "", f.Contexts["prod"].Project)
	assert.Equal(t, "/etc/new.pem", f.Contexts["prod"].CACert)

	// Mistakes.
	_, _, err = run(t, env, "", "context", "set", "new-without-url")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "--url")

	_, _, err = run(t, env, "", "context", "set", "empty")
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "nothing to change")

	// The first context created in an empty file becomes current.
	env2 := map[string]string{config.EnvConfig: filepath.Join(t.TempDir(), "client.yml")}
	_, _, err = run(t, env2, "", "context", "set", "first", "--url", "https://first.example.com")
	require.NoError(t, err)
	assert.Equal(t, "first", loadCfg(t, env2[config.EnvConfig]).CurrentContext)
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
	assert.Equal(t, []string{"prod", "staging"}, loadCfg(t, cfgPath).ContextNames())

	// Interactive decline, then accept; deleting the current context clears it.
	_, _, err = runWith(t, env, "n\n", true, "context", "delete", "prod")
	assert.ErrorIs(t, err, ErrAborted)

	_, _, err = runWith(t, env, "y\n", true, "context", "delete", "prod")
	require.NoError(t, err)
	f := loadCfg(t, cfgPath)
	assert.Equal(t, "", f.CurrentContext)
	assert.Equal(t, []string{"staging"}, f.ContextNames())

	_, _, err = run(t, env, "", "context", "delete", "nope", "-y")
	assert.ErrorIs(t, err, config.ErrUnknownContext)
}
