package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func createTestBoltDb() BoltDb {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fn := "/tmp/test_semaphore_db_" + strconv.Itoa(r.Int())
	return BoltDb{
		Filename: fn,
	}
}

func createTestStore() db.Store {
	store := createTestBoltDb()
	err := store.Connect()
	if err != nil {
		panic(err)
	}
	return &store
}

func TestTask_GetVersion(t *testing.T) {
	VERSION := "1.54.48"

	store := createTestStore()

	build, err := store.CreateTemplate(db.Template{
		ProjectID: 0,
		Type:      db.TemplateBuild,
		Alias:     "Build",
		Playbook:  "build.yml",
	})
	if err != nil {
		t.Fatal(err)
	}

	deploy, err := store.CreateTemplate(db.Template{
		ProjectID:       0,
		Type:            db.TemplateDeploy,
		BuildTemplateID: &build.ID,
		Alias:           "Deploy",
		Playbook:        "deploy.yml",
	})
	if err != nil {
		t.Fatal(err)
	}

	deploy2, err := store.CreateTemplate(db.Template{
		ProjectID:       0,
		Type:            db.TemplateDeploy,
		BuildTemplateID: &deploy.ID,
		Alias:           "Deploy2",
		Playbook:        "deploy2.yml",
	})
	if err != nil {
		t.Fatal(err)
	}

	buildTask, err := store.CreateTask(db.Task{
		ProjectID:  0,
		TemplateID: build.ID,
		Version:    &VERSION,
	})
	if err != nil {
		t.Fatal(err)
	}

	deployTask, err := store.CreateTask(db.Task{
		ProjectID:   0,
		TemplateID:  deploy.ID,
		BuildTaskID: &buildTask.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	deploy2Task, err := store.CreateTask(db.Task{
		ProjectID:   0,
		TemplateID:  deploy2.ID,
		BuildTaskID: &deployTask.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	version := deployTask.GetIncomingVersion(store)
	if version == nil {
		t.Fatal()
		return
	}
	if *version != VERSION {
		t.Fatal()
		return
	}

	version = deploy2Task.GetIncomingVersion(store)
	if version == nil {
		t.Fatal()
		return
	}
	if *version != VERSION {
		t.Fatal()
		return
	}
}
