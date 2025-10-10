package tasks

import (
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/pkg/ssh"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/stretchr/testify/assert"

	"github.com/semaphoreui/semaphore/db_lib"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
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

	store := bolt.CreateTestStore()
	keyInstaller := &KeyInstallerMock{}

	pool := CreateTaskPool(
		store,
		&MemoryTaskStateStore{},
		nil,
		&InventoryServiceMock{},
		nil,
		keyInstaller,
		&mockLogWriteService{},
	)

	go pool.Run()

	var task db.Task

	var err error

	db.StoreSession(store, "", func() {
		task, err = store.CreateTask(db.Task{}, 0)
	})

	if err != nil {
		t.Fatal(err)
	}

	taskRunner := TaskRunner{
		Task:         task,
		pool:         &pool,
		keyInstaller: keyInstaller,
	}
	taskRunner.job = &LocalJob{
		Task:         taskRunner.Task,
		Template:     taskRunner.Template,
		Inventory:    taskRunner.Inventory,
		Repository:   taskRunner.Repository,
		Environment:  taskRunner.Environment,
		Logger:       &taskRunner,
		KeyInstaller: keyInstaller,
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
	tsk.job = &LocalJob{
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

	dir := tsk.job.(*LocalJob).App.(*db_lib.AnsibleApp).GetPlaybookDir()
	if dir != "/tmp/project_0/repository_0_template_0/deploy" {
		t.Fatal("Invalid playbook dir: " + dir)
	}
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
	tsk.job = &LocalJob{
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

	dir := tsk.job.(*LocalJob).App.(*db_lib.AnsibleApp).GetPlaybookDir()
	if dir != "/tmp/project_0/repository_0_template_0/deploy" {
		t.Fatal("Invalid playbook dir: " + dir)
	}
}

func TestPopulateDetails(t *testing.T) {
	store := bolt.CreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   &inv.ID,
		EnvironmentID: &env.ID,
	})

	if err != nil {
		t.Fatal(err)
	}

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
	tsk.job = &LocalJob{
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
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `{"author":"Denis","comment":"Just do it!","time":"2021-11-02"}`, tsk.Environment.JSON)

}

func TestPopulateDetailsInventory(t *testing.T) {
	store := bolt.CreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	if err != nil {
		t.Fatal(err)
	}
	inv2, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        2,
	})
	if err != nil {
		t.Fatal(err)
	}
	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   &inv.ID,
		EnvironmentID: &env.ID,
		TaskParams: map[string]any{
			"allow_override_inventory": true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

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
	tsk.job = &LocalJob{
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
	if err != nil {
		t.Fatal(err)
	}

	//if tsk.Inventory.ID != 2 {
	//	t.Fatal(err)
	//}
}

func TestPopulateDetailsInventory1(t *testing.T) {
	store := bolt.CreateTestStore()

	proj, err := store.CreateProject(db.Project{})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		ID:        1,
	})
	if err != nil {
		t.Fatal(err)
	}
	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "test",
		JSON:      `{"author": "Denis", "comment": "Hello, World!"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   &inv.ID,
		EnvironmentID: &env.ID,
	})

	if err != nil {
		t.Fatal(err)
	}

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
	tsk.job = &LocalJob{
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
	if err != nil {
		t.Fatal(err)
	}

	//if tsk.Inventory.ID != 1 {
	//	t.Fatal(err)
	//}
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
	tsk.job = &LocalJob{
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

	args, _, err := tsk.job.(*LocalJob).getPlaybookArgs("", nil)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} test.yml" {
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
	tsk.job = &LocalJob{
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

	args, _, err := tsk.job.(*LocalJob).getPlaybookArgs("", nil)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} test.yml" {
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

	tsk.job = &LocalJob{
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

	args, _, err := tsk.job.(*LocalJob).getPlaybookArgs("", nil)

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/project_0/inventory_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"commit_hash\":null,\"commit_message\":\"\",\"id\":0,\"inventory_id\":0,\"inventory_name\":\"\",\"repository_id\":0,\"repository_name\":\"\",\"url\":null,\"username\":\"\"}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestCheckTmpDir(t *testing.T) {
	//It should be able to create a random dir in /tmp
	dirName := path.Join(os.TempDir(), util.RandString(rand.Intn(10-4)+4))
	err := checkTmpDir(dirName)
	if err != nil {
		t.Fatal(err)
	}

	//checking again for this directory should return no error, as it exists
	err = checkTmpDir(dirName)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chmod(dirName, os.FileMode(int(0550)))
	if err != nil {
		t.Fatal(err)
	}

	//nolint: vetshadow
	if stat, err := os.Stat(dirName); err != nil {
		t.Fatal(err)
	} else if stat.Mode() != os.FileMode(int(0550)) {
		// File System is not support 0550 mode, skip this test
		return
	}

	err = checkTmpDir(dirName + "/noway")
	if err == nil {
		t.Fatal("You should not be able to write in this folder, causing an error")
	}
	err = os.Remove(dirName)
	if err != nil {
		t.Log(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tsk.Environment.JSON, "{\"a\":11,\"b\":22,\"c\":33,\"d\":4}")
}
