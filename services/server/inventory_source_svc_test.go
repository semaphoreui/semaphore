package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInventoryFetcher struct {
	doc     db.InventoryDocument
	err     error
	release chan struct{}
	path    atomic.Value
	entered atomic.Int32
	late    bool
}

func (f *fakeInventoryFetcher) Fetch(ctx context.Context, source db.InventorySource) (db.InventoryDocument, error) {
	f.path.Store(source.Params["path"])
	f.entered.Add(1)
	if f.release != nil {
		<-f.release
	}
	if f.late {
		<-ctx.Done()
	}
	return f.doc, f.err
}

func createTestInventorySource(t *testing.T) (*sql.SqlDb, db.InventorySource) {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "inv"})
	require.NoError(t, err)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "managed", Type: db.InventoryManaged})
	require.NoError(t, err)
	source, err := store.CreateInventorySource(db.InventorySource{
		ProjectID: project.ID, InventoryID: inventory.ID, Name: "git", Type: db.InventorySourceGit, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})
	require.NoError(t, err)
	return store, source
}

func syncAndWait(t *testing.T, svc *InventorySourceService, source db.InventorySource) error {
	t.Helper()
	done := make(chan error, 1)
	started, err := svc.Sync(source, func(_ db.InventorySyncResult, err error) { done <- err })
	require.NoError(t, err)
	require.True(t, started)
	return <-done
}

func TestInventorySourceService_SyncImports(t *testing.T) {
	store, source := createTestInventorySource(t)
	svc := NewInventorySourceService(store, &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}}, nil)

	err := syncAndWait(t, svc, source)

	require.NoError(t, err)
	hosts, _, err := store.GetInventoryHosts(source.ProjectID, source.InventoryID, db.InventoryHostListParams{SourceID: source.ID})
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, "web-01", hosts[0].Name)
	got, err := store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, db.InventorySourceSuccess, got.Status)
	require.NotNil(t, got.LastResult)
	assert.Equal(t, 1, got.LastResult.HostsAdded)
}

func TestInventorySourceService_FailureKeepsHosts(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}}
	svc := NewInventorySourceService(store, fetcher, nil)
	require.NoError(t, syncAndWait(t, svc, source))
	before, err := store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	require.NoError(t, err)
	fetcher.err = errors.New("repository is not reachable")

	err = syncAndWait(t, svc, source)

	assert.EqualError(t, err, "repository is not reachable")
	hosts, _, err := store.GetInventoryHosts(source.ProjectID, source.InventoryID, db.InventoryHostListParams{SourceID: source.ID})
	require.NoError(t, err)
	assert.Len(t, hosts, 1)
	got, err := store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, db.InventorySourceFailed, got.Status)
	assert.Equal(t, "repository is not reachable", got.LastError)
	assert.Equal(t, before.LastSuccess, got.LastSuccess)
}

func TestInventorySourceService_StoredErrorHasNoCredentials(t *testing.T) {
	store, source := createTestInventorySource(t)
	svc := NewInventorySourceService(store, &fakeInventoryFetcher{err: errors.New("unable to access 'https://u:p@h/x': 403")}, nil)

	_ = syncAndWait(t, svc, source)

	got, err := store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, "unable to access 'https://h/x': 403", got.LastError)
}

func TestInventorySourceService_EmptyIsFailure(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}}
	svc := NewInventorySourceService(store, fetcher, nil)
	require.NoError(t, syncAndWait(t, svc, source))
	fetcher.doc = db.InventoryDocument{}

	err := syncAndWait(t, svc, source)

	assert.ErrorIs(t, err, db.ErrInventorySourceEmpty)
	hosts, _, err := store.GetInventoryHosts(source.ProjectID, source.InventoryID, db.InventoryHostListParams{SourceID: source.ID})
	require.NoError(t, err)
	assert.Len(t, hosts, 1)
	got, err := store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, "source returned no hosts", got.LastError)
}

func TestInventorySourceService_SecondSyncRefused(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}, release: make(chan struct{})}
	svc := NewInventorySourceService(store, fetcher, nil)
	done := make(chan error, 1)
	release := sync.OnceFunc(func() { close(fetcher.release) })
	t.Cleanup(release)
	started, err := svc.Sync(source, func(_ db.InventorySyncResult, err error) { done <- err })
	require.NoError(t, err)
	require.True(t, started)

	started, err = svc.Sync(source, nil)

	require.NoError(t, err)
	assert.False(t, started)
	release()
	assert.NoError(t, <-done)
}

func TestInventorySourceService_PresentStale(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	svc := NewInventorySourceService(nil, nil, nil)
	svc.now = func() time.Time { return now }
	stale := now.Add(-31 * time.Minute)
	recent := now.Add(-time.Minute)

	got := svc.Present(db.InventorySource{Status: db.InventorySourceSyncing, SyncStarted: &stale})
	running := svc.Present(db.InventorySource{Status: db.InventorySourceSyncing, SyncStarted: &recent})

	assert.Equal(t, db.InventorySourceFailed, got.Status)
	assert.Equal(t, "sync was interrupted", got.LastError)
	assert.Equal(t, db.InventorySourceSyncing, running.Status)
	assert.Empty(t, running.LastError)
}

func TestInventorySourceService_GitNeedsPath(t *testing.T) {
	svc := NewInventorySourceService(nil, nil, nil)
	repository := 5

	err := svc.Validate(db.InventorySource{Name: "git", Type: db.InventorySourceGit, RepositoryID: &repository})
	rootErr := svc.Validate(db.InventorySource{Name: "git", Type: db.InventorySourceGit, RepositoryID: &repository, Params: db.MapStringAnyField{"path": "./"}})

	assert.EqualError(t, err, "source path is required")
	assert.EqualError(t, rootErr, "source path is required")
}

func TestInventorySourceService_UnknownType(t *testing.T) {
	svc := NewInventorySourceService(nil, nil, nil)

	err := svc.Validate(db.InventorySource{Name: "other", Type: "other"})

	assert.ErrorIs(t, err, db.ErrInventorySourceTypeUnsupported)
}

func TestInventorySourceService_SyncReadsTheSourceAgain(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}}
	svc := NewInventorySourceService(store, fetcher, nil)
	edited := source
	edited.Params = db.MapStringAnyField{"path": "other.yml"}
	require.NoError(t, store.UpdateInventorySource(edited))

	err := syncAndWait(t, svc, source)

	require.NoError(t, err)
	assert.Equal(t, "other.yml", fetcher.path.Load())
}

func TestInventorySourceService_LimitsSyncsAtOnce(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}, release: make(chan struct{})}
	svc := NewInventorySourceService(store, fetcher, nil)
	done := make(chan error, maxInventorySyncsAtOnce+1)
	release := sync.OnceFunc(func() { close(fetcher.release) })
	t.Cleanup(release)
	for i := range maxInventorySyncsAtOnce + 1 {
		source.Name = fmt.Sprintf("git-%d", i)
		created, err := store.CreateInventorySource(source)
		require.NoError(t, err)
		started, err := svc.Sync(created, func(_ db.InventorySyncResult, err error) { done <- err })
		require.NoError(t, err)
		require.True(t, started)
	}

	require.Eventually(t, func() bool { return fetcher.entered.Load() == maxInventorySyncsAtOnce }, 5*time.Second, 10*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	waiting := fetcher.entered.Load()
	release()
	for range maxInventorySyncsAtOnce + 1 {
		<-done
	}

	assert.EqualValues(t, maxInventorySyncsAtOnce, waiting)
	assert.EqualValues(t, maxInventorySyncsAtOnce+1, fetcher.entered.Load())
}

func TestInventorySourceService_LateDocumentIsNotApplied(t *testing.T) {
	store, source := createTestInventorySource(t)
	fetcher := &fakeInventoryFetcher{doc: db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web-01"}}}, late: true}
	svc := NewInventorySourceService(store, fetcher, nil)
	svc.timeout = 20 * time.Millisecond

	err := syncAndWait(t, svc, source)

	assert.EqualError(t, err, db.InventorySourceInterrupted)
	hosts, _, err := store.GetInventoryHosts(source.ProjectID, source.InventoryID, db.InventoryHostListParams{})
	require.NoError(t, err)
	assert.Empty(t, hosts)
}
