package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queuedFetcher returns the queued documents one per sync
type queuedFetcher struct {
	docs chan db.InventoryDocument
}

func (f queuedFetcher) Fetch(ctx context.Context, source db.InventorySource) (db.InventoryDocument, error) {
	return <-f.docs, nil
}

func sourceCall(f resourceFixture, c *InventorySourceController, inventory db.Inventory, method, body string, sourceID int, h func(*InventorySourceController, http.ResponseWriter, *http.Request)) (*httptest.ResponseRecorder, *audittest.Recorder) {
	vars := map[string]string{"inventory_id": strconv.Itoa(inventory.ID)}
	if sourceID != 0 {
		vars["inventory_source_id"] = strconv.Itoa(sourceID)
	}
	r, rec := f.request(method, body, map[string]any{"inventory": inventory}, vars)
	w := httptest.NewRecorder()
	h(c, w, r)
	return w, rec
}

func newSourceRepository(t *testing.T, f resourceFixture) db.Repository {
	t.Helper()
	repo, err := f.store.CreateRepository(db.Repository{ProjectID: f.project.ID, SSHKeyID: f.key.ID, Name: "hosts", GitURL: "git@example.com:hosts", GitBranch: "main"})
	require.NoError(t, err)
	return repo
}

func TestInventorySources_CreateListUpdateDelete(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "prod")
	repo := newSourceRepository(t, f)
	c := NewInventorySourceController(server.NewInventorySourceService(f.store, queuedFetcher{}, nil))
	body := `{"name":"static","type":"git","repository_id":` + strconv.Itoa(repo.ID) + `,"params":{"path":"hosts.yml"}}`

	w, rec := sourceCall(f, c, inventory, http.MethodPost, body, 0, (*InventorySourceController).AddInventorySource)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventorySource
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, "static", created.Name)
	assert.Equal(t, inventory.ID, created.InventoryID)
	got := only(t, rec, audit.ResourceInventorySourceCreate)
	assert.Equal(t, audit.ResourceTarget(audit.TargetInventorySource, created.ID, "static"), got.Event.Target)
	assert.Equal(t, audit.InventorySourceMetadata{InventoryID: inventory.ID, InventoryName: "prod", Type: "git", Repository: "hosts", Params: map[string]any{"path": "hosts.yml"}}, got.Event.Metadata)

	w, _ = sourceCall(f, c, inventory, http.MethodGet, "", 0, (*InventorySourceController).GetInventorySources)
	require.Equal(t, http.StatusOK, w.Code)
	var sources []db.InventorySource
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sources))
	require.Len(t, sources, 1)
	assert.Equal(t, created.ID, sources[0].ID)

	body = `{"name":"static2","type":"other","repository_id":` + strconv.Itoa(repo.ID) + `,"params":{"path":"inv/hosts.yml"}}`
	w, rec = sourceCall(f, c, inventory, http.MethodPut, body, created.ID, (*InventorySourceController).UpdateInventorySource)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got = only(t, rec, audit.ResourceInventorySourceUpdate)
	assert.Equal(t, audit.InventorySourceMetadata{InventoryID: inventory.ID, InventoryName: "prod", Type: "git", Repository: "hosts", Params: map[string]any{"path": "inv/hosts.yml"}}, got.Event.Metadata)

	w, _ = sourceCall(f, c, inventory, http.MethodGet, "", created.ID, (*InventorySourceController).GetInventorySource)
	require.Equal(t, http.StatusOK, w.Code)
	var source db.InventorySource
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &source))
	assert.Equal(t, "static2", source.Name)
	assert.Equal(t, db.InventorySourceGit, source.Type)

	started := time.Now().UTC().Truncate(time.Second)
	run, err := f.store.StartInventorySourceSync(f.project.ID, inventory.ID, created.ID, started, started)
	require.NoError(t, err)
	_, err = f.store.ApplyInventoryDocument(f.project.ID, inventory.ID, created.ID, run, db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1"}}},
	})
	require.NoError(t, err)
	w, rec = sourceCall(f, c, inventory, http.MethodDelete, "", created.ID, (*InventorySourceController).RemoveInventorySource)
	require.Equal(t, http.StatusNoContent, w.Code)
	got = only(t, rec, audit.ResourceInventorySourceDelete)
	assert.Equal(t, audit.ResourceTarget(audit.TargetInventorySource, created.ID, "static2"), got.Event.Target)
	assert.Equal(t, audit.InventorySourceMetadata{
		InventoryID: inventory.ID, InventoryName: "prod", Type: "git", Repository: "hosts", Params: map[string]any{"path": "inv/hosts.yml"}, Hosts: 2, Groups: 1,
	}, got.Event.Metadata)
	_, err = f.store.GetInventorySource(f.project.ID, inventory.ID, created.ID)
	require.ErrorIs(t, err, db.ErrNotFound)

	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	otherKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &other.ID})
	require.NoError(t, err)
	foreign, err := f.store.CreateRepository(db.Repository{ProjectID: other.ID, SSHKeyID: otherKey.ID, Name: "foreign", GitURL: "git@example.com:f", GitBranch: "main"})
	require.NoError(t, err)
	w, rec = sourceCall(f, c, inventory, http.MethodPost, `{"name":"a","type":"git","repository_id":`+strconv.Itoa(foreign.ID)+`,"params":{"path":"hosts.yml"}}`, 0, (*InventorySourceController).AddInventorySource)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, rec.All())

	viewer := f
	viewer.user.Admin = false
	r, rec := viewer.request(http.MethodPost, body, map[string]any{"inventory": inventory, "permissions": db.CanReadProject}, map[string]string{"inventory_id": strconv.Itoa(inventory.ID)})
	w = httptest.NewRecorder()
	GetMustCanMiddleware(db.CanManageProjectResources)(http.HandlerFunc(c.AddInventorySource)).ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
	only(t, rec, audit.AuthAuthorizationDeny)
}

func TestInventorySources_GitNeedsPath(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "prod")
	repo := newSourceRepository(t, f)
	c := NewInventorySourceController(server.NewInventorySourceService(f.store, queuedFetcher{}, nil))
	body := `{"name":"static","type":"git","repository_id":` + strconv.Itoa(repo.ID)

	for _, path := range []string{``, `,"params":{"path":"  "}`} {
		w, rec := sourceCall(f, c, inventory, http.MethodPost, body+path+`}`, 0, (*InventorySourceController).AddInventorySource)

		assert.Equal(t, http.StatusBadRequest, w.Code, path)
		assert.JSONEq(t, `{"error":"source path is required"}`, w.Body.String(), path)
		assert.Empty(t, rec.All(), path)
	}

	w, _ := sourceCall(f, c, inventory, http.MethodPost, body+`,"params":{"path":" inventory/prod.yml "}}`, 0, (*InventorySourceController).AddInventorySource)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventorySource
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	got, err := f.store.GetInventorySource(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "inventory/prod.yml", got.Params["path"])

	w, _ = sourceCall(f, c, inventory, http.MethodPut, body+`,"params":{"path":"  "}}`, created.ID, (*InventorySourceController).UpdateInventorySource)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventorySources_BodyCannotSetState(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "prod")
	repo := newSourceRepository(t, f)
	c := NewInventorySourceController(server.NewInventorySourceService(f.store, queuedFetcher{}, nil))
	body := `{"name":"static","type":"git","repository_id":` + strconv.Itoa(repo.ID) + `,"params":{"path":"hosts.yml"},"status":"success","variables":"a: 1","last_error":"boom"}`

	w, _ := sourceCall(f, c, inventory, http.MethodPost, body, 0, (*InventorySourceController).AddInventorySource)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventorySource
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	got, err := f.store.GetInventorySource(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Status)
	assert.Empty(t, got.Variables)
	assert.Empty(t, got.LastError)
}

func TestInventorySources_SyncRecordsAudit(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "prod")
	repo := newSourceRepository(t, f)
	one := db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web1"}}}
	two := db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}}}
	fetcher := queuedFetcher{docs: make(chan db.InventoryDocument, 4)}
	c := NewInventorySourceController(server.NewInventorySourceService(f.store, fetcher, nil))
	source, err := f.store.CreateInventorySource(db.InventorySource{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "static", Type: db.InventorySourceGit, RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"}})
	require.NoError(t, err)
	sync := func() *audittest.Recorder {
		w, rec := sourceCall(f, c, inventory, http.MethodPost, "", source.ID, (*InventorySourceController).SyncInventorySource)
		require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
		return rec
	}
	recorded := func(rec *audittest.Recorder) bool { return len(rec.All()) > 0 }
	finished := func() bool {
		got, err := f.store.GetInventorySource(f.project.ID, inventory.ID, source.ID)
		return err == nil && got.Status != db.InventorySourceSyncing
	}

	// The fetcher waits for a document, so the first sync is still running
	first := sync()
	w, _ := sourceCall(f, c, inventory, http.MethodPost, "", source.ID, (*InventorySourceController).SyncInventorySource)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"source is already syncing"}`, w.Body.String())
	fetcher.docs <- one
	fetcher.docs <- one
	fetcher.docs <- two
	require.Eventually(t, func() bool { return recorded(first) }, 5*time.Second, 10*time.Millisecond)
	got := only(t, first, audit.ResourceInventorySourceSync)
	assert.Equal(t, audit.ResourceTarget(audit.TargetInventorySource, source.ID, "static"), got.Event.Target)
	assert.Equal(t, f.project.ID, got.Event.ProjectID)
	assert.Equal(t, audit.InventorySourceSyncMetadata{InventoryID: inventory.ID, InventoryName: "prod", HostsAdded: 1}, got.Event.Metadata)

	// A third sync with changes ends after the second, so the second has called done by then
	second := sync()
	require.Eventually(t, finished, 5*time.Second, 10*time.Millisecond)
	third := sync()
	require.Eventually(t, func() bool { return recorded(third) }, 5*time.Second, 10*time.Millisecond)
	assert.Empty(t, second.All())
	assert.Equal(t, audit.InventorySourceSyncMetadata{InventoryID: inventory.ID, InventoryName: "prod", HostsAdded: 1}, only(t, third, audit.ResourceInventorySourceSync).Event.Metadata)

	// A sync that changes only the variables of the source is recorded too
	two.Variables = map[string]any{"env": "prod"}
	fetcher.docs <- two
	fourth := sync()
	require.Eventually(t, func() bool { return recorded(fourth) }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, audit.InventorySourceSyncMetadata{InventoryID: inventory.ID, InventoryName: "prod", VariablesChanged: true}, only(t, fourth, audit.ResourceInventorySourceSync).Event.Metadata)
}
