package tasks

import (
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/util"
)

func CreateBoltDB() db.Store {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fn := "/tmp/test_semaphore_db_" + strconv.Itoa(r.Int())
	store := bolt.BoltDb{
		Filename: fn,
	}
	return &store
}

func TestTaskRunnerRun(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	store := CreateBoltDB()

	pool := CreateTaskPool(store)

	go pool.Run()

	var task db.Task

	var err error

	db.StoreSession(store, "", func() {
		task, err = store.CreateTask(db.Task{})
	})

	if err != nil {
		t.Fatal(err)
	}

	taskRunner := TaskRunner{
		task: task,
		pool: &pool,
	}

	taskRunner.run()
}

func TestGetRepoPath(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		template: db.Template{
			Playbook: "deploy/test.yml",
		},
	}

	dir := tsk.getPlaybookDir()
	if dir != "/tmp/repository_0_0/deploy" {
		t.Fatal("Invalid playbook dir: " + dir)
	}
}

func TestGetRepoPath_whenStartsWithSlash(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		template: db.Template{
			Playbook: "/deploy/test.yml",
		},
	}

	dir := tsk.getPlaybookDir()
	if dir != "/tmp/repository_0_0/deploy" {
		t.Fatal("Invalid playbook dir: " + dir)
	}
}

func TestPopulateDetails(t *testing.T) {
	store := CreateBoltDB()
	store.Connect("")

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
		InventoryID:   inv.ID,
		EnvironmentID: &env.ID,
	})

	if err != nil {
		t.Fatal(err)
	}

	pool := TaskPool{store: store}

	tsk := TaskRunner{
		pool: &pool,
		task: db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   proj.ID,
			Environment: `{"comment": "Just do it!", "time": "2021-11-02"}`,
		},
	}

	err = tsk.populateDetails()
	if err != nil {
		t.Fatal(err)
	}
	if tsk.environment.JSON != `{"author":"Denis","comment":"Hello, World!","time":"2021-11-02"}` {
		t.Fatal(err)
	}
}

func TestTaskGetPlaybookArgs(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
			},
			Type: db.InventoryStatic,
		},
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --private-key=/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"id\":0}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs2(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		task: db.Task{},
		inventory: db.Inventory{
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
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"id\":0}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs3(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := TaskRunner{
		task: db.Task{},
		inventory: db.Inventory{
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
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{\"id\":0}}} test.yml" {
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
