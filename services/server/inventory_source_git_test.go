package server

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInventoryPath(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "inventory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "inventory", "prod.yml"), []byte("all:\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "x"), []byte("all:\n"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link")))
	realRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	got, err := resolveInventoryPath(root, "inventory/prod.yml")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(realRoot, "inventory", "prod.yml"), got)

	got, err = resolveInventoryPath(root, "inventory")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(realRoot, "inventory"), got)

	require.NoError(t, os.Symlink("prod.yml", filepath.Join(root, "inventory", "alias.yml")))
	got, err = resolveInventoryPath(root, "inventory/alias.yml")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(realRoot, "inventory", "alias.yml"), got)

	up, err := filepath.Rel(root, filepath.Join(outside, "x"))
	require.NoError(t, err)
	_, err = resolveInventoryPath(root, up)
	assert.ErrorContains(t, err, "outside")

	_, err = resolveInventoryPath(root, "link/x")
	assert.ErrorContains(t, err, "outside")
}

func TestRejectOutsideLinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "hosts.yml"), []byte("all:\n"), 0o644))
	require.NoError(t, os.Symlink("hosts.yml", filepath.Join(root, "inside.yml")))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, ".git", "out")))

	assert.NoError(t, rejectOutsideLinks(root))

	require.NoError(t, os.MkdirAll(filepath.Join(root, "group_vars"), 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "group_vars", "all.yml")))

	err := rejectOutsideLinks(root)
	assert.EqualError(t, err, "repository has a link that points outside of it: "+filepath.Join("group_vars", "all.yml"))
}

func TestStripServerPaths(t *testing.T) {
	got := stripServerPaths("/private/var/x/repo/hosts.yml:2: bad", "/var/x/repo/", "/private/var/x/repo/")

	assert.Equal(t, "hosts.yml:2: bad", got)
}

func TestCappedBuffer(t *testing.T) {
	var cause error
	output := cappedBuffer{limit: 4, cancel: func(err error) { cause = err }}

	_, err := output.Write([]byte("abc"))
	require.NoError(t, err)
	_, err = output.Write([]byte("de"))

	assert.EqualError(t, err, "inventory is too large")
	assert.EqualError(t, cause, "inventory is too large")
	assert.Equal(t, "abc", output.String())

	errors := cappedBuffer{limit: 4}
	_, err = errors.Write([]byte("abc"))
	require.NoError(t, err)
	_, err = errors.Write([]byte("def"))

	require.NoError(t, err)
	assert.Equal(t, "cdef", errors.String())

	lines := cappedBuffer{limit: 12}
	_, err = lines.Write([]byte("https://bob:s3cr3t@host failed\nsecond\nthird"))
	require.NoError(t, err)
	assert.Equal(t, "third", lines.String())
	_, readsFrom := any(&errors).(io.ReaderFrom)
	_, writesTo := any(&errors).(io.WriterTo)
	assert.False(t, readsFrom)
	assert.False(t, writesTo)
}

func TestRunInventoryCommand_OutputCap(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not installed")
	}
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := runInventoryCommand(ctx, t.TempDir(), 1000, "sh", "-c", "while :; do echo 0123456789; done")

	assert.EqualError(t, err, "ansible-inventory: inventory is too large")
	assert.NoError(t, ctx.Err())
}

func TestRunInventoryCommand_Environment(t *testing.T) {
	if _, err := exec.LookPath("env"); err != nil {
		t.Skip("env is not installed")
	}
	t.Setenv("INVENTORY_TEST_SECRET", "hidden")
	previousConfig := util.Config
	util.Config = &util.ConfigType{EnvVars: map[string]string{"ANSIBLE_FORCE_COLOR": "false", "ANSIBLE_INVENTORY_ENABLED": "script"}}
	t.Cleanup(func() { util.Config = previousConfig })

	out, err := runInventoryCommand(context.Background(), t.TempDir(), 64<<10, "env")
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	assert.Contains(t, lines, "ANSIBLE_FORCE_COLOR=false")
	assert.Contains(t, lines, "ANSIBLE_INVENTORY_ENABLED=yaml,ini")
	assert.Contains(t, lines, "ANSIBLE_INVENTORY_IGNORE=.pyc,.pyo,.swp,.bak,~,.rpm,.md,.txt,.rst,.orig,.cfg,.retry")
	assert.NotContains(t, lines, "ANSIBLE_INVENTORY_ENABLED=script")
	assert.NotContains(t, lines, "INVENTORY_TEST_SECRET=hidden")
}

func TestSanitizeSourceError(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"user and password", "https://bob:tok3n@git.example.com/x.git"},
		{"token with a slash", "https://tok/3n@git.example.com/x.git"},
		{"token with an at sign", "https://bob:tok@3n@git.example.com/x.git"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSourceError("fatal: unable to access '" + tt.url + "': 403 at /tmp/x")

			assert.Equal(t, "fatal: unable to access 'https://git.example.com/x.git': 403 at /tmp/x", got)
		})
	}
}

func TestGitInventoryFetcher_LocalRepository(t *testing.T) {
	if _, err := exec.LookPath("ansible-inventory"); err != nil {
		t.Skip("ansible-inventory is not installed")
	}

	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "inv"})
	require.NoError(t, err)

	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hosts.yml"),
		[]byte("all:\n  vars:\n    ntp: pool\n  children:\n    web:\n      hosts:\n        web-01:\n          port: 80\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "script.sh"),
		[]byte("#!/bin/sh\ntouch '"+marker+"'\necho '{}'\n"), 0o755))

	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{ProjectID: project.ID, Name: "local", GitURL: dir, SSHKeyID: key.ID})
	require.NoError(t, err)

	fetcher := NewGitInventoryFetcher(store, nil, nil)
	source := db.InventorySource{ProjectID: project.ID, Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"}}

	doc, err := fetcher.Fetch(context.Background(), source)
	require.NoError(t, err)
	require.Len(t, doc.Hosts, 1)
	assert.Equal(t, "web-01", doc.Hosts[0].Name)
	assert.Equal(t, map[string]any{"ntp": "pool"}, doc.Variables)
	assert.DirExists(t, dir)

	source.Params = db.MapStringAnyField{"path": "script.sh"}
	_, err = fetcher.Fetch(context.Background(), source)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), dir)
	assert.Contains(t, err.Error(), "script.sh")
	assert.NoFileExists(t, marker)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "inv", "group_vars"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "inv", "a.yml"),
		[]byte("all:\n  hosts:\n    db-01:\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "inv", "group_vars", "all.yml"), []byte("ntp: pool\n"), 0o644))

	source.Params = db.MapStringAnyField{"path": "inv"}
	doc, err = fetcher.Fetch(context.Background(), source)
	require.NoError(t, err)
	require.Len(t, doc.Hosts, 1)
	assert.Equal(t, "db-01", doc.Hosts[0].Name)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "inv", "b.yml"), []byte("all: [\n"), 0o644))
	_, err = fetcher.Fetch(context.Background(), source)
	assert.ErrorContains(t, err, "b.yml")

	secret := filepath.Join(t.TempDir(), "secret.yml")
	require.NoError(t, os.WriteFile(secret, []byte("token: s3cr3t\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "group_vars"), 0o755))
	require.NoError(t, os.Symlink(secret, filepath.Join(dir, "group_vars", "all.yml")))

	source.Params = db.MapStringAnyField{"path": "hosts.yml"}
	_, err = fetcher.Fetch(context.Background(), source)
	require.ErrorContains(t, err, "points outside")
	assert.NotContains(t, err.Error(), "s3cr3t")
}

func TestGitInventoryFetcher_MissingLocalRepository(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "inv"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	dir := filepath.Join(t.TempDir(), "missing")
	repo, err := store.CreateRepository(db.Repository{ProjectID: project.ID, Name: "local", GitURL: dir, SSHKeyID: key.ID})
	require.NoError(t, err)

	_, err = NewGitInventoryFetcher(store, nil, nil).Fetch(context.Background(), db.InventorySource{
		ProjectID: project.ID, Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), filepath.Dir(dir))
}

func TestGitInventoryFetcher_CloneFailure(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "inv"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID, Name: "remote", GitURL: "https://bob:s3cr3t@127.0.0.1:9/x.git", GitBranch: "main", SSHKeyID: key.ID,
	})
	require.NoError(t, err)
	encryption := NewAccessKeyEncryptionService(store, store, store, store)
	fetcher := NewGitInventoryFetcher(store, NewAccessKeyInstallationService(encryption), encryption)
	hook := logtest.NewGlobal()
	defer hook.Reset()

	_, err = fetcher.Fetch(context.Background(), db.InventorySource{
		ID: 3, ProjectID: project.ID, Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to access")
	assert.NotContains(t, err.Error(), "s3cr3t")
	assert.NotContains(t, err.Error(), "bob:")
	entry := hook.LastEntry()
	require.NotNil(t, entry)
	assert.Equal(t, 3, entry.Data["source_id"])
	line, err := entry.String()
	require.NoError(t, err)
	assert.NotContains(t, line, "s3cr3t")
}

func TestGitInventoryFetcher_WaitForABusyCloneEnds(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "inv"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID, Name: "remote", GitURL: "https://127.0.0.1:9/x.git", GitBranch: "main", SSHKeyID: key.ID,
	})
	require.NoError(t, err)
	encryption := NewAccessKeyEncryptionService(store, store, store, store)
	fetcher := NewGitInventoryFetcher(store, NewAccessKeyInstallationService(encryption), encryption)
	busy := make(chan struct{}, 1)
	busy <- struct{}{}
	fetcher.cloneLocks.Store("inventory_source_3", busy)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := fetcher.Fetch(ctx, db.InventorySource{
			ID: 3, ProjectID: project.ID, Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
		})
		done <- err
	}()

	select {
	case err = <-done:
		assert.EqualError(t, err, "another sync of the source is still running")
	case <-time.After(5 * time.Second):
		t.Fatal("the fetch still waits for the clone lock")
	}
}

func TestRunInventoryCommand_ErrorHasNoCredentialsCutByTheLimit(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not installed")
	}
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })
	script := "printf 'clone of https://bob:s3cr3t@git.example.com/x.git failed %0462d' 0 >&2; exit 1"

	_, err := runInventoryCommand(context.Background(), t.TempDir(), 1000, "sh", "-c", script)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cr3t")
}
