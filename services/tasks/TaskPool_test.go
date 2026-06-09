package tasks

import (
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ClaimAndDequeue is the path the queue loop actually uses. Returning false
// keeps tests from starting real tasks while still counting claim attempts.
func (s *spyTaskStateStore) ClaimAndDequeue(_ int) bool {
	s.tryClaimCalls++
	return false
}

func TestTaskPool_RequeuedEventCleansRunningStateAndSkipsImmediateRetry(t *testing.T) {
	// Ensure util.Config is non-nil for p.blocks() checks.
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{MaxParallelTasks: 0}

	store := sql.CreateTestStore()
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

func TestTaskPool_FinalizeRemoteTask_ReleasesPoolStateWhenEndAlreadyPersisted(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.CreateTestStore()
	util.Config.HA = &util.HAConfig{Enabled: true}

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

	inv, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
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
		Status:     task_logger.TaskSuccessStatus,
	}, 0)
	require.NoError(t, err)

	end := tz.Now()
	task.End = &end
	require.NoError(t, store.UpdateTask(task))

	state := NewMemoryTaskStateStore()
	pool := TaskPool{
		state:       state,
		store:       store,
		queueEvents: make(chan PoolEvent, 1),
	}

	tr := &TaskRunner{
		Task:     task,
		Template: tpl,
		pool:     &pool,
		Alias:    "remote-alias",
	}
	state.SetRunning(tr)
	state.AddActive(tr.Task.ProjectID, tr)
	state.SetAlias(tr.Alias, tr)

	pool.FinalizeRemoteTask(tr, nil)

	assert.Equal(t, 0, state.RunningCount(), "pool running set must be released when End is already in DB")
	assert.Equal(t, 0, state.ActiveCount(tr.Task.ProjectID), "active-by-project must be released when End is already in DB")
	assert.Nil(t, state.GetByAlias(tr.Alias), "alias mapping must be cleared")
}
