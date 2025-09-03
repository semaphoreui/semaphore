package project

import (
	"encoding/json"
	"github.com/semaphoreui/semaphore/db/sql"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
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
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   &inv.ID,
		EnvironmentID: &env.ID,
	})
	assert.NoError(t, err)

	backup, err := GetBackup(proj.ID, store)
	assert.NoError(t, err)
	assert.Equal(t, proj.ID, backup.Meta.ID)

	str, err := backup.Marshal()
	assert.NoError(t, err)
	//assert.Equal(t, "{\"environments\":[{\"json\":\"{\\\"author\\\": \\\"Denis\\\", \\\"comment\\\": \\\"Hello, World!\\\"}\",\"name\":\"test\"}],\"integration_aliases\":[],\"integrations\":[],\"inventories\":[{\"inventory\":\"\",\"name\":\"\",\"type\":\"\"}],\"keys\":[{\"name\":\"\",\"type\":\"none\"}],\"meta\":{\"alert\":false,\"max_parallel_tasks\":0,\"name\":\"Test 123\",\"type\":\"\"},\"repositories\":[{\"git_branch\":\"master\",\"git_url\":\"git@example.com:test/test\",\"name\":\"Test\",\"ssh_key\":\"\"}],\"templates\":[{\"allow_override_args_in_task\":false,\"app\":\"\",\"autorun\":false,\"environment\":\"test\",\"inventory\":\"\",\"name\":\"Test\",\"playbook\":\"test.yml\",\"repository\":\"Test\",\"suppress_success_alerts\":false,\"survey_vars\":[],\"task_params\":{},\"type\":\"\",\"vaults\":[]}],\"views\":[]}", str)

	restoredBackup := &BackupFormat{}
	err = restoredBackup.Unmarshal(str)
	assert.NoError(t, err)
	assert.Equal(t, proj.Name, restoredBackup.Meta.Name)

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
	assert.Equal(t, proj.Name, restoredProj.Name)
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

	str, err := backup.Marshal()
	assert.NoError(t, err)

	var res map[string]any
	json.Unmarshal([]byte(str), &res)

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
      "type": "none"
    }
  ],
  "meta": {
    "alert": false,
    "max_parallel_tasks": 0,
    "name": "Test 123",
    "type": ""
  },
  "repositories": [],
  "schedules": [],
  "secret_storages": [
    {
      "name": "Test",
      "params": {},
      "readonly": false,
      "type": "vault"
    }
  ],
  "templates": [],
  "views": []
}`, str)

	restoredBackup := &BackupFormat{}
	err = restoredBackup.Unmarshal(str)
	assert.NoError(t, err)
	assert.Equal(t, proj.Name, restoredBackup.Meta.Name)

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

func TestBackupRestoreScheduleDuplicationFix(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := sql.CreateTestStore()

	// Create a simple project
	proj, err := store.CreateProject(db.Project{
		Name: "Schedule Test Project",
	})
	assert.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
		Name:      "Test Key",
	})
	assert.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test Repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	assert.NoError(t, err)

	// Create template for completeness but we'll use the backup format directly
	_, err = store.CreateTemplate(db.Template{
		Name:         "Test Template",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		Type:         db.TemplateTask,
	})
	assert.NoError(t, err)

	// Create a schedule manually and simulate a backup format that would create duplicates
	backupFormat := &BackupFormat{
		Meta: BackupMeta{
			Project: db.Project{
				Name: "Schedule Test Project",
			},
		},
		Templates: []BackupTemplate{
			{
				Template: db.Template{
					Name:         "Test Template",
					Playbook:     "test.yml",
					Type:         db.TemplateTask,
				},
				Repository: "Test Repo",
				Cron:       stringPtr("0 */4 * * *"), // This would trigger automatic schedule creation
			},
		},
		Schedules: []BackupSchedule{
			{
				Schedule: db.Schedule{
					Name:       "Explicit Schedule",
					CronFormat: "0 */4 * * *",
					Active:     true,
				},
				Template: "Test Template",
			},
		},
		Repositories: []BackupRepository{
			{
				Repository: db.Repository{
					Name:      "Test Repo",
					GitURL:    "git@example.com:test/test",
					GitBranch: "master",
				},
				SSHKey: stringPtr("Test Key"),
			},
		},
		Keys: []BackupAccessKey{
			{
				AccessKey: db.AccessKey{
					Name: "Test Key",
					Type: db.AccessKeyNone,
				},
			},
		},
	}

	// Create user for restore
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

	// Restore project - this should test our fix
	restoredProj, err := backupFormat.Restore(user, store)
	assert.NoError(t, err)
	assert.Equal(t, proj.Name, restoredProj.Name)

	// Check that only one schedule was created (not duplicated)
	restoredSchedules, err := store.GetProjectSchedules(restoredProj.ID, true)
	assert.NoError(t, err)
	
	// This test verifies our fix: should have exactly 1 schedule, not 2
	assert.Len(t, restoredSchedules, 1, "Expected exactly one schedule, but found %d. Duplication issue not fixed.", len(restoredSchedules))
	
	if len(restoredSchedules) > 0 {
		// Verify the schedule name is the explicit one, not a random one
		assert.Equal(t, "Explicit Schedule", restoredSchedules[0].Schedule.Name, "Schedule should have the explicit name, not a random one")
	}
}

func stringPtr(s string) *string {
	return &s
}
