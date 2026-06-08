package runners

import (
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJob(id int) *job {
	return &job{
		job: &tasks.LocalJob{Task: db.Task{ID: id}},
	}
}

func TestJobPool_QueueOrder(t *testing.T) {
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
	assert.Equal(t, 1, j.job.Task.ID)

	j, ok = p.dequeue()
	require.True(t, ok)
	assert.Equal(t, 2, j.job.Task.ID)

	j, ok = p.dequeue()
	require.True(t, ok)
	assert.Equal(t, 3, j.job.Task.ID)

	_, ok = p.dequeue()
	assert.False(t, ok)
}

func TestJobPool_RunningJobsLifecycle(t *testing.T) {
	p := NewJobPool(nil)

	assert.Equal(t, 0, p.runningJobsCount())
	assert.Nil(t, p.getRunningJob(1))

	rj := &runningJob{
		job:    &tasks.LocalJob{Task: db.Task{ID: 1}},
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
		job:    &tasks.LocalJob{Task: db.Task{ID: 1}},
		status: task_logger.TaskSuccessStatus, // finished
	})
	assert.False(t, p.hasRunningJobs())

	p.addRunningJob(2, &runningJob{
		job:    &tasks.LocalJob{Task: db.Task{ID: 2}},
		status: task_logger.TaskRunningStatus, // not finished
	})
	assert.True(t, p.hasRunningJobs())
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
					job:    &tasks.LocalJob{Task: db.Task{ID: id}},
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
