package sql

import (
	"fmt"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventorySources_CRUD(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "git", Type: db.InventorySourceGit, Params: db.MapStringAnyField{"path": "hosts.yml"},
	}

	created, err := f.store.CreateInventorySource(source)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	got, err := f.store.GetInventorySource(f.project.ID, f.inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "git", got.Name)
	assert.Equal(t, "hosts.yml", got.Params["path"])

	list, err := f.store.GetInventorySources(f.project.ID, f.inventory.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 0, list[0].HostCount)
	assert.Equal(t, 0, list[0].GroupCount)

	got.Name = "renamed"
	got.Params["path"] = "other.yml"
	require.NoError(t, f.store.UpdateInventorySource(got))
	got, err = f.store.GetInventorySource(f.project.ID, f.inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)
	assert.Equal(t, "other.yml", got.Params["path"])

	_, err = f.store.CreateInventorySource(db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "renamed", Type: db.InventorySourceGit, Params: got.Params,
	})
	assert.ErrorContains(t, err, "a source with this name already exists")

	require.NoError(t, f.store.DeleteInventorySource(f.project.ID, f.inventory.ID, created.ID))
	_, err = f.store.GetInventorySource(f.project.ID, f.inventory.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestUpdateInventoryHost_KeepsSource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	host := f.host(t, "web1")
	_, err := f.store.Sql().Exec("update project__inventory_host set source_id = ?, variables = ? where id = ?", 7, "port: 22\n", host.ID)
	require.NoError(t, err)

	source := 7
	host.SourceID = &source
	host.Description = "front"
	_, _, err = f.store.UpdateInventoryHost(host, nil, nil)
	require.NoError(t, err)

	got, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, "front", got.Description)
	assert.Equal(t, "port: 22\n", got.Variables)
	require.NotNil(t, got.SourceID)
	assert.Equal(t, 7, *got.SourceID)
}

func TestUpdateInventoryGroup_KeepsSource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	group := f.group(t, "web")
	_, err := f.store.Sql().Exec("update project__inventory_group set source_id = ?, variables = ? where id = ?", 7, "port: 22\n", group.ID)
	require.NoError(t, err)

	source := 7
	group.SourceID = &source
	group.Description = "front"
	_, _, err = f.store.UpdateInventoryGroup(group, nil, nil)
	require.NoError(t, err)

	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "front", got.Description)
	assert.Equal(t, "port: 22\n", got.Variables)
	require.NotNil(t, got.SourceID)
	assert.Equal(t, 7, *got.SourceID)
}

func TestUpdateInventoryHost_OwnerChanged(t *testing.T) {
	f := newInventoryGroupFixture(t)
	host := f.host(t, "web1")
	_, err := f.store.Sql().Exec("update project__inventory_host set source_id = ? where id = ?", 7, host.ID)
	require.NoError(t, err)

	host.Name = "web2"
	_, _, err = f.store.UpdateInventoryHost(host, nil, nil)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)
	_, _, err = f.store.UpdateInventoryHost(host, []int{f.group(t, "web").ID}, nil)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)

	got, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, "web1", got.Name)
	assert.Empty(t, got.Groups)
}

func TestDeleteInventoryHost_OwnerChanged(t *testing.T) {
	f := newInventoryGroupFixture(t)
	host := f.host(t, "web1")
	_, err := f.store.Sql().Exec("update project__inventory_host set source_id = ? where id = ?", 7, host.ID)
	require.NoError(t, err)

	err = f.store.DeleteInventoryHost(f.project.ID, f.inventory.ID, host.ID)

	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)
	_, err = f.store.GetInventoryHost(f.project.ID, f.inventory.ID, host.ID)
	assert.NoError(t, err)
}

func TestUpdateInventoryGroup_OwnerChanged(t *testing.T) {
	f := newInventoryGroupFixture(t)
	group := f.group(t, "web")
	_, err := f.store.Sql().Exec("update project__inventory_group set source_id = ? where id = ?", 7, group.ID)
	require.NoError(t, err)

	group.Name = "app"
	_, _, err = f.store.UpdateInventoryGroup(group, nil, nil)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)
	err = f.store.DeleteInventoryGroup(f.project.ID, f.inventory.ID, group.ID)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)

	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "web", got.Name)
}

func TestGetInventoryHosts_FiltersBySource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := 7
	owned, err := f.store.CreateInventoryHost(db.InventoryHost{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "owned", Enabled: true, SourceID: &source,
	}, nil)
	require.NoError(t, err)
	manual := f.host(t, "manual")

	hosts, total, err := f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{SourceID: source})
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, owned.ID, hosts[0].ID)

	hosts, total, err = f.store.GetInventoryHosts(f.project.ID, f.inventory.ID, db.InventoryHostListParams{Manual: true})
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, manual.ID, hosts[0].ID)
}

func (f inventoryGroupFixture) source(t *testing.T) db.InventorySource {
	t.Helper()
	source, err := f.store.CreateInventorySource(db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "git", Type: db.InventorySourceGit, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})
	require.NoError(t, err)
	return source
}

// apply starts a sync of the source and applies doc as that sync
func (f inventoryGroupFixture) apply(sourceID int, doc db.InventoryDocument) (db.InventorySyncResult, error) {
	started := tz.Now()
	run, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, sourceID, started, started.Add(time.Hour))
	if err != nil {
		return db.InventorySyncResult{}, err
	}
	return f.store.ApplyInventoryDocument(f.project.ID, f.inventory.ID, sourceID, run, doc)
}

// inventoryRows counts hosts, groups, group hosts and group children
type inventoryRows struct{ Hosts, Groups, Links, Children int64 }

func (f inventoryGroupFixture) rows(t *testing.T) inventoryRows {
	t.Helper()
	var rows inventoryRows
	err := f.store.Sql().SelectOne(&rows, `select
		(select count(*) from project__inventory_host) as Hosts,
		(select count(*) from project__inventory_group) as Groups,
		(select count(*) from project__inventory_group_host) as Links,
		(select count(*) from project__inventory_group_child) as Children`)
	require.NoError(t, err)
	return rows
}

func (f inventoryGroupFixture) hostID(t *testing.T, name string) int {
	t.Helper()
	id, err := f.store.Sql().SelectInt("select id from project__inventory_host where name = ?", name)
	require.NoError(t, err)
	return int(id)
}

func (f inventoryGroupFixture) groupID(t *testing.T, name string) int {
	t.Helper()
	id, err := f.store.Sql().SelectInt("select id from project__inventory_group where name = ?", name)
	require.NoError(t, err)
	return int(id)
}

func TestApplyInventoryDocument_ImportsAndRepeats(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	doc := db.InventoryDocument{
		Variables: map[string]any{"env": "prod"},
		Hosts: []db.InventoryDocumentHost{
			{Name: "web1", Variables: map[string]any{"port": 22}}, {Name: "web2"}, {Name: "db1"},
		},
		Groups: []db.InventoryDocumentGroup{
			{Name: "web", Hosts: []string{"web1", "web2"}},
			{Name: "servers", Hosts: []string{"db1"}, Children: []string{"web"}},
		},
	}

	result, err := f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.Equal(t, 3, result.HostsAdded)
	assert.Equal(t, 2, result.GroupsAdded)
	assert.Equal(t, inventoryRows{Hosts: 3, Groups: 2, Links: 3, Children: 1}, f.rows(t))

	web1, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, f.hostID(t, "web1"))
	require.NoError(t, err)
	assert.Equal(t, "port: 22\n", web1.Variables)
	assert.Equal(t, 1, web1.VariableCount)
	require.NotNil(t, web1.SourceID)
	assert.Equal(t, source.ID, *web1.SourceID)
	got, err := f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, "env: prod\n", got.Variables)
	assert.Equal(t, 1, got.VariableCount)

	result, err = f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.False(t, result.Changed())
	assert.Equal(t, inventoryRows{Hosts: 3, Groups: 2, Links: 3, Children: 1}, f.rows(t))
}

func TestApplyInventoryDocument_VariablesChanged(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	doc := db.InventoryDocument{Variables: map[string]any{"env": "prod"}, Hosts: []db.InventoryDocumentHost{{Name: "web1"}}}
	_, err := f.apply(source.ID, doc)
	require.NoError(t, err)

	doc.Variables = map[string]any{"env": "dev"}
	result, err := f.apply(source.ID, doc)

	require.NoError(t, err)
	assert.Equal(t, db.InventorySyncResult{VariablesChanged: true}, result)
	assert.True(t, result.Changed())
}

func TestApplyInventoryDocument_UpdatesChanged(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	doc := db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1", "web2"}}},
	}
	_, err := f.apply(source.ID, doc)
	require.NoError(t, err)

	doc.Hosts[0].Variables = map[string]any{"port": 2222}
	doc.Groups[0].Hosts = []string{"web1"}
	result, err := f.apply(source.ID, doc)

	require.NoError(t, err)
	assert.Equal(t, 1, result.HostsChanged)
	assert.Equal(t, 1, result.GroupsChanged)
	web1, err := f.store.GetInventoryHost(f.project.ID, f.inventory.ID, f.hostID(t, "web1"))
	require.NoError(t, err)
	assert.Equal(t, "port: 2222\n", web1.Variables)
	links, err := f.store.GetInventoryGroupHosts(f.project.ID, f.inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.InventoryGroupHost{{GroupID: f.groupID(t, "web"), HostID: web1.ID}}, links)
}

func TestApplyInventoryDocument_KeepsManualLinks(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	manual := f.group(t, "manual")
	doc := db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web2"}}},
	}
	_, err := f.apply(source.ID, doc)
	require.NoError(t, err)
	web1 := f.hostID(t, "web1")
	web := f.groupID(t, "web")
	_, _, err = f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{web1}, []int{manual.ID}, nil, false)
	require.NoError(t, err)
	_, _, err = f.store.UpdateInventoryGroup(manual, []int{web}, nil)
	require.NoError(t, err)

	_, err = f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.Equal(t, inventoryRows{Hosts: 2, Groups: 2, Links: 2, Children: 1}, f.rows(t))

	doc = db.InventoryDocument{Hosts: []db.InventoryDocumentHost{{Name: "web2"}}}
	result, err := f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.Equal(t, 1, result.HostsRemoved)
	assert.Equal(t, 1, result.GroupsRemoved)
	assert.Equal(t, inventoryRows{Hosts: 1, Groups: 1}, f.rows(t))
}

func TestApplyInventoryDocument_FailureChangesNothing(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	_, err := f.apply(source.ID, db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1"}}},
	})
	require.NoError(t, err)

	_, err = f.apply(source.ID, db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "db1"}},
		Groups: []db.InventoryDocumentGroup{{Name: "a", Children: []string{"b"}}, {Name: "b", Children: []string{"a"}}},
	})

	assert.ErrorContains(t, err, "group cycle")
	assert.Equal(t, inventoryRows{Hosts: 2, Groups: 1, Links: 1}, f.rows(t))
}

func TestApplyInventoryDocument_RollsBackAfterWrites(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	_, err := f.apply(source.ID, db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1"}}},
	})
	require.NoError(t, err)
	links, err := f.store.GetInventoryGroupHosts(f.project.ID, f.inventory.ID)
	require.NoError(t, err)

	_, err = f.apply(source.ID, db.InventoryDocument{
		Hosts: []db.InventoryDocumentHost{{Name: "db1"}, {Name: "db1"}},
	})

	require.Error(t, err)
	assert.Equal(t, inventoryRows{Hosts: 2, Groups: 1, Links: 1}, f.rows(t))
	after, err := f.store.GetInventoryGroupHosts(f.project.ID, f.inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, links, after)
	assert.NotZero(t, f.hostID(t, "web2"))
}

func TestApplyInventoryDocument_ManyHosts(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	doc := db.InventoryDocument{}
	for i := range 50 {
		doc.Groups = append(doc.Groups, db.InventoryDocumentGroup{Name: fmt.Sprintf("group%d", i)})
	}
	for i := range 3000 {
		name := fmt.Sprintf("host%d", i)
		doc.Hosts = append(doc.Hosts, db.InventoryDocumentHost{Name: name, Variables: map[string]any{"index": i}})
		doc.Groups[i%50].Hosts = append(doc.Groups[i%50].Hosts, name)
	}

	result, err := f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.Equal(t, 3000, result.HostsAdded)
	assert.Equal(t, 50, result.GroupsAdded)
	assert.Equal(t, inventoryRows{Hosts: 3000, Groups: 50, Links: 3000}, f.rows(t))

	result, err = f.apply(source.ID, doc)
	require.NoError(t, err)
	assert.False(t, result.Changed())
	assert.Equal(t, inventoryRows{Hosts: 3000, Groups: 50, Links: 3000}, f.rows(t))
}

func TestDeleteInventorySource_RemovesOwnedRows(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	manual := f.group(t, "manual")
	f.host(t, "m1", manual.ID)
	_, err := f.apply(source.ID, db.InventoryDocument{
		Hosts:  []db.InventoryDocumentHost{{Name: "web1"}, {Name: "web2"}},
		Groups: []db.InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1", "web2"}}},
	})
	require.NoError(t, err)
	_, _, err = f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{f.hostID(t, "web1")}, []int{manual.ID}, nil, false)
	require.NoError(t, err)
	_, _, err = f.store.UpdateInventoryGroup(manual, []int{f.groupID(t, "web")}, nil)
	require.NoError(t, err)

	require.NoError(t, f.store.DeleteInventorySource(f.project.ID, f.inventory.ID, source.ID))

	assert.Equal(t, inventoryRows{Hosts: 1, Groups: 1, Links: 1}, f.rows(t))
	_, err = f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestStartInventorySourceSync_OnlyOnce(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	now := tz.Now()

	first, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, now, now.Add(-db.InventorySourceStaleAfter))
	require.NoError(t, err)
	later := now.Add(time.Minute)
	refused, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, later, later.Add(-db.InventorySourceStaleAfter))
	require.NoError(t, err)
	afterStale, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, later, now.Add(time.Second))
	require.NoError(t, err)

	assert.Equal(t, 1, first)
	assert.Zero(t, refused)
	assert.Equal(t, 2, afterStale)
}

func newSourceRepository(t *testing.T, f inventoryGroupFixture) db.Repository {
	t.Helper()
	key, err := f.store.CreateAccessKey(db.AccessKey{ProjectID: &f.project.ID, Name: "none", Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := f.store.CreateRepository(db.Repository{
		ProjectID: f.project.ID, Name: "repo", GitURL: "https://example.com/repo.git", GitBranch: "main", SSHKeyID: key.ID,
	})
	require.NoError(t, err)
	return repo
}

func TestDeleteRepository_UsedByInventorySource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	repo := newSourceRepository(t, f)
	source, err := f.store.CreateInventorySource(db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "git", Type: db.InventorySourceGit,
		RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})
	require.NoError(t, err)

	err = f.store.DeleteRepository(f.project.ID, repo.ID)

	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	_, err = f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	assert.NoError(t, err)
}

func TestGetRepositoryRefs_ListsInventorySources(t *testing.T) {
	f := newInventoryGroupFixture(t)
	repo := newSourceRepository(t, f)
	source, err := f.store.CreateInventorySource(db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "git", Type: db.InventorySourceGit,
		RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})
	require.NoError(t, err)

	refs, err := f.store.GetRepositoryRefs(f.project.ID, repo.ID)

	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: source.ID, Name: "git"}}, refs.InventorySources)

	inventoryRefs, err := f.store.GetInventoryRefs(f.project.ID, f.inventory.ID)

	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{}, inventoryRefs.InventorySources)
}

func TestDeleteProject_WithInventorySource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	repo := newSourceRepository(t, f)
	source, err := f.store.CreateInventorySource(db.InventorySource{
		ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "git", Type: db.InventorySourceGit,
		RepositoryID: &repo.ID, Params: db.MapStringAnyField{"path": "hosts.yml"},
	})
	require.NoError(t, err)
	_, err = f.apply(source.ID, db.InventoryDocument{
		Hosts: []db.InventoryDocumentHost{{Name: "web1"}},
	})
	require.NoError(t, err)

	err = f.store.DeleteProject(f.project.ID)

	require.NoError(t, err)
	assert.Equal(t, inventoryRows{}, f.rows(t))
	sources, err := f.store.Sql().SelectInt("select count(*) from project__inventory_source")
	require.NoError(t, err)
	assert.Zero(t, sources)
	_, err = f.store.GetRepository(f.project.ID, repo.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestHostGroupLinks_GroupTakenBySource(t *testing.T) {
	f := newInventoryGroupFixture(t)
	host := f.host(t, "web1")
	group := f.group(t, "web")
	_, err := f.store.Sql().Exec("update project__inventory_group set source_id = ? where id = ?", 7, group.ID)
	require.NoError(t, err)

	_, _, err = f.store.ChangeInventoryHostGroups(f.project.ID, f.inventory.ID, []int{host.ID}, []int{group.ID}, nil, true)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)
	_, _, err = f.store.UpdateInventoryHost(host, []int{group.ID}, nil)
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)
	_, err = f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: f.inventory.ID, Name: "web2"}, []int{group.ID})
	assert.ErrorIs(t, err, db.ErrInventoryRowOwnerChanged)

	links, err := f.store.GetInventoryGroupHosts(f.project.ID, f.inventory.ID)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestInventorySourceSync_ReplacedRunWritesNothing(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	now := tz.Now()
	first, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, now, now)
	require.NoError(t, err)
	second, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, now, now.Add(time.Second))
	require.NoError(t, err)
	require.NotZero(t, second)

	_, err = f.store.ApplyInventoryDocument(f.project.ID, f.inventory.ID, source.ID, first, db.InventoryDocument{
		Hosts: []db.InventoryDocumentHost{{Name: "old"}},
	})
	assert.ErrorIs(t, err, db.ErrNotFound)
	require.NoError(t, f.store.FailInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, first, now, "late"))

	got, err := f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, db.InventorySourceSyncing, got.Status)
	assert.Empty(t, got.LastError)
	assert.Zero(t, f.rows(t).Hosts)

	result, err := f.store.ApplyInventoryDocument(f.project.ID, f.inventory.ID, source.ID, second, db.InventoryDocument{
		Hosts: []db.InventoryDocumentHost{{Name: "new"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.HostsAdded)
	got, err = f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, db.InventorySourceSuccess, got.Status)
}

func TestUpdateInventorySource_RefusedDuringASync(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	now := tz.Now()
	_, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, now, now)
	require.NoError(t, err)
	source.Params = db.MapStringAnyField{"path": "other.yml"}

	err = f.store.UpdateInventorySource(source)

	assert.ErrorIs(t, err, db.ErrInventorySourceSyncing)
	got, err := f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, "hosts.yml", got.Params["path"])
}

func TestUpdateInventorySource_ClosesAnInterruptedSync(t *testing.T) {
	f := newInventoryGroupFixture(t)
	source := f.source(t)
	long := tz.Now().Add(-time.Hour)
	run, err := f.store.StartInventorySourceSync(f.project.ID, f.inventory.ID, source.ID, long, long)
	require.NoError(t, err)
	source.Params = db.MapStringAnyField{"path": "other.yml"}

	require.NoError(t, f.store.UpdateInventorySource(source))
	_, err = f.store.ApplyInventoryDocument(f.project.ID, f.inventory.ID, source.ID, run, db.InventoryDocument{
		Hosts: []db.InventoryDocumentHost{{Name: "old"}},
	})

	assert.ErrorIs(t, err, db.ErrNotFound)
	got, err := f.store.GetInventorySource(f.project.ID, f.inventory.ID, source.ID)
	require.NoError(t, err)
	assert.Equal(t, "other.yml", got.Params["path"])
	assert.Equal(t, db.InventorySourceFailed, got.Status)
	assert.Equal(t, db.InventorySourceInterrupted, got.LastError)
}

func TestUpdateInventoryGroup_ImportedNameIsNotChecked(t *testing.T) {
	f := newInventoryGroupFixture(t)
	group := f.group(t, "web")
	_, err := f.store.Sql().Exec("update project__inventory_group set source_id = ?, name = ? where id = ?", 7, "app", group.ID)
	require.NoError(t, err)
	f.group(t, "web")
	source := 7
	group.SourceID = &source
	group.Description = "front"

	_, _, err = f.store.UpdateInventoryGroup(group, nil, nil)

	require.NoError(t, err)
	got, err := f.store.GetInventoryGroup(f.project.ID, f.inventory.ID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "app", got.Name)
	assert.Equal(t, "front", got.Description)
}
