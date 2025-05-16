package tasks

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createTask(store db.Store, pool *TaskPool, templateId int, job *TestJob) (runner *TaskRunner, err error) {

	var task db.Task
	db.StoreSession(store, "", func() {
		task, err = store.CreateTask(db.Task{TemplateID: templateId, ProjectID: 1}, 0)
	})

	if err != nil {
		return
	}

	runner = &TaskRunner{
		Task: task,
		pool: pool,
	}

	runner.job = job
	return
}

func TestTaskPoolRun(t *testing.T) {
	const maxParallelTasks = 5
	util.Config = &util.ConfigType{
		TmpPath:          "/tmp",
		MaxParallelTasks: maxParallelTasks,
	}

	store := bolt.CreateTestStore()
	pool := CreateTaskPool(store)

	go pool.Run()

	const size = 10

	proj, err := store.CreateProject(db.Project{MaxParallelTasks: 0})
	if err != nil {
		t.Fatal(err)
	}
	log.Info(proj)

	for i := 0; i < size; i++ {
		taskRunner, err := createTask(store, &pool, i, &TestJob{DurationMs: 300})
		done := make(chan bool)
		pool.queueEvents <- PoolEvent{EventTypeNew, taskRunner, done}
		<-done
		if err != nil {
			t.Fatal(err)
		}
	}

	assert.Equal(t, size-maxParallelTasks, len(pool.Queue))
}
