package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/util"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPopulateDetails(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fn := "/tmp/test_semaphore_db_" + strconv.Itoa(r.Int())
	store := bolt.BoltDb{
		Filename: fn,
	}
	err := store.Connect()
	if err != nil {
		t.Fatal(err)
	}

	proj, err := store.CreateProject(db.Project{})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
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
		Alias:         "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		RepositoryID:  repo.ID,
		InventoryID:   inv.ID,
		EnvironmentID: &env.ID,
	})

	if err != nil {
		t.Fatal(err)
	}

	tsk := task{
		store:     &store,
		projectID: proj.ID,
		task: db.Task{
			TemplateID:  tpl.ID,
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

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID:   12345,
				Type: db.AccessKeySSH,
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
	if res != "-i /tmp/inventory_0 --private-key=/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs2(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
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
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs3(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
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
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_0 --extra-vars {\"semaphore_vars\":{\"task_details\":{}}} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestCheckTmpDir(t *testing.T) {
	//It should be able to create a random dir in /tmp
	dirName := os.TempDir() + "/" + randString(rand.Intn(10-4)+4)
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

//HELPERS

//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
