package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ansibleTaskRepoFake struct {
	db.AnsibleTaskRepository
	calls     int
	projectID int
	taskID    int
	inventory *int
	hosts     []db.AnsibleTaskHost
}

func (f *ansibleTaskRepoFake) ReplaceAnsibleTaskHosts(projectID int, taskID int, inventoryID *int, hosts []db.AnsibleTaskHost) error {
	f.calls++
	f.projectID = projectID
	f.taskID = taskID
	f.inventory = inventoryID
	f.hosts = hosts
	return nil
}

func newHostsPool(repo *ansibleTaskRepoFake) (*TaskPool, *TaskRunner) {
	tsk := &TaskRunner{Task: db.Task{ID: 42, ProjectID: 7}}
	return &TaskPool{ansibleTaskRepo: repo}, tsk
}

func TestTaskPool_StoreTaskHosts(t *testing.T) {
	repo := &ansibleTaskRepoFake{}
	pool, tsk := newHostsPool(repo)
	document := `{"version":1,"hosts":{"web-02":{"ok":1},"web-01":{"ok":2,"failed":1}}}`

	require.NoError(t, pool.StoreTaskHosts(tsk, document))

	assert.Equal(t, 1, repo.calls)
	assert.Equal(t, 7, repo.projectID)
	assert.Equal(t, 42, repo.taskID)
	require.Len(t, repo.hosts, 2)
	assert.Equal(t, "web-01", repo.hosts[0].Host)
	assert.Equal(t, "web-02", repo.hosts[1].Host)
}

func TestTaskPool_StoreTaskHosts_Inventory(t *testing.T) {
	document := `{"version":1,"hosts":{"web-01":{"ok":1}}}`

	t.Run("task inventory is passed", func(t *testing.T) {
		repo := &ansibleTaskRepoFake{}
		pool, tsk := newHostsPool(repo)
		tsk.Inventory.ID = 5

		require.NoError(t, pool.StoreTaskHosts(tsk, document))

		require.NotNil(t, repo.inventory)
		assert.Equal(t, 5, *repo.inventory)
	})

	t.Run("recorded inventory wins", func(t *testing.T) {
		repo := &ansibleTaskRepoFake{}
		pool, tsk := newHostsPool(repo)
		recorded := 3
		tsk.Task.EffectiveInventoryID = &recorded
		tsk.Inventory.ID = 5

		require.NoError(t, pool.StoreTaskHosts(tsk, document))

		require.NotNil(t, repo.inventory)
		assert.Equal(t, 3, *repo.inventory)
	})

	t.Run("no inventory gives nil", func(t *testing.T) {
		repo := &ansibleTaskRepoFake{}
		pool, tsk := newHostsPool(repo)

		require.NoError(t, pool.StoreTaskHosts(tsk, document))

		assert.Nil(t, repo.inventory)
	})
}

// idInventoryService returns an inventory carrying the requested id.
type idInventoryService struct {
	InventoryServiceMock
}

func (s *idInventoryService) GetInventory(projectID int, inventoryID int) (db.Inventory, error) {
	return db.Inventory{ID: inventoryID, ProjectID: projectID}, nil
}

func TestPopulateDetails_OverriddenInventoryReachesHostResults(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)
	templateInventory, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	taskInventory, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Test",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &templateInventory.ID,
		TaskParams:   map[string]any{"allow_override_inventory": true},
	})
	require.NoError(t, err)
	fake := &ansibleTaskRepoFake{}
	pool := &TaskPool{
		store:             store,
		inventoryService:  &idInventoryService{},
		encryptionService: &EncryptionServiceMock{},
		ansibleTaskRepo:   fake,
	}
	tsk := &TaskRunner{
		pool: pool,
		Task: db.Task{ID: 42, TemplateID: tpl.ID, ProjectID: proj.ID, InventoryID: &taskInventory.ID},
	}
	require.NoError(t, tsk.populateDetails())

	require.NoError(t, pool.StoreTaskHosts(tsk, `{"version":1,"hosts":{"web-01":{"ok":1}}}`))

	require.NotNil(t, fake.inventory)
	assert.Equal(t, taskInventory.ID, *fake.inventory)
	assert.NotEqual(t, templateInventory.ID, *fake.inventory)
}

// A runner rebuilt from the database reads the current template, which may
// name another inventory than the one the task ran with.
func TestHydratedTask_HostResultsGoToRecordedInventory(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)
	runInventory, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	newInventory, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Test",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &runInventory.ID,
	})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: proj.ID, TemplateID: tpl.ID}, 0)
	require.NoError(t, err)
	require.NoError(t, store.UpdateTaskEffectiveInventory(proj.ID, task.ID, &runInventory.ID))
	tpl.InventoryID = &newInventory.ID
	require.NoError(t, store.UpdateTemplate(tpl))
	fake := &ansibleTaskRepoFake{}
	pool := &TaskPool{
		store:             store,
		inventoryService:  &idInventoryService{},
		encryptionService: &EncryptionServiceMock{},
		ansibleTaskRepo:   fake,
	}

	tsk, err := pool.HydrateTaskRunnerFromDB(task.ID)
	require.NoError(t, err)
	require.NoError(t, pool.StoreTaskHosts(tsk, `{"version":1,"hosts":{"web-01":{"ok":1}}}`))

	require.NotNil(t, fake.inventory)
	assert.Equal(t, runInventory.ID, *fake.inventory)
}
