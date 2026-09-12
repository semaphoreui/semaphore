package project

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db/sql"

	"github.com/semaphoreui/semaphore/db"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testItem struct {
	Name string
}

func TestBackupProject(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "Test 123",
	})
	assert.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	assert.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	assert.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	assert.NoError(t, err)

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	assert.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Name:                  "Test",
		Playbook:              "test.yml",
		ProjectID:             proj.ID,
		RepositoryID:          repo.ID,
		InventoryID:           &inv.ID,
		EnvironmentIDs:        []int{env.ID},
		SuppressSuccessAlerts: true,
		SuppressErrorAlerts:   true,
	})
	assert.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	assert.NoError(t, err)
	assert.Equal(t, proj.ID, backup.Meta.ID)

	str, err := backup.Marshal()
	assert.NoError(t, err)

	restoredBackup := &BackupFormat{}
	err = restoredBackup.Unmarshal(str)
	assert.NoError(t, err)
	assert.Equal(t, restoredBackup.Meta.Name, "Test 123")

	restoredBackup.Meta.Name = "Test 1234"

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "test",
			Name:     "Test",
			Email:    "test@example.com",
			Admin:    true,
		},
	})
	assert.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	assert.NoError(t, err)
	assert.Equal(t, restoredProj.Name, "Test 1234")

	restoredTemplates, err := store.GetTemplates(restoredProj.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	assert.NoError(t, err)
	assert.Len(t, restoredTemplates, 1)
	assert.Len(t, restoredTemplates[0].EnvironmentIDs, 1)
	assert.True(t, restoredTemplates[0].SuppressSuccessAlerts)
	assert.True(t, restoredTemplates[0].SuppressErrorAlerts)

	restoredEnvs, err := store.GetEnvironments(restoredProj.ID, db.RetrieveQueryParams{})
	assert.NoError(t, err)
	assert.Len(t, restoredEnvs, 1)
	assert.Equal(t, restoredEnvs[0].ID, restoredTemplates[0].EnvironmentIDs[0])
	assert.Equal(t, "test", restoredEnvs[0].Name)
}

func TestBackup_BackupSecretStorage(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "Test 123",
	})
	assert.NoError(t, err)

	storage, err := store.CreateSecretStorage(db.SecretStorage{
		ProjectID: proj.ID,
		Type:      "vault",
		Name:      "Test",
	})
	assert.NoError(t, err)

	_, err = store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
		StorageID: &storage.ID,
		Name:      "Test Key",
		Owner:     "vault",
	})
	assert.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	assert.NoError(t, err)
	assert.Equal(t, proj.ID, backup.Meta.ID)
	backup.Meta.Name = "Test 1234"

	str, err := backup.Marshal()
	assert.NoError(t, err)

	var res map[string]any
	if err := json.Unmarshal([]byte(str), &res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `{
  "environments": [],
  "integration_aliases": [],
  "integrations": [],
  "inventories": [],
  "keys": [
    {
      "name": "Test Key",
      "owner": "vault",
      "storage": "Test",
      "synchronized": false,
      "type": "none"
    }
  ],
  "meta": {
    "alert": false,
    "max_parallel_tasks": 0,
    "name": "Test 1234",
    "type": ""
  },
  "proxies": [],
  "repositories": [],
  "roles": [],
  "runners": [],
  "schedules": [],
  "secret_storages": [
    {
      "name": "Test",
      "params": {},
      "readonly": false,
      "sync_enabled": false,
      "sync_interval": 0,
      "type": "vault"
    }
  ],
  "templates": [],
  "views": [],
  "workflows": []
}`, str)

	restoredBackup := &BackupFormat{}
	err = restoredBackup.Unmarshal(str)
	assert.NoError(t, err)
	assert.Equal(t, restoredBackup.Meta.Name, "Test 1234")

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "test",
			Name:     "Test",
			Email:    "test@example.com",
			Admin:    true,
		},
	})
	assert.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	assert.Nil(t, err)

	restoredStorages, err := store.GetSecretStorages(restoredProj.ID)
	assert.NoError(t, err)
	assert.Len(t, restoredStorages, 1)

	restoredKeys, err := store.GetAccessKeys(restoredProj.ID, db.GetAccessKeyOptions{IgnoreOwner: true}, db.RetrieveQueryParams{})
	assert.NoError(t, err)
	assert.Len(t, restoredKeys, 1)

	assert.Equal(t, *restoredKeys[0].StorageID, restoredStorages[0].ID)
}

// TestBackup_RestoreScheduleWithoutTaskParams is a regression test for
// https://github.com/semaphoreui/semaphore/issues/3858 . Backups written by
// older Semaphore versions omit the per-schedule "task_params" object; on
// restore, BackupSchedule.Restore used to dereference the nil pointer and
// crash the HTTP handler with a runtime nil-pointer panic.
func TestBackup_RestoreScheduleWithoutTaskParams(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := sql.InitConfigCreateTestStore()

	// An old-format backup payload: a single template plus a single
	// schedule with no "task_params" object at all. Restore() should
	// succeed and recreate the schedule, not panic.
	payload := `{
  "environments": [],
  "integration_aliases": [],
  "integrations": [],
  "inventories": [],
  "keys": [
    {
      "name": "noop",
      "owner": "",
      "type": "none"
    }
  ],
  "meta": {
    "alert": false,
    "max_parallel_tasks": 0,
    "name": "Restored Project",
    "type": ""
  },
  "repositories": [
    {
      "git_branch": "master",
      "git_url": "git@example.com:test/test.git",
      "name": "Test Repo",
      "ssh_key": "noop"
    }
  ],
  "roles": [],
  "runners": [],
  "schedules": [
    {
      "active": true,
      "cron_format": "0 0 * * *",
      "delete_after_run": false,
      "name": "nightly",
      "template": "Test Template",
      "type": ""
    }
  ],
  "secret_storages": [],
  "templates": [
    {
      "allow_override_args_in_task": false,
      "app": "",
      "autorun": false,
      "name": "Test Template",
      "playbook": "test.yml",
      "repository": "Test Repo",
      "roles": [],
      "suppress_success_alerts": false,
      "suppress_error_alerts": false,
      "type": "",
      "vaults": [],
      "view": null,
      "environments": []
    }
  ],
  "views": []
}`

	restoredBackup := &BackupFormat{}
	err := restoredBackup.Unmarshal(payload)
	assert.NoError(t, err)

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "schedrestore",
			Name:     "Test",
			Email:    "schedrestore@example.com",
			Admin:    true,
		},
	})
	assert.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	assert.NoError(t, err)

	restoredSchedules, err := store.GetSchedules()
	assert.NoError(t, err)
	var found bool
	for _, s := range restoredSchedules {
		if s.ProjectID == restoredProj.ID && s.Name == "nightly" {
			found = true
			break
		}
	}
	assert.True(t, found, "restored schedule should be persisted")
}

// TestBackup_Workflow moved to pro_impl/db/sql/backup_workflow_test.go because
// workflow persistence is a Pro feature requiring the real workflow store
// (the open-source build only has the no-op stub).

func isUnique(items []testItem) bool {
	for i, item := range items {
		for k, other := range items {
			if i == k {
				continue
			}

			if item.Name == other.Name {
				return false
			}
		}
	}

	return true
}

func TestMakeUniqueNames(t *testing.T) {
	items := []testItem{
		{Name: "Project"},
		{Name: "Solution"},
		{Name: "Project"},
		{Name: "Project"},
		{Name: "Project"},
		{Name: "Project"},
	}

	makeUniqueNames(items, func(item *testItem) string {
		return item.Name
	}, func(item *testItem, name string) {
		item.Name = name
	})

	assert.True(t, isUnique(items), "Not unique names")
}

// TestBackupProject_Proxy checks that a proxy and the inventory referencing it
// survive a backup/restore round trip.
func TestBackupProject_Proxy(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "Proxy 123"})
	require.NoError(t, err)

	// An ssh jump host must carry an ssh key: ValidateProxy rejects any other
	// type, and the key picker only offers ssh keys for an ssh proxy.
	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Name:      "bastion key",
		Type:      db.AccessKeySSH,
		SshKey:    db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	port := 2222
	user := "ansible-proxy"
	proxy, err := store.CreateProxy(db.Proxy{
		ProjectID: proj.ID,
		Name:      "bastion-projA",
		Type:      db.ProxySSH,
		Host:      "bastion.example.org",
		Port:      &port,
		User:      &user,
		SSHKeyID:  &key.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		Name:      "managed hosts",
		Type:      db.InventoryStatic,
		ProxyID:   &proxy.ID,
	})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	require.Len(t, backup.Proxies, 1)
	assert.Equal(t, "bastion-projA", backup.Proxies[0].Name)
	require.NotNil(t, backup.Proxies[0].SSHKey)
	assert.Equal(t, "bastion key", *backup.Proxies[0].SSHKey)
	require.Len(t, backup.Inventories, 1)
	require.NotNil(t, backup.Inventories[0].Proxy)
	assert.Equal(t, "bastion-projA", *backup.Inventories[0].Proxy)

	str, err := backup.Marshal()
	require.NoError(t, err)

	restoredBackup := &BackupFormat{}
	require.NoError(t, restoredBackup.Unmarshal(str))
	restoredBackup.Meta.Name = "Proxy 1234"

	user2, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "proxytest",
			Name:     "Proxy Test",
			Email:    "proxy@example.com",
			Admin:    true,
		},
	})
	require.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user2, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	restoredProxies, err := store.GetProxies(restoredProj.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restoredProxies, 1)
	assert.Equal(t, "bastion.example.org", restoredProxies[0].Host)
	require.NotNil(t, restoredProxies[0].Port)
	assert.Equal(t, 2222, *restoredProxies[0].Port)

	restoredInventories, err := store.GetInventories(restoredProj.ID, db.RetrieveQueryParams{}, []db.InventoryType{})
	require.NoError(t, err)
	require.Len(t, restoredInventories, 1)
	require.NotNil(t, restoredInventories[0].ProxyID, "inventory must keep its proxy after restore")
	assert.Equal(t, restoredProxies[0].ID, *restoredInventories[0].ProxyID)
}

// TestRestore_RejectsInvalidProxy proves a hand-edited backup goes through the
// same validation as the API. Without it a restore is a way around every proxy
// rule, including the host allowlist which keeps shell syntax out of the ssh
// ProxyCommand.
func TestRestore_RejectsInvalidProxy(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "restoreproxy", Name: "Restore Proxy",
			Email: "restoreproxy@example.com", Admin: true,
		},
	})
	require.NoError(t, err)

	newBackup := func(name string, proxies []BackupProxy) *BackupFormat {
		return &BackupFormat{
			Meta:    BackupMeta{Project: db.Project{Name: name}},
			Proxies: proxies,
		}
	}

	t.Run("a host carrying shell syntax is rejected", func(t *testing.T) {
		backup := newBackup("bad host", []BackupProxy{{
			Proxy: db.Proxy{Name: "evil", Type: db.ProxySSH, Host: "bastion.example.org;id;"},
		}})

		_, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "proxy host")
	})

	t.Run("a chain on a non-ssh proxy is rejected", func(t *testing.T) {
		inner := "inner"
		backup := newBackup("bad chain", []BackupProxy{
			{Proxy: db.Proxy{Name: "inner", Type: db.ProxySSH, Host: "inner.example.org"}},
			{
				Proxy:         db.Proxy{Name: "socks", Type: db.ProxySOCKS5, Host: "socks.example.org"},
				RequiresProxy: &inner,
			},
		})

		_, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "only an ssh proxy can be chained")
	})

	t.Run("a valid proxy still restores", func(t *testing.T) {
		backup := newBackup("good proxy", []BackupProxy{{
			Proxy: db.Proxy{Name: "bastion", Type: db.ProxySSH, Host: "bastion.example.org"},
		}})

		project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

		require.NoError(t, err)
		proxies, err := store.GetProxies(project.ID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		require.Len(t, proxies, 1)
		assert.Equal(t, "bastion.example.org", proxies[0].Host)
	})
}
