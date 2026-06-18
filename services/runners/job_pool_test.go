package runners

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJob(id int) *job {
	return &job{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: id}},
		taskID: id,
	}
}

func TestJobPool_QueueOrder(t *testing.T) {
	util.Config = &util.ConfigType{
		Runner: &util.RunnerConfig{
			Executor: &util.ExecutorConfig{},
		},
	}

	p := NewJobPool(nil)

	assert.Equal(t, 0, p.queueLen())

	p.enqueue(newTestJob(1))
	p.enqueue(newTestJob(2))
	p.enqueue(newTestJob(3))

	assert.Equal(t, 3, p.queueLen())
	assert.True(t, p.existsInQueue(2))
	assert.False(t, p.existsInQueue(99))

	// FIFO order.
	j, ok := p.dequeue()
	require.True(t, ok)
	assert.Equal(t, 1, j.taskID)

	j, ok = p.dequeue()
	require.True(t, ok)
	assert.Equal(t, 2, j.taskID)

	j, ok = p.dequeue()
	require.True(t, ok)
	assert.Equal(t, 3, j.taskID)

	_, ok = p.dequeue()
	assert.False(t, ok)
}

func TestJobPool_RunningJobsLifecycle(t *testing.T) {
	p := NewJobPool(nil)

	assert.Equal(t, 0, p.runningJobsCount())
	assert.Nil(t, p.getRunningJob(1))

	rj := &runningJob{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: 1}},
		status: task_logger.TaskRunningStatus,
	}
	p.addRunningJob(1, rj)

	assert.Equal(t, 1, p.runningJobsCount())
	assert.Same(t, rj, p.getRunningJob(1))

	snapshot := p.snapshotRunningJobs()
	assert.Len(t, snapshot, 1)
	// Snapshot is a copy: deleting from it must not affect the pool.
	delete(snapshot, 1)
	assert.Equal(t, 1, p.runningJobsCount())

	p.deleteRunningJob(1)
	assert.Equal(t, 0, p.runningJobsCount())
	assert.Nil(t, p.getRunningJob(1))
}

func TestJobPool_HasRunningJobs(t *testing.T) {
	p := NewJobPool(nil)
	assert.False(t, p.hasRunningJobs())

	p.addRunningJob(1, &runningJob{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: 1}},
		status: task_logger.TaskSuccessStatus, // finished
	})
	assert.False(t, p.hasRunningJobs())

	p.addRunningJob(2, &runningJob{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: 2}},
		status: task_logger.TaskRunningStatus, // not finished
	})
	assert.True(t, p.hasRunningJobs())
}

func TestJobPool_ApplyTerminatedJobs(t *testing.T) {
	p := NewJobPool(nil)

	// Running job: must be emergency stopped and removed.
	lj := &tasks.LocalExecutor{Task: db.Task{ID: 1}}
	rj := &runningJob{job: lj, status: task_logger.TaskRunningStatus}
	lj.Logger = rj
	p.addRunningJob(1, rj)

	// Already finished job: must be removed without a status change or kill.
	lj2 := &tasks.LocalExecutor{Task: db.Task{ID: 2}}
	rj2 := &runningJob{job: lj2, status: task_logger.TaskSuccessStatus}
	lj2.Logger = rj2
	p.addRunningJob(2, rj2)

	// Unknown task ID (99) must be a no-op.
	p.applyTerminatedJobs([]int{1, 2, 99})

	assert.Equal(t, task_logger.TaskStoppedStatus, rj.getStatus())
	assert.True(t, lj.IsKilled())

	assert.Equal(t, task_logger.TaskSuccessStatus, rj2.getStatus())
	assert.False(t, lj2.IsKilled())

	assert.Equal(t, 0, p.runningJobsCount())
}

func TestJobPool_CheckNewJobsExecutorErrorUsesTaskProjectID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := json.NewEncoder(w).Encode(RunnerState{
			NewJobs: []JobData{{
				Task: db.Task{
					ID:        42,
					ProjectID: 7,
				},
			}},
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	previousConfig := util.Config
	util.Config = &util.ConfigType{
		WebHost: server.URL,
		Runner: &util.RunnerConfig{
			Token:      "test-token",
			Connection: &util.RunnerConnectionConfig{},
		},
	}
	defer func() {
		util.Config = previousConfig
	}()

	p := &JobPool{
		runningJobs: make(map[int]*runningJob),
		queue:       make([]*job, 0),
		startedAt:   time.Now(),
	}

	require.NotPanics(t, p.checkNewJobs)
	assert.Equal(t, 0, p.queueLen())
}

// TestJobPool_ConcurrentAccess models the three actors that touch the pool
// concurrently in production: the Run loop (dequeue + addRunningJob), the poll
// goroutine (snapshot + delete + enqueue) and status readers. Run with -race to
// catch concurrent map/slice access, which otherwise aborts the process with
// "fatal error: concurrent map read and map write".
func TestJobPool_ConcurrentAccess(t *testing.T) {
	p := NewJobPool(nil)

	const workers = 8
	const iterations = 300

	var wg sync.WaitGroup
	start := make(chan struct{})

	// Running-jobs map mutators/readers.
	for w := 0; w < workers; w++ {
		base := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				id := base*iterations + i
				p.addRunningJob(id, &runningJob{
					job:    &tasks.LocalExecutor{Task: db.Task{ID: id}},
					status: task_logger.TaskRunningStatus,
				})
				p.getRunningJob(id)
				p.runningJobsCount()
				p.snapshotRunningJobs()
				p.hasRunningJobs()
				p.deleteRunningJob(id)
			}
		}()
	}

	// Queue mutators/readers.
	for w := 0; w < workers; w++ {
		base := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				id := base*iterations + i
				p.enqueue(newTestJob(id))
				p.existsInQueue(id)
				p.queueLen()
				p.dequeue()
			}
		}()
	}

	close(start)
	wg.Wait()
}

func TestJobPool_checkNewJobs_ExecutorErrorWithoutCacheCleanProjectID(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := RunnerState{
			NewJobs: []JobData{
				{
					Task:        db.Task{ID: 1, ProjectID: 42, TemplateID: 1},
					Template:    db.Template{ID: 1, App: db.AppAnsible},
					Inventory:   db.Inventory{ID: 1},
					Repository:  db.Repository{ID: 1},
					Environment: db.Environment{ID: 1},
				},
			},
			AccessKeys: map[int]db.AccessKey{},
		}
		require.NoError(t, json.NewEncoder(w).Encode(state))
	}))
	t.Cleanup(srv.Close)

	util.Config = &util.ConfigType{
		WebHost: srv.URL,
		Runner: &util.RunnerConfig{
			Token:      "test-token",
			Executor:   &util.ExecutorConfig{},
			Connection: &util.RunnerConnectionConfig{},
		},
	}

	p := NewJobPool(nil)
	p.provider = nil // simulate OSS k8s/docker stub or failed provider init

	require.NotPanics(t, func() {
		p.checkNewJobs()
	})
	assert.Equal(t, 0, p.queueLen())
}
