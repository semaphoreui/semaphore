package tasks

import (
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRunner(id, projectID, templateID int) *TaskRunner {
	runnerID := 7
	return &TaskRunner{
		Task: db.Task{
			ID:         id,
			ProjectID:  projectID,
			TemplateID: templateID,
			Status:     task_logger.TaskRunningStatus,
			RunnerID:   &runnerID,
		},
		Username: "admin",
		Alias:    "",
	}
}

func TestMemoryTaskStateStore_Snapshot(t *testing.T) {
	s := NewMemoryTaskStateStore()

	queued := newTestRunner(1, 10, 100)
	s.Enqueue(queued)

	running := newTestRunner(2, 10, 100)
	s.SetRunning(running)

	active := newTestRunner(3, 20, 200)
	s.AddActive(20, active)

	aliased := newTestRunner(4, 30, 300)
	aliased.Alias = "my-alias"
	s.SetAlias("my-alias", aliased)

	snap := s.Snapshot()

	require.Len(t, snap.Queue, 1)
	assert.Equal(t, 1, snap.Queue[0].TaskID)
	assert.Equal(t, 10, snap.Queue[0].ProjectID)
	assert.Equal(t, 100, snap.Queue[0].TemplateID)
	assert.Equal(t, "running", snap.Queue[0].Status)
	assert.Equal(t, 7, snap.Queue[0].RunnerID)
	assert.Equal(t, "admin", snap.Queue[0].Username)

	require.Len(t, snap.Running, 1)
	assert.Equal(t, 2, snap.Running[0].TaskID)

	require.Contains(t, snap.ActiveByProj, 20)
	require.Len(t, snap.ActiveByProj[20], 1)
	assert.Equal(t, 3, snap.ActiveByProj[20][0].TaskID)

	require.Contains(t, snap.Aliases, "my-alias")
	assert.Equal(t, 4, snap.Aliases["my-alias"])

	// The memory store does not track claims.
	assert.Empty(t, snap.Claims)
}

func TestMemoryTaskStateStore_Snapshot_Empty(t *testing.T) {
	s := NewMemoryTaskStateStore()
	snap := s.Snapshot()

	assert.Empty(t, snap.Queue)
	assert.Empty(t, snap.Running)
	assert.Empty(t, snap.ActiveByProj)
	assert.Empty(t, snap.Aliases)
	assert.Empty(t, snap.Claims)
}

func TestMemoryTaskStateStore_ClearTasks(t *testing.T) {
	tests := []struct {
		name        string
		scope       ClearScope
		wantQueue   int
		wantRunning int
		wantActive  int
		wantAliases int
		wantDeleted int
	}{
		{
			name:        "clear queue only",
			scope:       ClearScope{Queue: true},
			wantQueue:   0,
			wantRunning: 1,
			wantActive:  1,
			wantAliases: 1,
			wantDeleted: 2,
		},
		{
			name:        "clear running only",
			scope:       ClearScope{Running: true},
			wantQueue:   2,
			wantRunning: 0,
			wantActive:  1,
			wantAliases: 1,
			wantDeleted: 1,
		},
		{
			name:        "clear all groups",
			scope:       ClearScope{Queue: true, Running: true, Active: true, Aliases: true},
			wantQueue:   0,
			wantRunning: 0,
			wantActive:  0,
			wantAliases: 0,
			wantDeleted: 5,
		},
		{
			name:        "clear nothing",
			scope:       ClearScope{},
			wantQueue:   2,
			wantRunning: 1,
			wantActive:  1,
			wantAliases: 1,
			wantDeleted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemoryTaskStateStore()
			s.Enqueue(newTestRunner(1, 10, 100))
			s.Enqueue(newTestRunner(2, 10, 100))
			s.SetRunning(newTestRunner(3, 10, 100))
			s.AddActive(20, newTestRunner(4, 20, 200))
			aliased := newTestRunner(5, 30, 300)
			s.SetAlias("a", aliased)

			res, err := s.ClearTasks(tt.scope)
			require.NoError(t, err)

			assert.Equal(t, tt.wantQueue, s.QueueLen())
			assert.Equal(t, tt.wantRunning, s.RunningCount())
			assert.Equal(t, tt.wantActive, s.ActiveCount(20))
			if tt.wantAliases == 0 {
				assert.Nil(t, s.GetByAlias("a"))
			} else {
				assert.NotNil(t, s.GetByAlias("a"))
			}
			assert.Equal(t, tt.wantDeleted, res.DeletedKeys)
		})
	}
}

func TestMemoryTaskStateStore_ClearTasks_PerGroupCounts(t *testing.T) {
	s := NewMemoryTaskStateStore()
	s.Enqueue(newTestRunner(1, 10, 100))
	s.Enqueue(newTestRunner(2, 10, 100))
	s.SetRunning(newTestRunner(3, 10, 100))
	s.AddActive(20, newTestRunner(4, 20, 200))
	s.AddActive(20, newTestRunner(5, 20, 200))

	res, err := s.ClearTasks(ClearScope{Queue: true, Running: true, Active: true})
	require.NoError(t, err)

	assert.Equal(t, 2, res.PerGroup["queue"])
	assert.Equal(t, 1, res.PerGroup["running"])
	assert.Equal(t, 2, res.PerGroup["active"])
	assert.Equal(t, 5, res.DeletedKeys)
}

func TestMemoryTaskStateStore_ClaimAndDequeue(t *testing.T) {
	s := NewMemoryTaskStateStore()
	s.Enqueue(newTestRunner(1, 10, 100))
	s.Enqueue(newTestRunner(2, 10, 100))

	// First claim succeeds: task leaves the queue and enters the running set.
	assert.True(t, s.ClaimAndDequeue(1))
	assert.Equal(t, 1, s.QueueLen(), "claimed task must be removed from the queue")
	assert.Equal(t, 1, s.RunningCount(), "claimed task must be added to the running set")
	assert.Nil(t, s.QueueGet(1), "only one task should remain queued")

	// Second claim of the same task fails (no longer queued).
	assert.False(t, s.ClaimAndDequeue(1))

	// Unknown task ID fails without side effects.
	assert.False(t, s.ClaimAndDequeue(999))
	assert.Equal(t, 1, s.QueueLen())
}

func TestMemoryTaskStateStore_DequeueByID(t *testing.T) {
	s := NewMemoryTaskStateStore()
	s.Enqueue(newTestRunner(1, 10, 100))
	s.Enqueue(newTestRunner(2, 10, 100))
	s.Enqueue(newTestRunner(3, 10, 100))

	// Remove the middle element by value; the others keep their order.
	s.DequeueByID(2)
	require.Equal(t, 2, s.QueueLen())
	assert.Equal(t, 1, s.QueueGet(0).Task.ID)
	assert.Equal(t, 3, s.QueueGet(1).Task.ID)

	// Removing the running task (not queued) is a no-op.
	s.DequeueByID(99)
	assert.Equal(t, 2, s.QueueLen())
}

// TestMemoryTaskStateStore_ClaimAndDequeue_ExactlyOnce ensures concurrent
// claimers never run the same task twice and never lose a task — the contract
// that the Redis store must also honor across nodes.
func TestMemoryTaskStateStore_ClaimAndDequeue_ExactlyOnce(t *testing.T) {
	s := NewMemoryTaskStateStore()
	const n = 200
	for i := 1; i <= n; i++ {
		s.Enqueue(newTestRunner(i, 10, 100))
	}

	var mu sync.Mutex
	claimed := map[int]int{} // taskID -> times claimed

	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := 1; id <= n; id++ {
				if s.ClaimAndDequeue(id) {
					mu.Lock()
					claimed[id]++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, s.QueueLen(), "every task must be dequeued")
	assert.Len(t, claimed, n, "every task must be claimed")
	for id, count := range claimed {
		assert.Equal(t, 1, count, "task %d must be claimed exactly once", id)
	}
}

// TestMemoryTaskStateStore_TryFinalize_ExactlyOnce ensures concurrent
// finalizers never both succeed — the contract Redis-backed stores must honor
// across HA nodes.
func TestMemoryTaskStateStore_TryFinalize_ExactlyOnce(t *testing.T) {
	s := NewMemoryTaskStateStore()
	const n = 200

	var mu sync.Mutex
	finalized := map[int]int{}

	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := 1; id <= n; id++ {
				if s.TryFinalize(id) {
					mu.Lock()
					finalized[id]++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	for id := 1; id <= n; id++ {
		s.DeleteFinalize(id)
	}

	assert.Len(t, finalized, n, "every task must be finalized exactly once")
	for id, count := range finalized {
		assert.Equal(t, 1, count, "task %d must be finalized exactly once", id)
	}
}

// MemoryTaskStateStore must satisfy the TaskStateInspector contract.
var _ TaskStateInspector = (*MemoryTaskStateStore)(nil)
