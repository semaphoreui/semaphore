package project

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db/sql"

	"github.com/semaphoreui/semaphore/db"
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

	store := sql.CreateTestStore()

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
		Name:           "Test",
		Playbook:       "test.yml",
		ProjectID:      proj.ID,
		RepositoryID:   repo.ID,
		InventoryID:    &inv.ID,
		EnvironmentIDs: []int{env.ID},
	})
	assert.NoError(t, err)

	backup, err := GetBackup(proj.ID, store)
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

	restoredProj, err := restoredBackup.Restore(user, store)
	assert.NoError(t, err)
	assert.Equal(t, restoredProj.Name, "Test 1234")

	restoredTemplates, err := store.GetTemplates(restoredProj.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	assert.NoError(t, err)
	assert.Len(t, restoredTemplates, 1)
	assert.Len(t, restoredTemplates[0].EnvironmentIDs, 1)

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

	store := sql.CreateTestStore()

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

	backup, err := GetBackup(proj.ID, store)
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

	restoredProj, err := restoredBackup.Restore(user, store)
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

	store := sql.CreateTestStore()

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

	restoredProj, err := restoredBackup.Restore(user, store)
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

func TestBackup_Workflow(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := sql.CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "Test WF",
	})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		Name:      "Inv",
	})
	require.NoError(t, err)

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "Env",
		JSON:      `{}`,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Tpl",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: proj.ID,
		Name:      "WF",
		Nodes: []db.WorkflowNode{
			{
				ID:            1,
				Kind:          db.WorkflowNodeTaskKind,
				TemplateID:    tpl.ID,
				InventoryID:   &inv.ID,
				EnvironmentID: &env.ID,
			},
			{
				ID:         2,
				Kind:       db.WorkflowNodeTaskKind,
				TemplateID: tpl.ID,
			},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store)
	require.NoError(t, err)
	require.Len(t, backup.Workflows, 1)
	assert.Equal(t, "WF", backup.Workflows[0].Name)
	require.Len(t, backup.Workflows[0].Nodes, 2)
	// Node references must be exported by name, not ID.
	require.NotNil(t, backup.Workflows[0].Nodes[0].Template)
	assert.Equal(t, "Tpl", *backup.Workflows[0].Nodes[0].Template)
	require.NotNil(t, backup.Workflows[0].Nodes[0].Inventory)
	assert.Equal(t, "Inv", *backup.Workflows[0].Nodes[0].Inventory)
	require.NotNil(t, backup.Workflows[0].Nodes[0].Environment)
	assert.Equal(t, "Env", *backup.Workflows[0].Nodes[0].Environment)

	str, err := backup.Marshal()
	require.NoError(t, err)

	restoredBackup := &BackupFormat{}
	require.NoError(t, restoredBackup.Unmarshal(str))

	restoredBackup.Meta.Name = "Test WF Restored"

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "wf",
			Name:     "WF",
			Email:    "wf@example.com",
			Admin:    true,
		},
	})
	require.NoError(t, err)

	restoredProj, err := restoredBackup.Restore(user, store)
	require.NoError(t, err)

	restoredWorkflows, err := store.GetWorkflowTemplates(restoredProj.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restoredWorkflows, 1)
	assert.Equal(t, "WF", restoredWorkflows[0].Name)
	require.Len(t, restoredWorkflows[0].Nodes, 2)
	require.Len(t, restoredWorkflows[0].Edges, 1)

	// The restored task node must point at the restored template/inventory/env,
	// whose IDs differ from the originals.
	restoredTemplates, err := store.GetTemplates(restoredProj.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, restoredTemplates, 1)
	assert.Equal(t, restoredTemplates[0].ID, restoredWorkflows[0].Nodes[0].TemplateID)
	require.NotNil(t, restoredWorkflows[0].Nodes[0].InventoryID)
	require.NotNil(t, restoredWorkflows[0].Nodes[0].EnvironmentID)

	// The edge must still connect the two restored nodes (IDs were remapped).
	edge := restoredWorkflows[0].Edges[0]
	assert.Equal(t, restoredWorkflows[0].Nodes[0].ID, edge.SourceNodeID)
	assert.Equal(t, restoredWorkflows[0].Nodes[1].ID, edge.DestinationNodeID)
}

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
