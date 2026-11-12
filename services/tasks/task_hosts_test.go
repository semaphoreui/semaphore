package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHostsDocument(t *testing.T) {
	doc := `{"version":1,"hosts":{"web-02":{"ok":1},"Web-01":{"ok":2,"changed":1,"unreachable":0,"failed":1,"skipped":0,"rescued":0,"ignored":0,"extra":5}}}`
	hosts, err := parseHostsDocument(doc, 7, 42)
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	assert.Equal(t, db.AnsibleTaskHost{ProjectID: 7, TaskID: 42, Host: "Web-01", Ok: 2, Changed: 1, Failed: 1}, hosts[0])
	assert.Equal(t, db.AnsibleTaskHost{ProjectID: 7, TaskID: 42, Host: "web-02", Ok: 1}, hosts[1])
}

func TestParseHostsDocument_Empty(t *testing.T) {
	hosts, err := parseHostsDocument(`{"version":1,"hosts":{}}`, 7, 42)
	require.NoError(t, err)
	assert.Empty(t, hosts)
}

func TestParseHostsDocument_Invalid(t *testing.T) {
	for name, doc := range map[string]string{
		"not json":     `{`,
		"not object":   `[]`,
		"version":      `{"version":2,"hosts":{}}`,
		"no hosts":     `{"version":1}`,
		"empty name":   `{"version":1,"hosts":{"":{}}}`,
		"long name":    `{"version":1,"hosts":{"` + strings.Repeat("a", 251) + `":{}}}`,
		"negative":     `{"version":1,"hosts":{"a":{"ok":-1}}}`,
		"null stats":   `{"version":1,"hosts":{"a":null}}`,
		"null counter": `{"version":1,"hosts":{"a":{"ok":null}}}`,
		"not a number": `{"version":1,"hosts":{"a":{"ok":"1"}}}`,
	} {
		_, err := parseHostsDocument(doc, 7, 42)
		assert.Error(t, err, name)
	}
}

func TestHostsCapture(t *testing.T) {
	dir := t.TempDir()
	c, err := beginHostsCapture(dir, 42)
	require.NoError(t, err)
	path := strings.TrimPrefix(c.Env()[0], HostsFileEnv+"=")
	assert.Equal(t, dir, filepath.Dir(filepath.Dir(path)))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	got, err := c.Read()
	require.NoError(t, err)
	assert.Nil(t, got)

	require.NoError(t, os.WriteFile(path, []byte(`{"version":1,"hosts":{}}`), 0o600))
	got, err = c.Read()
	require.NoError(t, err)
	assert.Equal(t, `{"version":1,"hosts":{}}`, *got)

	plugins, err := c.CallbackDir()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(plugins, "semaphore_hosts.py"))
	require.NoError(t, err)

	require.NoError(t, os.Truncate(path, MaxHostsDocumentSize+1))
	_, err = c.Read()
	assert.Error(t, err)

	c.Close()
	_, err = os.Stat(filepath.Dir(path))
	assert.ErrorIs(t, err, os.ErrNotExist)
}
