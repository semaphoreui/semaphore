package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
  "version": 4,
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
			Version: BackupFormatVersion,
			Meta:    BackupMeta{Project: db.Project{Name: name}},
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
  "version": 4,
  "keys": [{"name": "noop", "owner": "", "type": "none"}],
  "meta": {"name": "Restored Roles", "type": ""},
  "repositories": [{"git_branch": "master", "git_url": "git@example.com:test/test.git", "name": "Repo", "ssh_key": "noop"}],
  "roles": [{"name": "ops", "slug": "ops", "permissions": 1048577}],
  "templates": [{
    "app": "", "name": "Tpl", "playbook": "test.yml", "repository": "Repo", "type": "",
    "roles": [{"slug": "ops", "permissions": 1048580}],
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

// A template integration round-trips with its target by name; an
// integration whose workflow the store cannot return (the open-source
// workflow store persists nothing) is left out of the backup instead of being
// written as an empty entry.
func TestBackup_IntegrationTargets(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "integrations"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID, SSHKeyID: key.ID, Name: "Repo", GitURL: "git@example.com:test/test", GitBranch: "master",
	})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{Name: "Deploy", Playbook: "deploy.yml", ProjectID: proj.ID, RepositoryID: repo.ID})
	require.NoError(t, err)

	_, err = store.CreateIntegration(db.Integration{Name: "template hook", ProjectID: proj.ID, TemplateID: &tpl.ID})
	require.NoError(t, err)

	_, err = store.Sql().Exec(store.PrepareQuery(
		"insert into project__workflow_template (project_id, name) values (?, ?)"), proj.ID, "WF")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt(store.PrepareQuery(
		"select id from project__workflow_template where project_id = ? and name = ?"), proj.ID, "WF")
	require.NoError(t, err)
	wfID := int(workflowID)
	_, err = store.CreateIntegration(db.Integration{Name: "workflow hook", ProjectID: proj.ID, WorkflowID: &wfID})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	// With the open-source workflow store the backup carries no workflows and
	// the workflow hook is left out; with the Pro store (workspace build) it is
	// exported by the workflow's name. Either way it is never an empty entry.
	if len(backup.Workflows) == 0 {
		require.Len(t, backup.Integration, 1, "an integration without its workflow in the backup is skipped")
	} else {
		require.Len(t, backup.Integration, 2)
		require.NotNil(t, backup.Integration[1].Workflow)
		assert.Equal(t, "WF", *backup.Integration[1].Workflow)
		assert.Nil(t, backup.Integration[1].Template)
	}
	require.NotNil(t, backup.Integration[0].Template)
	assert.Equal(t, "Deploy", *backup.Integration[0].Template)
	assert.Nil(t, backup.Integration[0].Workflow)

	str, err := backup.Marshal()
	require.NoError(t, err)
	assert.Contains(t, string(str), `"template": "Deploy"`)

	restoredBackup := &BackupFormat{}
	require.NoError(t, restoredBackup.Unmarshal(str))
	restoredBackup.Meta.Name = "integrations restored"
	// The hand-inserted workflow has no graph and cannot be restored by the
	// Pro store; the workflow round-trip lives in pro_impl's backup test.
	restoredBackup.Workflows = nil
	restoredBackup.Integration = restoredBackup.Integration[:1]

	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "u", Name: "U", Email: "u@example.com", Admin: true}})
	require.NoError(t, err)
	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	restoredTemplates, err := store.GetTemplates(restoredProj.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restoredTemplates, 1)
	restored, err := store.GetIntegrations(restoredProj.ID, db.RetrieveQueryParams{}, false)
	require.NoError(t, err)
	require.Len(t, restored, 1)
	assert.Equal(t, "template hook", restored[0].Name)
	require.NotNil(t, restored[0].TemplateID)
	assert.Equal(t, restoredTemplates[0].ID, *restored[0].TemplateID)
	assert.Nil(t, restored[0].WorkflowID)
}

// Verify refuses an integration whose target is missing, doubled or unknown
// before anything is written.
func TestRestore_IntegrationVerify(t *testing.T) {
	templateName := "Deploy"
	workflowName := "WF"
	base := func() *BackupFormat {
		return &BackupFormat{
			Templates: []BackupTemplate{{Template: db.Template{Name: templateName}}},
			Workflows: []BackupWorkflow{{WorkflowTemplate: db.WorkflowTemplate{Name: workflowName}}},
		}
	}

	tests := []struct {
		name        string
		integration BackupIntegration
		wantErr     string
	}{
		{"template target", BackupIntegration{Template: &templateName}, ""},
		{"workflow target", BackupIntegration{Workflow: &workflowName}, ""},
		{"both", BackupIntegration{Template: &templateName, Workflow: &workflowName}, "either a template or a workflow"},
		{"none", BackupIntegration{}, "neither a template nor a workflow"},
		{"unknown template", BackupIntegration{Template: new("missing")}, "template does not exist"},
		{"unknown workflow", BackupIntegration{Workflow: new("missing")}, "workflow does not exist"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.integration.Verify(base())
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
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
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: p.ID, TemplateID: tpl.ID, RoleID: runner.ID, Permissions: db.CanManageProjectResources})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: p.ID, TemplateID: tpl.ID, RoleID: global.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)

	backup, err := GetBackup(p.ID, store, workflows)
	require.NoError(t, err)
	require.Len(t, backup.Templates, 1)
	assert.ElementsMatch(t, []BackupTemplateRole{
		{Slug: "task_runner", Permissions: db.CanManageProjectResources},
		{Slug: "deployer", Permissions: db.CanRunProjectTasks},
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
	require.Len(t, grants, 1)
	assert.Equal(t, "task_runner", grants[0].RoleSlug)
	assert.Equal(t, []UnassignedRole{{Slug: "deployer", Reason: "not_in_file", Templates: []UnassignedTemplate{
		{ID: templates[0].ID, Name: templates[0].Name, Permissions: db.CanRunProjectTasks},
	}, Workflows: []UnassignedWorkflow{}}}, restored.UnassignedRoles)
}

func TestRestore_ReportsRolesOutsideTheFile(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	p, _ := createBackupFixture(t, store)
	backup, err := GetBackup(p.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	backup.Meta.Name = "Missing role"
	backup.Templates[0].Roles = []BackupTemplateRole{{Slug: "nobody", Permissions: db.CanRunProjectTasks}}
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "m", Name: "m", Email: "m@example.com", Admin: true}})
	require.NoError(t, err)

	restored, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

	require.NoError(t, err)
	templates, err := store.GetTemplates(restored.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	grants, err := store.GetTemplateRoles(restored.ID, templates[0].ID)
	require.NoError(t, err)
	assert.Empty(t, grants)
	assert.Equal(t, []UnassignedRole{{Slug: "nobody", Reason: "not_in_file", Templates: []UnassignedTemplate{
		{ID: templates[0].ID, Name: templates[0].Name, Permissions: db.CanRunProjectTasks},
	}, Workflows: []UnassignedWorkflow{}}}, restored.UnassignedRoles)
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
	for _, key := range []string{"id", "scope", "builtin", "project_id"} {
		assert.NotContains(t, doc.Roles[0], key)
	}
	assert.Equal(t, "local", doc.Roles[0]["slug"])
}

func TestBackupProject_Roles(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	source := sql.InitConfigCreateTestStore()
	project, err := source.CreateProject(db.Project{Name: "Payments"})
	require.NoError(t, err)
	key, err := source.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	repo, err := source.CreateRepository(db.Repository{Name: "r", ProjectID: project.ID, GitURL: "git@example.com:r.git", GitBranch: "main", SSHKeyID: key.ID})
	require.NoError(t, err)
	template, err := source.CreateTemplate(db.Template{Name: "deploy", ProjectID: project.ID, RepositoryID: repo.ID, Playbook: "site.yml"})
	require.NoError(t, err)
	deployA, err := source.CreateRole(db.Role{Slug: "deploy-a", Name: "Deployer", Permissions: db.CanRunProjectTasks, ProjectID: &project.ID})
	require.NoError(t, err)
	_, err = source.CreateRole(db.Role{Slug: "deploy-b", Name: "Deployer", Permissions: db.CanRunProjectTasks, ProjectID: &project.ID})
	require.NoError(t, err)
	taskRunner, err := source.GetSharedRoleBySlug(string(db.ProjectTaskRunner))
	require.NoError(t, err)
	_, err = source.CreateTemplateRole(db.TemplateRolePerm{TemplateID: template.ID, ProjectID: project.ID, RoleID: deployA.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)
	_, err = source.CreateTemplateRole(db.TemplateRolePerm{TemplateID: template.ID, ProjectID: project.ID, RoleID: taskRunner.ID, Permissions: db.CanManageProjectResources})
	require.NoError(t, err)
	target := sql.InitConfigCreateTestStore()
	user, err := target.CreateUser(db.UserWithPwd{
		Pwd:  "3412341234123",
		User: db.User{Username: "restorer", Name: "Restorer", Email: "restorer@example.com", Admin: true},
	})
	require.NoError(t, err)

	backup, err := GetBackup(project.ID, source, proFactory.NewWorkflowStore(source))
	require.NoError(t, err)
	str, err := backup.Marshal()
	require.NoError(t, err)
	var file BackupFormat
	require.NoError(t, file.Unmarshal(str))
	require.NoError(t, file.Verify())
	copied, err := file.Restore(user, target, proFactory.NewWorkflowStore(target))
	require.NoError(t, err)

	roles, err := target.GetProjectRoles(copied.ID)
	require.NoError(t, err)
	named := map[string]string{}
	for _, role := range roles {
		named[role.Slug] = role.Name
	}
	assert.Equal(t, map[string]string{"deploy-a": "Deployer", "deploy-b": "Deployer"}, named)
	templates, err := target.GetTemplates(copied.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	grants, err := target.GetTemplateRoles(copied.ID, templates[0].ID)
	require.NoError(t, err)
	granted := map[string]db.ProjectUserPermission{}
	for _, grant := range grants {
		granted[grant.RoleSlug] = grant.Permissions
	}
	assert.Equal(t, map[string]db.ProjectUserPermission{"deploy-a": db.CanRunProjectTasks, "task_runner": db.CanManageProjectResources}, granted)
}

func TestBackupProject_KeepsAnOldSlug(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	source := sql.InitConfigCreateTestStore()
	project, err := source.CreateProject(db.Project{Name: "Payments"})
	require.NoError(t, err)
	_, err = source.CreateRole(db.Role{Slug: "my role", Name: "Old", Permissions: db.CanRunProjectTasks, ProjectID: &project.ID})
	require.NoError(t, err)
	target := sql.InitConfigCreateTestStore()
	user, err := target.CreateUser(db.UserWithPwd{
		Pwd:  "3412341234123",
		User: db.User{Username: "restorer", Name: "Restorer", Email: "restorer@example.com", Admin: true},
	})
	require.NoError(t, err)

	backup, err := GetBackup(project.ID, source, proFactory.NewWorkflowStore(source))
	require.NoError(t, err)
	str, err := backup.Marshal()
	require.NoError(t, err)
	var file BackupFormat
	require.NoError(t, file.Unmarshal(str))
	require.NoError(t, file.Verify())
	copied, err := file.Restore(user, target, proFactory.NewWorkflowStore(target))
	require.NoError(t, err)

	roles, err := target.GetProjectRoles(copied.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, "my role", roles[0].Slug)
}

func TestVerify_RefusesDuplicateRoleSlugs(t *testing.T) {
	backup := &BackupFormat{
		Version: BackupFormatVersion,
		Meta:    BackupMeta{Project: db.Project{Name: "p"}},
		Roles: []BackupRole{
			{Role: db.Role{Slug: "ops", Name: "Ops"}},
			{Role: db.Role{Slug: "OPS", Name: "Ops again"}},
		},
	}

	err := backup.Verify()

	assert.ErrorContains(t, err, "error at roles[0]: Role slug is duplicate: ops")
}

func TestVerify_RoleSlugIsAddressable(t *testing.T) {
	tests := []struct {
		name string
		slug string
	}{
		{"no slug", ""},
		{"reserved slug", "all"},
		{"slug longer than the column", strings.Repeat("x", 101)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backup := &BackupFormat{
				Version: BackupFormatVersion,
				Meta:    BackupMeta{Project: db.Project{Name: "p"}},
				Roles:   []BackupRole{{Role: db.Role{Slug: tt.slug, Name: "Deployer"}}},
			}

			err := backup.Verify()

			assert.ErrorContains(t, err, "error at roles[0]: Role slug")
		})
	}
}

func TestRestore_ReportsATakenSlug(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	p, tpl := createBackupFixture(t, store)
	own, err := store.CreateRole(db.Role{Slug: "ops", Name: "Deployer", Permissions: db.CanRunProjectTasks, ProjectID: &p.ID})
	require.NoError(t, err)
	_, err = store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: p.ID, TemplateID: tpl.ID, RoleID: own.ID, Permissions: db.CanManageProjectResources})
	require.NoError(t, err)
	backup, err := GetBackup(p.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)
	backup.Meta.Name = "Taken slug"
	// A shared role cannot take a slug a project role holds, so the source project goes first.
	require.NoError(t, store.DeleteProject(p.ID))
	_, err = store.CreateRole(db.Role{Slug: "ops", Name: "Ops"})
	require.NoError(t, err)
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "m", Name: "m", Email: "m@example.com", Admin: true}})
	require.NoError(t, err)

	restored, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

	require.NoError(t, err)
	roles, err := store.GetProjectRoles(restored.ID)
	require.NoError(t, err)
	assert.Empty(t, roles)
	templates, err := store.GetTemplates(restored.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	assert.Equal(t, []UnassignedRole{{Slug: "ops", Reason: "slug_taken", Name: "Deployer", Permissions: db.CanRunProjectTasks, Templates: []UnassignedTemplate{
		{ID: templates[0].ID, Name: templates[0].Name, Permissions: db.CanManageProjectResources},
	}, Workflows: []UnassignedWorkflow{}}}, restored.UnassignedRoles)
}

func TestRestore_UnassignedRolesIsEmptyList(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "m", Name: "m", Email: "m@example.com", Admin: true}})
	require.NoError(t, err)
	backup := &BackupFormat{Version: BackupFormatVersion, Meta: BackupMeta{Project: db.Project{Name: "plain"}}}

	restored, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))

	require.NoError(t, err)
	assert.NotNil(t, restored.UnassignedRoles)
	assert.Empty(t, restored.UnassignedRoles)
}

func TestVerify_NamesTheScheduleEntry(t *testing.T) {
	backup := &BackupFormat{
		Version:   BackupFormatVersion,
		Meta:      BackupMeta{Project: db.Project{Name: "p"}},
		Schedules: []BackupSchedule{{Schedule: db.Schedule{Name: "nightly"}}, {Schedule: db.Schedule{Name: "nightly"}}, {Schedule: db.Schedule{Name: "nightly"}}},
	}

	err := backup.Verify()

	assert.ErrorContains(t, err, "error at schedules[0]: nightly is duplicate")
}

type failingDelete struct{ db.Store }

func (failingDelete) DeleteProject(int) error { return errors.New("locked") }

func TestRestore_ReportsAFailedCleanup(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "m", Name: "m", Email: "m@example.com", Admin: true}})
	require.NoError(t, err)
	backup := &BackupFormat{
		Version:   BackupFormatVersion,
		Meta:      BackupMeta{Project: db.Project{Name: "stuck"}},
		Templates: []BackupTemplate{{Template: db.Template{Name: "deploy"}, Repository: "gone"}},
	}

	_, err = backup.Restore(user, failingDelete{store}, proFactory.NewWorkflowStore(store))

	require.ErrorContains(t, err, "error at templates[0]")
	assert.ErrorContains(t, err, "could not be removed: locked")
}

func TestVerify_RefusesOtherFormatVersions(t *testing.T) {
	grant := []BackupTemplateRole{{Slug: "ops", Permissions: db.CanRunProjectTasks}}
	tests := []struct {
		name    string
		backup  BackupFormat
		wantErr string
	}{
		{"no version with roles", BackupFormat{Roles: []BackupRole{{Role: db.Role{Slug: "ops", Name: "Ops"}}}}, "older version of Semaphore"},
		{"no version with a grant", BackupFormat{Templates: []BackupTemplate{{Template: db.Template{Name: "t"}, Roles: grant}}}, "older version of Semaphore"},
		{"newer", BackupFormat{Version: 5}, "newer version of Semaphore"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backup := tt.backup
			backup.Meta = BackupMeta{Project: db.Project{Name: "p"}}

			err := backup.Verify()

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestVerify_ReadsAnOldFileWithoutRoles(t *testing.T) {
	backup := &BackupFormat{Meta: BackupMeta{Project: db.Project{Name: "p"}}}

	err := backup.Verify()

	assert.NoError(t, err)
}

// A string key travels through a backup like every other key: by name and
// type only. Its value stays behind, and the restored key is empty until the
// owner fills it in again.
func TestBackupProject_StringKey(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "String key"})
	require.NoError(t, err)

	ciphertext := "ciphertext-of-s3cr3t"
	_, err = store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Name:      "api-token",
		Type:      db.AccessKeyString,
		String:    "s3cr3t",
		Secret:    &ciphertext,
	})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	str, err := backup.Marshal()
	require.NoError(t, err)
	assert.NotContains(t, str, "s3cr3t", "a backup must not carry the value")
	assert.NotContains(t, str, ciphertext, "a backup must not carry the ciphertext either")

	var doc struct {
		Keys []map[string]any `json:"keys"`
	}
	require.NoError(t, json.Unmarshal([]byte(str), &doc))
	require.Len(t, doc.Keys, 1)
	assert.Equal(t, "api-token", doc.Keys[0]["name"])
	assert.Equal(t, "string", doc.Keys[0]["type"])
	assert.NotContains(t, doc.Keys[0], "string", "the value field must not even be present")
	assert.NotContains(t, doc.Keys[0], "secret")

	restoredBackup := &BackupFormat{}
	require.NoError(t, restoredBackup.Unmarshal(str))
	restoredBackup.Meta.Name = "String key restored"
	require.NoError(t, restoredBackup.Verify())

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "stringkeytest", Name: "String Key Test",
			Email: "stringkey@example.com", Admin: true,
		},
	})
	require.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	restoredKeys, err := store.GetAccessKeys(restoredProj.ID, db.GetAccessKeyOptions{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restoredKeys, 1)
	restored := restoredKeys[0]
	assert.Equal(t, "api-token", restored.Name)
	assert.Equal(t, db.AccessKeyString, restored.Type)
	assert.Equal(t, db.AccessKeyShared, restored.Owner)
	assert.Nil(t, restored.Secret)
	assert.Empty(t, restored.String)
	assert.True(t, restored.Empty, "the restored key waits for a value")
}

// A hand-edited backup may point a credential mapping at a string key; the
// preflight applies the same rule as the API and rejects it before anything
// is written.
func TestRestore_RejectsStringKeyInHostConfig(t *testing.T) {
	keyName := "api-token"
	tests := []struct {
		name    string
		mapping db.HostConfig
		wantErr string
	}{
		{
			name:    "host mapping",
			mapping: db.HostConfig{Type: db.HostConfigHost, Name: "github.com"},
			wantErr: "a host mapping needs an SSH key",
		},
		{
			name:    "url mapping",
			mapping: db.HostConfig{Type: db.HostConfigURL, Name: "https://github.com/acme/"},
			wantErr: "a URL mapping needs an SSH key or a login/password credential",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backup := &BackupFormat{
				Meta: BackupMeta{Project: db.Project{Name: "string key mapping"}},
				Keys: []BackupAccessKey{{
					AccessKey: db.AccessKey{Name: keyName, Type: db.AccessKeyString},
				}},
				HostConfigs: []BackupHostConfig{{HostConfig: tt.mapping, SSHKey: &keyName}},
			}

			err := backup.Verify()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// A file written before the workflow verbs carries roles and grants without them; restoring it gives them the
// same bits the upgrade gives stored roles.
func TestRestore_LiftsMasksOfAVersion2File(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	store := sql.InitConfigCreateTestStore()

	payload := `{
  "version": 2,
  "keys": [{"name": "noop", "owner": "", "type": "none"}],
  "meta": {"name": "Old Roles", "type": ""},
  "repositories": [{"git_branch": "master", "git_url": "git@example.com:test/test.git", "name": "Repo", "ssh_key": "noop"}],
  "roles": [{"name": "ops", "slug": "ops", "permissions": 5}],
  "templates": [{
    "app": "", "name": "Tpl", "playbook": "test.yml", "repository": "Repo", "type": "",
    "roles": [{"slug": "ops", "permissions": 1}],
    "vaults": [], "view": null, "environments": []
  }]
}`
	backup := &BackupFormat{}
	require.NoError(t, backup.Unmarshal(payload))
	require.NoError(t, backup.Verify())

	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "old", Name: "old", Email: "old@example.com", Admin: true}})
	require.NoError(t, err)
	project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	roles, err := store.GetProjectRoles(project.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, db.ProjectUserPermission(245), roles[0].Permissions, "5 lifted: run and approve workflows, edit workflows, read")

	templates, err := store.GetTemplates(project.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	perms, err := store.GetTemplateRoles(project.ID, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, db.CanRunProjectTasks|db.CanReadProject, perms[0].Permissions, "a template grant gains only read")
}

// Version 3 files carry the workflow verbs but not read, version 4 files carry everything as written.
func TestRestore_ReadBitByFileVersion(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}
	tests := []struct {
		version   int
		roleMask  db.ProjectUserPermission
		wantRole  db.ProjectUserPermission
		wantGrant db.ProjectUserPermission
	}{
		{3, 0, 128, 129},
		{3, 1, 129, 129},
		{4, 0, 0, 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("version %d mask %d", tt.version, tt.roleMask), func(t *testing.T) {
			store := sql.InitConfigCreateTestStore()
			payload := fmt.Sprintf(`{
  "version": %d,
  "keys": [{"name": "noop", "owner": "", "type": "none"}],
  "meta": {"name": "Versioned", "type": ""},
  "repositories": [{"git_branch": "master", "git_url": "git@example.com:test/test.git", "name": "Repo", "ssh_key": "noop"}],
  "roles": [{"name": "ops", "slug": "ops", "permissions": %d}],
  "templates": [{
    "app": "", "name": "Tpl", "playbook": "test.yml", "repository": "Repo", "type": "",
    "roles": [{"slug": "ops", "permissions": 1}],
    "vaults": [], "view": null, "environments": []
  }]
}`, tt.version, tt.roleMask)
			backup := &BackupFormat{}
			require.NoError(t, backup.Unmarshal(payload))
			require.NoError(t, backup.Verify())
			user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "v", Name: "v", Email: "v@example.com", Admin: true}})
			require.NoError(t, err)

			project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))
			require.NoError(t, err)

			roles, err := store.GetProjectRoles(project.ID)
			require.NoError(t, err)
			require.Len(t, roles, 1)
			assert.Equal(t, tt.wantRole, roles[0].Permissions)
			templates, err := store.GetTemplates(project.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
			require.NoError(t, err)
			perms, err := store.GetTemplateRoles(project.ID, templates[0].ID)
			require.NoError(t, err)
			require.Len(t, perms, 1)
			assert.Equal(t, tt.wantGrant, perms[0].Permissions)
		})
	}
}
