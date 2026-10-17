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
  "host_configs": [],
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

// TestBackupProject_HostConfig covers the credential mappings through a full
// backup/restore cycle. Without it a restored project silently loses them and
// its tasks lose the credentials they reach other hosts with.
func TestBackupProject_HostConfig(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "Host config 123"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Name:      "GitHub deploy key",
		Type:      db.AccessKeySSH,
		SshKey:    db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	_, err = store.CreateHostConfig(db.HostConfig{
		ProjectID: proj.ID, Type: db.HostConfigHost,
		Name: "github.com", SSHKeyID: key.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateHostConfig(db.HostConfig{
		ProjectID: proj.ID, Type: db.HostConfigURL,
		Name: "https://github.com/acme/private/", SSHKeyID: key.ID,
	})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	require.Len(t, backup.HostConfigs, 2)
	// The credential travels by name, never by id, and never as key material.
	require.NotNil(t, backup.HostConfigs[0].SSHKey)
	assert.Equal(t, "GitHub deploy key", *backup.HostConfigs[0].SSHKey)

	str, err := backup.Marshal()
	require.NoError(t, err)
	assert.NotContains(t, str, "PrivateKey", "a backup must not carry key material")

	restoredBackup := &BackupFormat{}
	require.NoError(t, restoredBackup.Unmarshal(str))
	restoredBackup.Meta.Name = "Host config 1234"

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "hostconfigtest", Name: "Host Config Test",
			Email: "hostconfig@example.com", Admin: true,
		},
	})
	require.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	restored, err := store.GetHostConfigs(restoredProj.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restored, 2)

	byName := map[string]db.HostConfig{}
	for _, hc := range restored {
		byName[hc.Name] = hc
	}

	assert.Equal(t, db.HostConfigHost, byName["github.com"].Type)
	assert.Equal(t, db.HostConfigURL, byName["https://github.com/acme/private/"].Type)

	// The mapping must point at the restored key of the restored project, not at
	// the id it had in the original one.
	restoredKeys, err := store.GetAccessKeys(restoredProj.ID, db.GetAccessKeyOptions{}, db.RetrieveQueryParams{})
	require.NoError(t, err)

	var restoredKeyID int
	for _, k := range restoredKeys {
		if k.Name == "GitHub deploy key" {
			restoredKeyID = k.ID
		}
	}
	require.NotZero(t, restoredKeyID)
	assert.Equal(t, restoredKeyID, byName["github.com"].SSHKeyID)
}

// A hand-edited backup must not get a mapping past the checks the API applies.
func TestRestore_RejectsInvalidHostConfig(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "badhostconfig", Name: "Bad Host Config",
			Email: "badhostconfig@example.com", Admin: true,
		},
	})
	require.NoError(t, err)

	keyName := "k"
	newBackup := func(name string, hostConfigs []BackupHostConfig) *BackupFormat {
		return &BackupFormat{
			Meta: BackupMeta{Project: db.Project{Name: name}},
			Keys: []BackupAccessKey{{
				AccessKey: db.AccessKey{Name: keyName, Type: db.AccessKeySSH},
			}},
			HostConfigs: hostConfigs,
		}
	}

	// Preflight must reject it: Restore only runs after the project and its keys
	// exist, so a failure there leaves a half-restored project behind.
	t.Run("a malformed host is rejected by preflight", func(t *testing.T) {
		backup := newBackup("bad host preflight", []BackupHostConfig{{
			HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com\n  IdentityFile /etc/shadow"},
			SSHKey:     &keyName,
		}})

		err := backup.Verify()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "host")
	})

	// A valid backup must survive preflight even though the credential is still
	// carried by name and its id is not resolved yet.
	t.Run("preflight accepts a mapping whose key is not resolved yet", func(t *testing.T) {
		backup := newBackup("unresolved key", []BackupHostConfig{{
			HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"},
			SSHKey:     &keyName,
		}})

		require.Zero(t, backup.HostConfigs[0].SSHKeyID, "the key id is only known during the restore")
		assert.NoError(t, backup.Verify())
	})

	t.Run("a host carrying config syntax is rejected", func(t *testing.T) {
		backup := newBackup("bad host", []BackupHostConfig{{
			HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com\n  IdentityFile /etc/shadow"},
			SSHKey:     &keyName,
		}})

		_, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "host_configs")
	})

	// The API and the CLI both run Verify before Restore, so a duplicate must be
	// rejected there: by the time Restore runs, the project already exists.
	t.Run("a duplicate mapping is rejected by preflight", func(t *testing.T) {
		backup := newBackup("dup host", []BackupHostConfig{
			{HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"}, SSHKey: &keyName},
			{HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"}, SSHKey: &keyName},
		})

		err := backup.Verify()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	// The stored name is trimmed, so these are one mapping and the second would
	// hit the unique index after the project already exists.
	t.Run("a duplicate differing only in whitespace is rejected by preflight", func(t *testing.T) {
		backup := newBackup("dup whitespace", []BackupHostConfig{
			{HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"}, SSHKey: &keyName},
			{HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "  github.com  "}, SSHKey: &keyName},
		})

		err := backup.Verify()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	t.Run("the same name with a different type passes preflight and restores", func(t *testing.T) {
		backup := newBackup("same name two types", []BackupHostConfig{
			{HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"}, SSHKey: &keyName},
			{HostConfig: db.HostConfig{Type: db.HostConfigURL, Name: "https://github.com/acme/"}, SSHKey: &keyName},
		})

		require.NoError(t, backup.Verify())

		project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))
		require.NoError(t, err)

		hostConfigs, err := store.GetHostConfigs(project.ID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		assert.Len(t, hostConfigs, 2)
	})

	t.Run("a valid mapping still restores", func(t *testing.T) {
		backup := newBackup("good host", []BackupHostConfig{{
			HostConfig: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"},
			SSHKey:     &keyName,
		}})

		project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

		require.NoError(t, err)
		hostConfigs, err := store.GetHostConfigs(project.ID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		require.Len(t, hostConfigs, 1)
		assert.Equal(t, "github.com", hostConfigs[0].Name)
	})
}

// Old backups may carry bits the application never checks: restore keeps the roles and drops the bits.
func TestRestore_DropsUnknownPermissionBits(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	payload := `{
  "keys": [{"name": "noop", "owner": "", "type": "none"}],
  "meta": {"name": "Restored Roles", "type": ""},
  "repositories": [{"git_branch": "master", "git_url": "git@example.com:test/test.git", "name": "Repo", "ssh_key": "noop"}],
  "roles": [{"name": "ops", "permissions": 1048577}],
  "templates": [{
    "app": "", "name": "Tpl", "playbook": "test.yml", "repository": "Repo", "type": "",
    "roles": [{"role": "ops", "is_global": false, "permissions": 1048580}],
    "vaults": [], "view": null, "environments": []
  }]
}`
	backup := &BackupFormat{}
	require.NoError(t, backup.Unmarshal(payload))
	require.NoError(t, backup.Verify())

	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "roles", Name: "roles", Email: "roles@example.com", Admin: true}})
	require.NoError(t, err)
	project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	roles, err := store.GetProjectRoles(project.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, db.CanRunProjectTasks, roles[0].Permissions)

	templates, err := store.GetTemplates(project.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	perms, err := store.GetTemplateRoles(project.ID, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, db.CanManageProjectResources, perms[0].Permissions)
}

func createBackupFixture(t *testing.T, store db.Store) (db.Project, db.Template) {
	t.Helper()
	proj, err := store.CreateProject(db.Project{Name: "Grants"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{ProjectID: proj.ID, SSHKeyID: key.ID, Name: "Repo", GitURL: "git@example.com:test/test", GitBranch: "master"})
	require.NoError(t, err)
	inv, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{Name: "Tpl", Playbook: "test.yml", ProjectID: proj.ID, RepositoryID: repo.ID, InventoryID: &inv.ID})
	require.NoError(t, err)
	return proj, tpl
}

func TestBackupKeepsBuiltinAndGlobalGrants(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	workflows := proFactory.NewWorkflowStore(store)
	p, tpl := createBackupFixture(t, store)
	runner, err := store.GetSharedRoleBySlug("task_runner")
	require.NoError(t, err)
	global, err := store.CreateRole(db.Role{Slug: "deployer", Name: "Deployer Name"})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: p.ID, TemplateID: tpl.ID, RoleID: runner.ID, Permissions: db.CanUpdateProject})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: p.ID, TemplateID: tpl.ID, RoleID: global.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)

	backup, err := GetBackup(p.ID, store, workflows)
	require.NoError(t, err)
	require.Len(t, backup.Templates, 1)
	assert.ElementsMatch(t, []BackupTemplateRole{
		{Role: "task_runner", IsGlobal: true, Permissions: db.CanUpdateProject},
		{Role: "deployer", IsGlobal: true, Permissions: db.CanRunProjectTasks},
	}, backup.Templates[0].Roles)

	backup.Meta.Name = "Grants restored"
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "g", Name: "g", Email: "g@example.com", Admin: true}})
	require.NoError(t, err)
	restored, err := backup.Restore(user, store, workflows)
	require.NoError(t, err)
	templates, err := store.GetTemplates(restored.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	grants, err := store.GetTemplateRoles(restored.ID, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, grants, 2)
	assert.ElementsMatch(t, []string{"task_runner", "deployer"}, []string{grants[0].RoleSlug, grants[1].RoleSlug})
}

func TestRestore_FailsOnUnknownGlobalGrant(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	p, _ := createBackupFixture(t, store)
	backup, err := GetBackup(p.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	backup.Meta.Name = "Missing global"
	backup.Templates[0].Roles = []BackupTemplateRole{{Role: "nope", IsGlobal: true}}
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "m", Name: "m", Email: "m@example.com", Admin: true}})
	require.NoError(t, err)
	_, err = backup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.Error(t, err)
}

func TestBackupJSONHasNoRoleInternals(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	p, _ := createBackupFixture(t, store)
	_, err := store.CreateRole(db.Role{Slug: "local", Name: "Local", ProjectID: &p.ID})
	require.NoError(t, err)
	backup, err := GetBackup(p.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	raw, err := backup.Marshal()
	require.NoError(t, err)

	var doc struct {
		Roles []map[string]any `json:"roles"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &doc))
	require.Len(t, doc.Roles, 1)
	for _, key := range []string{"id", "scope", "builtin", "slug"} {
		assert.NotContains(t, doc.Roles[0], key)
	}
}
