package tasks

import (
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

type spyTaskStateStore struct {
	*MemoryTaskStateStore
	tryClaimCalls int
}

func newSpyTaskStateStore() *spyTaskStateStore {
	return &spyTaskStateStore{
		MemoryTaskStateStore: NewMemoryTaskStateStore(),
	}
}

// TryClaim returns false to ensure tests don't actually start tasks; we only want to
// observe whether the queue loop attempted to claim a task.
func (s *spyTaskStateStore) TryClaim(_ int) bool {
	s.tryClaimCalls++
	return false
}

func TestTaskPool_RequeuedEventCleansRunningStateAndSkipsImmediateRetry(t *testing.T) {
	// Ensure util.Config is non-nil for p.blocks() checks.
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{MaxParallelTasks: 0}

	store := bolt.CreateTestStore()
	proj, err := store.CreateProject(db.Project{})
	assert.NoError(t, err)

	state := newSpyTaskStateStore()

	pool := TaskPool{
		queueEvents: make(chan PoolEvent),
		state:       state,
		store:       store,
	}

	tr := &TaskRunner{
		Task: db.Task{
			ID:         42,
			ProjectID:  proj.ID,
			TemplateID: 7,
			Status:     task_logger.TaskWaitingStatus,
		},
		Template: db.Template{
			ID:   7,
			Name: "Test Template",
		},
		Alias: "alias-42",
	}

	// Simulate a task that was marked as running and then re-queued (the state that
	// exists right before EventTypeRequeued is handled).
	state.SetRunning(tr)
	state.AddActive(tr.Task.ProjectID, tr)
	state.SetAlias(tr.Alias, tr)
	state.Enqueue(tr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pool.handleQueue()
	}()

	pool.queueEvents <- PoolEvent{EventTypeRequeued, tr}
	close(pool.queueEvents)
	wg.Wait()

	assert.Equal(t, 0, state.RunningCount(), "requeued task must be removed from running set")
	assert.Equal(t, 0, state.ActiveCount(tr.Task.ProjectID), "requeued task must be removed from active-by-project set")
	assert.Nil(t, state.GetByAlias(tr.Alias), "requeued task alias mapping must be cleared")
	assert.Equal(t, 1, state.QueueLen(), "requeued task must remain queued")
	assert.Equal(t, 0, state.tryClaimCalls, "requeued task should not be immediately retried in the same queue pass")
}


