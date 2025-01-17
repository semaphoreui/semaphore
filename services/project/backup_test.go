package project

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
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

	store := bolt.CreateTestStore()

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
	assert.Equal(t, "{\"environments\":[{\"json\":\"{\\\"author\\\": \\\"Denis\\\", \\\"comment\\\": \\\"Hello, World!\\\"}\",\"name\":\"test\"}],\"integration_aliases\":[],\"integrations\":[],\"inventories\":[{\"inventory\":\"\",\"name\":\"\",\"type\":\"\"}],\"keys\":[{\"name\":\"\",\"type\":\"none\"}],\"meta\":{\"alert\":false,\"max_parallel_tasks\":0,\"name\":\"Test 123\",\"type\":\"\"},\"repositories\":[{\"git_branch\":\"master\",\"git_url\":\"git@example.com:test/test\",\"name\":\"Test\",\"ssh_key\":\"\"}],\"templates\":[{\"allow_override_args_in_task\":false,\"app\":\"\",\"autorun\":false,\"environment\":\"test\",\"inventory\":\"\",\"name\":\"Test\",\"playbook\":\"test.yml\",\"repository\":\"Test\",\"suppress_success_alerts\":false,\"survey_vars\":[],\"task_params\":{},\"type\":\"\",\"vaults\":[]}],\"views\":[]}", str)

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
