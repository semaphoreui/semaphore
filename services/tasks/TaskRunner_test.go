package tasks

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/ssh"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/semaphoreui/semaphore/db_lib"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

type KeyInstallerMock struct {
}

func (s *KeyInstallerMock) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error) {
	return ssh.AccessKeyInstallation{}, nil
}

type InventoryServiceMock struct {
}

func (s *InventoryServiceMock) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	return db.Inventory{}, nil
}

type EncryptionServiceMock struct {
}

func (s *EncryptionServiceMock) RekeyAccessKeys(oldKey string) (err error) {
	return nil
}

func (s *EncryptionServiceMock) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (s *EncryptionServiceMock) SerializeSecret(key *db.AccessKey) error {
	return nil
}

func (s *EncryptionServiceMock) DeserializeSecret(key *db.AccessKey) error {
	return nil
}

func (s *EncryptionServiceMock) FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error {
	return nil
}

func (s *EncryptionServiceMock) CreateTaskSurveySecrets(projectID int, taskID int, secrets string, expireAt time.Time) error {
	return nil
}

func (s *EncryptionServiceMock) GetTaskSurveySecrets(projectID int, taskID int) (string, error) {
	return "", nil
}

func (s *EncryptionServiceMock) DeleteTaskSurveySecrets(projectID int, taskID int) error {
	return nil
}

type mockLogWriteService struct {
}

func (l *mockLogWriteService) WriteEventLog(event pro_interfaces.EventLogRecord) error {
	return nil
}

func (l *mockLogWriteService) WriteTaskLog(task pro_interfaces.TaskLogRecord) error {
	return nil
}
func (l *mockLogWriteService) WriteResult(task any) error {
	return nil
}

func TestTaskRunnerRun(t *testing.T) {

	store := sql.InitConfigCreateTestStore()
	keyInstaller := &KeyInstallerMock{}

	pool := CreateTaskPool(
		store,
		&MemoryTaskStateStore{},
		nil,
		&InventoryServiceMock{},
		&EncryptionServiceMock{},
		keyInstaller,
		&mockLogWriteService{},
		nil,
		nil,
	)

	go pool.Run()
	// Stop the pool's background loops (notably the runner-task reconcile loop)
	// before the test returns, so the loop does not outlive this test and race
	// with later tests that mutate the util.Config global.
	t.Cleanup(pool.Stop)

	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Test",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &inv.ID,
	})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		ProjectID:  proj.ID,
		TemplateID: tpl.ID,
	}, 0)
	require.NoError(t, err)

	taskRunner := TaskRunner{
		Task:         task,
		pool:         &pool,
		keyInstaller: keyInstaller,
	}
	taskRunner.job = &LocalExecutor{
		Task:         taskRunner.Task,
		Template:     taskRunner.Template,
		Inventory:    taskRunner.Inventory,
		Repository:   taskRunner.Repository,
		Environment:  taskRunner.Environment,
		Logger:       &taskRunner,
		KeyInstaller: keyInstaller,
		RepoLock:     &KeyLock{},
		App: &db_lib.AnsibleApp{
			Template:   taskRunner.Template,
			Repository: taskRunner.Repository,
			Logger:     &taskRunner,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &taskRunner,
				TemplateID: taskRunner.Template.ID,
				Repository: taskRunner.Repository,
			},
		},
	}
	taskRunner.run()
}

func TestGetRepoPath(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		Task: db.Task{},
		Inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		Template: db.Template{
			Playbook: "deploy/test.yml",
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	dir := tsk.job.(*LocalExecutor).App.(*db_lib.AnsibleApp).GetPlaybookDir()
	assert.Equal(t, "/tmp/project_0/repository_0_template_0_da39a3ee5e6b4b0d3255bfef95601890/deploy", dir)
}

func TestGetRepoPath_whenStartsWithSlash(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		Task: db.Task{},
		Inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		Template: db.Template{
			Playbook: "/deploy/test.yml",
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	dir := tsk.job.(*LocalExecutor).App.(*db_lib.AnsibleApp).GetPlaybookDir()
	assert.Equal(t, "/tmp/project_0/repository_0_template_0_da39a3ee5e6b4b0d3255bfef95601890/deploy", dir)
}

func TestPopulateDetails(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
	})
	require.NoError(t, err)

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:           "Test",
		Playbook:       "test.yml",
		ProjectID:      proj.ID,
		RepositoryID:   repo.ID,
		InventoryID:    &inv.ID,
		EnvironmentIDs: []int{env.ID},
	})
	require.NoError(t, err)

	pool := TaskPool{
		store:             store,
		inventoryService:  &InventoryServiceMock{},
		encryptionService: &EncryptionServiceMock{},
	}

	tsk := TaskRunner{
		pool: &pool,
		Task: db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   proj.ID,
			Environment: `{"comment": "Just do it!", "time": "2021-11-02"}`,
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	err = tsk.populateDetails()
	require.NoError(t, err)

	assert.Equal(t, `{"author":"Denis","comment":"Just do it!","time":"2021-11-02"}`, tsk.Environment.JSON)

}

func TestPopulateDetailsInventory(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	require.NoError(t, err)
	inv2, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        2,
	})
	require.NoError(t, err)
	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:           "Test",
		Playbook:       "test.yml",
		ProjectID:      proj.ID,
		RepositoryID:   repo.ID,
		InventoryID:    &inv.ID,
		EnvironmentIDs: []int{env.ID},
		TaskParams: map[string]any{
			"allow_override_inventory": true,
		},
	})
	require.NoError(t, err)

	pool := TaskPool{
		store:             store,
		inventoryService:  &InventoryServiceMock{},
		encryptionService: &EncryptionServiceMock{},
	}

	tsk := TaskRunner{
		pool: &pool,
		Task: db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   proj.ID,
			Environment: `{"comment": "Just do it!", "time": "2021-11-02"}`,
			InventoryID: &inv2.ID,
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	err = tsk.populateDetails()
	require.NoError(t, err)
}

func TestPopulateDetailsInventory1(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	require.NoError(t, err)
	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:           "Test",
		Playbook:       "test.yml",
		ProjectID:      proj.ID,
		RepositoryID:   repo.ID,
		InventoryID:    &inv.ID,
		EnvironmentIDs: []int{env.ID},
	})
	require.NoError(t, err)

	pool := TaskPool{
		store:             store,
		inventoryService:  &InventoryServiceMock{},
		encryptionService: &EncryptionServiceMock{},
	}

	tsk := TaskRunner{
		pool: &pool,
		Task: db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   proj.ID,
			Environment: `{"comment": "Just do it!", "time": "2021-11-02"}`,
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	err = tsk.populateDetails()
	require.NoError(t, err)
}

func TestTaskGetPlaybookArgs(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		Task: db.Task{},
		Inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		Template: db.Template{
			Playbook: "test.yml",
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	args, _, err := tsk.job.(*LocalExecutor).getPlaybookArgs("", nil)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "--inventory /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} /tmp/project_0/repository_0_template_0_da39a3ee5e6b4b0d3255bfef95601890/test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs2(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		Task: db.Task{},
		Inventory: db.Inventory{
			Type:     db.InventoryStatic,
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{
					Password: "123456",
					Login:    "root",
				},
			},
		},
		Template: db.Template{
			Playbook: "test.yml",
		},
	}
	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	args, _, err := tsk.job.(*LocalExecutor).getPlaybookArgs("", nil)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "--inventory /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} /tmp/project_0/repository_0_template_0_da39a3ee5e6b4b0d3255bfef95601890/test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs3(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		Task: db.Task{},
		Inventory: db.Inventory{
			Type:        db.InventoryStatic,
			BecomeKeyID: &inventoryID,
			BecomeKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{
					Password: "123456",
					Login:    "root",
				},
			},
		},
		Template: db.Template{
			Playbook: "test.yml",
		},
	}

	tsk.job = &LocalExecutor{
		Task:        tsk.Task,
		Template:    tsk.Template,
		Inventory:   tsk.Inventory,
		Repository:  tsk.Repository,
		Environment: tsk.Environment,
		Logger:      &tsk,
		App: &db_lib.AnsibleApp{
			Template:   tsk.Template,
			Repository: tsk.Repository,
			Logger:     &tsk,
			Playbook: &db_lib.AnsiblePlaybook{
				Logger:     &tsk,
				TemplateID: tsk.Template.ID,
				Repository: tsk.Repository,
			},
		},
	}

	args, _, err := tsk.job.(*LocalExecutor).getPlaybookArgs("", nil)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "--inventory /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} /tmp/project_0/repository_0_template_0_da39a3ee5e6b4b0d3255bfef95601890/test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestCheckTmpDir(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	// It should be able to create a new dir inside the temp dir
	dirName := path.Join(t.TempDir(), "tmp")
	require.NoError(t, checkTmpDir(dirName))

	// checking again for this directory should return no error, as it exists
	require.NoError(t, checkTmpDir(dirName))

	require.NoError(t, os.Chmod(dirName, os.FileMode(0550)))

	stat, err := os.Stat(dirName)
	require.NoError(t, err)
	if stat.Mode() != os.FileMode(0550) {
		t.Skip("file system does not support 0550 mode")
	}

	assert.Error(t, checkTmpDir(dirName+"/noway"), "should not be able to write in this folder")
}

func TestTaskRunner_populateTaskEnvironment(t *testing.T) {
	tsk := TaskRunner{
		Task: db.Task{
			Environment: "{\"a\":11, \"b\": 22, \"c\": 33}",
		},
		Environment: db.Environment{
			JSON: "{\"a\":1, \"d\": 4}",
		},
	}

	err := tsk.populateTaskEnvironment()
	require.NoError(t, err)

	assert.Equal(t, "{\"a\":11,\"b\":22,\"c\":33,\"d\":4}", tsk.Environment.JSON)
}
