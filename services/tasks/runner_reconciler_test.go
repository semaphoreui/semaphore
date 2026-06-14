package tasks

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reconcilerStoreStub wraps a real store and forces errors on selected methods.
type reconcilerStoreStub struct {
	db.Store
	globalRunnerErr error
	updateTaskErr   error
}

func (s *reconcilerStoreStub) GetGlobalRunner(runnerID int) (db.Runner, error) {
	if s.globalRunnerErr != nil {
		return db.Runner{}, s.globalRunnerErr
	}
	return s.Store.GetGlobalRunner(runnerID)
}

func (s *reconcilerStoreStub) UpdateTask(task db.Task) error {
	if s.updateTaskErr != nil {
		return s.updateTaskErr
	}
	return s.Store.UpdateTask(task)
}

func newReconcilerTestPool(store db.Store, state TaskStateStore) TaskPool {
	return TaskPool{
		queueEvents:     make(chan PoolEvent, 10),
		logger:          make(chan logRecord, 100),
		state:           state,
		store:           store,
		logWriteService: &mockLogWriteService{},
		stop:            make(chan struct{}),
		reconcileDone:   make(chan struct{}),
	}
}

// dispatchPendingQueueEvents drains buffered queueEvents through the same logic
// as handleQueue. Used by reconciler unit tests that do not start the full pool.
func dispatchPendingQueueEvents(pool *TaskPool) {
	for {
		select {
		case ev := <-pool.queueEvents:
			pool.dispatchQueueEvent(ev)
		default:
			return
		}
	}
}

func setupReconcilerConfig(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{}
}

// createReconcilerTestTask persists the FK chain a task row requires
// (project -> key -> repo -> inventory -> template) plus a runner, and
// returns the created task and the runner ID.
func createReconcilerTestTask(
	t *testing.T,
	store *sql.SqlDb,
	status task_logger.TaskStatus,
	start *time.Time,
) (db.Task, int) {
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
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

	runner, err := store.CreateRunner(db.Runner{Name: "test-runner"})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		ProjectID:  proj.ID,
		TemplateID: tpl.ID,
		Status:     status,
		Start:      start,
		RunnerID:   &runner.ID,
	}, 0)
	require.NoError(t, err)

	return task, runner.ID
}

// createReconcilerTestTaskNoRunner persists the FK chain and a task row with no
// runner assigned (runner_id NULL) — the state of a task whose dispatch never
// reached a runner.
func createReconcilerTestTaskNoRunner(
	t *testing.T,
	store *sql.SqlDb,
	status task_logger.TaskStatus,
) db.Task {
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
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
		Status:     status,
	}, 0)
	require.NoError(t, err)

	return task
}

func TestDecideRunnerTaskAction(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	offlineTimeout := 2 * time.Minute
	taskFailTimeout := 7 * time.Minute

	ago := func(d time.Duration) *time.Time {
		v := now.Add(-d)
		return &v
	}

	tests := []struct {
		name      string
		status    task_logger.TaskStatus
		taskStart *time.Time
		runner    *db.Runner
		expected  RunnerTaskAction
	}{
		{
			"alive runner, starting task",
			task_logger.TaskStartingStatus, nil,
			&db.Runner{Touched: ago(10 * time.Second)},
			RunnerTaskKeep,
		},
		{
			"alive runner, running task",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Touched: ago(10 * time.Second), StartedAt: ago(2 * time.Hour)},
			RunnerTaskKeep,
		},
		{
			"starting task, runner offline",
			task_logger.TaskStartingStatus, nil,
			&db.Runner{Touched: ago(3 * time.Minute)},
			RunnerTaskRequeue,
		},
		{
			"waiting task with runner, runner offline",
			task_logger.TaskWaitingStatus, nil,
			&db.Runner{Touched: ago(3 * time.Minute)},
			RunnerTaskRequeue,
		},
		{
			"running task, silence within recovery window",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Touched: ago(5 * time.Minute)},
			RunnerTaskKeep,
		},
		{
			"running task, silence past recovery window",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Touched: ago(8 * time.Minute)},
			RunnerTaskFail,
		},
		{
			"running task, runner restarted after task start",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Touched: ago(5 * time.Second), StartedAt: ago(10 * time.Minute)},
			RunnerTaskFail,
		},
		{
			"running task, restart within skew margin",
			task_logger.TaskRunningStatus, ago(time.Minute),
			&db.Runner{Touched: ago(5 * time.Second), StartedAt: ago(50 * time.Second)},
			RunnerTaskFail,
		},
		{
			"starting task, runner restarted (self-heals via NewJobs)",
			task_logger.TaskStartingStatus, nil,
			&db.Runner{Touched: ago(5 * time.Second), StartedAt: ago(time.Minute)},
			RunnerTaskKeep,
		},
		{
			"runner started before task",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Touched: ago(time.Minute), StartedAt: ago(2 * time.Hour)},
			RunnerTaskKeep,
		},
		{
			"runner deleted, starting task",
			task_logger.TaskStartingStatus, nil,
			nil,
			RunnerTaskRequeue,
		},
		{
			"runner deleted, running task",
			task_logger.TaskRunningStatus, ago(time.Hour),
			nil,
			RunnerTaskFail,
		},
		{
			"finished task",
			task_logger.TaskSuccessStatus, ago(time.Hour),
			nil,
			RunnerTaskKeep,
		},
		{
			"stopping task is out of scope",
			task_logger.TaskStoppingStatus, ago(time.Hour),
			&db.Runner{Touched: ago(time.Hour)},
			RunnerTaskKeep,
		},
		{
			"webhook runner, starting task, stale heartbeat",
			task_logger.TaskStartingStatus, nil,
			&db.Runner{Webhook: "https://example.com/hook", Touched: ago(time.Hour)},
			RunnerTaskKeep,
		},
		{
			"webhook runner, running task, silence past recovery window",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{Webhook: "https://example.com/hook", Touched: ago(8 * time.Minute)},
			RunnerTaskFail,
		},
		{
			"poll runner never polled, starting task",
			task_logger.TaskStartingStatus, nil,
			&db.Runner{},
			RunnerTaskRequeue,
		},
		{
			"poll runner never polled, running task",
			task_logger.TaskRunningStatus, ago(time.Hour),
			&db.Runner{},
			RunnerTaskKeep,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, reason := DecideRunnerTaskAction(
				tt.status, tt.taskStart, tt.runner, now, offlineTimeout, taskFailTimeout)
			assert.Equal(t, tt.expected, action)
			if action != RunnerTaskKeep {
				assert.NotEmpty(t, reason)
			}
		})
	}
}

func TestSelectRunner(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	offlineTimeout := 2 * time.Minute

	ago := func(d time.Duration) *time.Time {
		v := now.Add(-d)
		return &v
	}

	online := db.Runner{ID: 1, Touched: ago(10 * time.Second), MaxParallelTasks: 0}
	offline := db.Runner{ID: 2, Touched: ago(time.Hour), MaxParallelTasks: 0}
	webhook := db.Runner{ID: 3, Webhook: "https://example.com/hook", MaxParallelTasks: 0}
	busyOnline := db.Runner{ID: 4, Touched: ago(10 * time.Second), MaxParallelTasks: 1}

	noBusy := func(int) int { return 0 }

	t.Run("offline runner excluded", func(t *testing.T) {
		r := selectRunner([]db.Runner{offline, online}, now, offlineTimeout, noBusy)
		require.NotNil(t, r)
		assert.Equal(t, online.ID, r.ID)
	})

	t.Run("all offline returns nil", func(t *testing.T) {
		assert.Nil(t, selectRunner([]db.Runner{offline}, now, offlineTimeout, noBusy))
	})

	t.Run("webhook runner selectable regardless of heartbeat", func(t *testing.T) {
		r := selectRunner([]db.Runner{offline, webhook}, now, offlineTimeout, noBusy)
		require.NotNil(t, r)
		assert.Equal(t, webhook.ID, r.ID)
	})

	t.Run("online but at capacity skipped", func(t *testing.T) {
		busy := func(runnerID int) int {
			if runnerID == busyOnline.ID {
				return 1
			}
			return 0
		}
		r := selectRunner([]db.Runner{busyOnline, online}, now, offlineTimeout, busy)
		require.NotNil(t, r)
		assert.Equal(t, online.ID, r.ID)
	})

	t.Run("previously offline runner selectable after fresh poll", func(t *testing.T) {
		revived := offline
		revived.Touched = ago(5 * time.Second)
		r := selectRunner([]db.Runner{revived}, now, offlineTimeout, noBusy)
		require.NotNil(t, r)
		assert.Equal(t, revived.ID, r.ID)
	})

	t.Run("no runners returns nil", func(t *testing.T) {
		assert.Nil(t, selectRunner(nil, now, offlineTimeout, noBusy))
	})
}

func TestRequeueTaskRunnerOffline(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()

	pool := TaskPool{
		queueEvents:     make(chan PoolEvent, 10),
		logger:          make(chan logRecord, 100),
		state:           state,
		store:           store,
		logWriteService: &mockLogWriteService{},
	}

	newTask, runnerID := createReconcilerTestTask(t, store, task_logger.TaskStartingStatus, nil)

	tsk := &TaskRunner{
		Task: newTask,
		pool: &pool,
	}
	state.SetRunning(tsk)

	pool.requeueTaskRunnerOffline(tsk, runnerID, "runner is offline")

	// RunnerID cleared, status reset to waiting; enqueue happens when the
	// requeue event is processed.
	assert.Nil(t, tsk.Task.RunnerID)
	assert.Equal(t, task_logger.TaskWaitingStatus, tsk.Task.Status)
	assert.Equal(t, 0, state.QueueLen())

	// EventTypeRequeued emitted so the pool releases running/active state.
	select {
	case ev := <-pool.queueEvents:
		assert.Equal(t, EventTypeRequeued, ev.eventType)
		assert.Equal(t, tsk.Task.ID, ev.task.Task.ID)
		pool.dispatchQueueEvent(ev)
	default:
		t.Fatal("expected EventTypeRequeued in queueEvents")
	}

	assert.Equal(t, 1, state.QueueLen())

	// Cleared RunnerID persisted: the old runner can no longer pull the task.
	row, err := store.GetTaskByID(newTask.ID)
	require.NoError(t, err)
	assert.Nil(t, row.RunnerID)
	assert.Equal(t, task_logger.TaskWaitingStatus, row.Status)
}

func TestRequeueTaskRunnerOffline_NoopWhenAlreadyRunning(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()

	pool := TaskPool{
		queueEvents:     make(chan PoolEvent, 10),
		logger:          make(chan logRecord, 100),
		state:           state,
		store:           store,
		logWriteService: &mockLogWriteService{},
	}

	runnerID := 42
	tsk := &TaskRunner{
		Task: db.Task{
			ID:       1,
			Status:   task_logger.TaskRunningStatus,
			RunnerID: &runnerID,
		},
		pool: &pool,
	}

	pool.requeueTaskRunnerOffline(tsk, runnerID, "runner is offline")

	// The task started executing concurrently: it must not be reassigned.
	assert.Equal(t, task_logger.TaskRunningStatus, tsk.Task.Status)
	assert.NotNil(t, tsk.Task.RunnerID)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestFinalizeRemoteTask_DoesNotOverwriteConcurrentTerminalSuccess(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()

	pool := TaskPool{
		queueEvents:     make(chan PoolEvent, 10),
		logger:          make(chan logRecord, 100),
		state:           state,
		store:           store,
		logWriteService: &mockLogWriteService{},
	}

	now := time.Now()
	newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)

	// Simulate a runner report on another node that has already persisted a
	// terminal success and won the finalization lock, while a stale finalizer
	// snapshot still says the task failed.
	require.True(t, state.TryFinalize(newTask.ID))
	defer state.DeleteFinalize(newTask.ID)
	completedTask := newTask
	completedTask.Status = task_logger.TaskSuccessStatus
	require.NoError(t, store.UpdateTask(completedTask))

	staleFinalizerTask := &TaskRunner{
		Task: newTask,
		pool: &pool,
	}
	staleFinalizerTask.Task.Status = task_logger.TaskFailStatus
	state.SetRunning(staleFinalizerTask)

	pool.FinalizeRemoteTask(staleFinalizerTask, nil)

	assert.Equal(t, task_logger.TaskFailStatus, staleFinalizerTask.Task.Status)
	assert.Nil(t, staleFinalizerTask.Task.End)
	assert.Empty(t, pool.queueEvents)

	row, err := store.GetTaskByID(newTask.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskSuccessStatus, row.Status)
	assert.Nil(t, row.End)
}

func TestFailTaskRunnerLost(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()

	pool := TaskPool{
		queueEvents:     make(chan PoolEvent, 10),
		logger:          make(chan logRecord, 100),
		state:           state,
		store:           store,
		logWriteService: &mockLogWriteService{},
	}

	now := time.Now()
	newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)

	tsk := &TaskRunner{
		Task: newTask,
		pool: &pool,
	}
	state.SetRunning(tsk)

	pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")

	assert.Equal(t, task_logger.TaskFailStatus, tsk.Task.Status)
	assert.NotNil(t, tsk.Task.End)

	// EventTypeFinished emitted by finalization (finishRun -> onTaskStop).
	select {
	case ev := <-pool.queueEvents:
		assert.Equal(t, EventTypeFinished, ev.eventType)
	default:
		t.Fatal("expected EventTypeFinished in queueEvents")
	}

	row, err := store.GetTaskByID(newTask.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, row.Status)
	assert.NotNil(t, row.End)

	// Second call is a no-op: the task is already finished.
	pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")
	assert.Empty(t, pool.queueEvents)
}

func TestReconcileRunnerTasks(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := newSpyTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	now := time.Now()

	// Kept: running task on an existing runner that never polled.
	keepTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)
	keepTsk := &TaskRunner{Task: keepTask, pool: &pool}
	state.SetRunning(keepTsk)

	// Kept: no runner assigned but a live goroutine is dispatching it right now.
	dispatchingTask := createReconcilerTestTaskNoRunner(t, store, task_logger.TaskStartingStatus)
	dispatchingTsk := &TaskRunner{Task: dispatchingTask, pool: &pool}
	dispatchingTsk.dispatching.Store(true)
	state.SetRunning(dispatchingTsk)

	// Requeued: starting task with no runner and no live dispatch goroutine — a
	// stub left in the running set after the node restarted mid-dispatch.
	stuckTask := createReconcilerTestTaskNoRunner(t, store, task_logger.TaskStartingStatus)
	stuckTsk := &TaskRunner{Task: stuckTask, pool: &pool}
	state.SetRunning(stuckTsk)

	// Skipped: already finished.
	finishedRunnerID := 1
	finishedTsk := &TaskRunner{
		Task: db.Task{ID: 900002, Status: task_logger.TaskSuccessStatus, RunnerID: &finishedRunnerID},
		pool: &pool,
	}
	state.SetRunning(finishedTsk)

	// Requeued: starting task whose runner row was deleted. The in-memory
	// copy keeps the stale RunnerID while the DB nulls it (FK on delete
	// set null) — exactly the production state after a runner removal.
	requeueTask, requeueRunnerID := createReconcilerTestTask(t, store, task_logger.TaskStartingStatus, nil)
	require.NoError(t, store.DeleteGlobalRunner(requeueRunnerID))
	requeueTsk := &TaskRunner{Task: requeueTask, pool: &pool}
	state.SetRunning(requeueTsk)

	// Failed: running task whose runner row was deleted.
	failTask, failRunnerID := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)
	require.NoError(t, store.DeleteGlobalRunner(failRunnerID))
	failTsk := &TaskRunner{Task: failTask, pool: &pool}
	state.SetRunning(failTsk)

	pool.reconcileRunnerTasks(time.Now())

	assert.Equal(t, task_logger.TaskRunningStatus, keepTsk.Task.Status)
	assert.NotNil(t, keepTsk.Task.RunnerID)

	// Actively dispatching: left alone.
	assert.Equal(t, task_logger.TaskStartingStatus, dispatchingTsk.Task.Status)

	assert.Equal(t, task_logger.TaskSuccessStatus, finishedTsk.Task.Status)

	// Stuck stub: returned to the queue.
	assert.Equal(t, task_logger.TaskWaitingStatus, stuckTsk.Task.Status)

	assert.Equal(t, task_logger.TaskWaitingStatus, requeueTsk.Task.Status)
	assert.Nil(t, requeueTsk.Task.RunnerID)

	var pending []PoolEvent
	for len(pool.queueEvents) > 0 {
		pending = append(pending, <-pool.queueEvents)
	}
	var events []EventType
	for _, ev := range pending {
		events = append(events, ev.eventType)
		pool.dispatchQueueEvent(ev)
	}
	assert.ElementsMatch(t,
		[]EventType{EventTypeRequeued, EventTypeRequeued, EventTypeFinished}, events)
	// Two tasks requeued (the deleted-runner task and the stuck stub).
	assert.Equal(t, 2, state.QueueLen())

	assert.Equal(t, task_logger.TaskFailStatus, failTsk.Task.Status)
	assert.NotNil(t, failTsk.Task.End)
}

func TestReconcileRunnerTasks_StoreErrorSkipsTask(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(
		&reconcilerStoreStub{Store: store, globalRunnerErr: errors.New("connection lost")},
		state,
	)

	runnerID := 42
	tsk := &TaskRunner{
		Task: db.Task{ID: 1, Status: task_logger.TaskStartingStatus, RunnerID: &runnerID},
		pool: &pool,
	}
	state.SetRunning(tsk)

	pool.reconcileRunnerTasks(time.Now())

	// The runner state is unknown: the task must be left alone.
	assert.Equal(t, task_logger.TaskStartingStatus, tsk.Task.Status)
	assert.NotNil(t, tsk.Task.RunnerID)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestRunnerTasksReconcileLoop(t *testing.T) {
	setupReconcilerConfig(t)

	// CreateTestStore replaces util.Config, so the interval is set after it.
	store := sql.CreateTestStore()
	util.Config.Runners = &util.RunnersConfig{ReconcileIntervalSec: 1}
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	// A starting task on a never-polled runner is requeued on the first tick.
	newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskStartingStatus, nil)
	tsk := &TaskRunner{Task: newTask, pool: &pool}
	state.SetRunning(tsk)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pool.handleQueue()
	}()

	done := make(chan struct{})
	go func() {
		pool.runnerTasksReconcileLoop()
		close(done)
	}()

	assert.Eventually(t, func() bool {
		return state.QueueLen() == 1
	}, 5*time.Second, 50*time.Millisecond)

	close(pool.stop)
	<-done
	close(pool.queueEvents)
	wg.Wait()
}

func TestRequeueTaskRunnerOffline_FinalizeLockHeld(t *testing.T) {
	setupReconcilerConfig(t)

	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(sql.CreateTestStore(), state)

	runnerID := 42
	tsk := &TaskRunner{
		Task: db.Task{ID: 7, Status: task_logger.TaskStartingStatus, RunnerID: &runnerID},
		pool: &pool,
	}

	// Another actor holds the finalize lock: requeue must bail out untouched.
	require.True(t, state.TryFinalize(tsk.Task.ID))
	defer state.DeleteFinalize(tsk.Task.ID)

	pool.requeueTaskRunnerOffline(tsk, runnerID, "runner is offline")

	assert.Equal(t, task_logger.TaskStartingStatus, tsk.Task.Status)
	assert.NotNil(t, tsk.Task.RunnerID)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestRequeueTaskRunnerOffline_NoopWhenReassigned(t *testing.T) {
	tests := []struct {
		name     string
		runnerID *int
	}{
		{"task moved to another runner", intPtr(43)},
		{"runner already cleared", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupReconcilerConfig(t)

			state := NewMemoryTaskStateStore()
			pool := newReconcilerTestPool(sql.CreateTestStore(), state)

			tsk := &TaskRunner{
				Task: db.Task{ID: 8, Status: task_logger.TaskStartingStatus, RunnerID: tt.runnerID},
				pool: &pool,
			}

			pool.requeueTaskRunnerOffline(tsk, 42, "runner is offline")

			assert.Equal(t, task_logger.TaskStartingStatus, tsk.Task.Status)
			assert.Equal(t, tt.runnerID, tsk.Task.RunnerID)
			assert.Equal(t, 0, state.QueueLen())
			assert.Empty(t, pool.queueEvents)
		})
	}
}

func TestRequeueTaskRunnerOffline_PersistError(t *testing.T) {
	setupReconcilerConfig(t)

	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(
		&reconcilerStoreStub{Store: sql.CreateTestStore(), updateTaskErr: errors.New("db down")},
		state,
	)

	runnerID := 42
	// Waiting status makes SetStatus a no-op, so the only UpdateTask call is
	// the explicit persist of the cleared RunnerID.
	tsk := &TaskRunner{
		Task: db.Task{ID: 9, Status: task_logger.TaskWaitingStatus, RunnerID: &runnerID},
		pool: &pool,
	}

	pool.requeueTaskRunnerOffline(tsk, runnerID, "runner is offline")

	// Persist failed: the task must not be enqueued (the old runner could
	// still pull it), and no requeue event must be emitted.
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestRequeueTaskRunnerOffline_HA(t *testing.T) {
	setupReconcilerConfig(t)

	// CreateTestStore replaces util.Config, so HA is enabled after it.
	store := sql.CreateTestStore()
	util.Config.HA = &util.HAConfig{Enabled: true}
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	newTask, runnerID := createReconcilerTestTask(t, store, task_logger.TaskStartingStatus, nil)

	// Stale in-memory copy: in HA mode the DB row is authoritative.
	staleTask := newTask
	staleTask.Status = task_logger.TaskRunningStatus
	tsk := &TaskRunner{Task: staleTask, pool: &pool}
	state.SetRunning(tsk)

	pool.requeueTaskRunnerOffline(tsk, runnerID, "runner is offline")

	// The DB refresh restored "starting", so the requeue proceeded.
	assert.Nil(t, tsk.Task.RunnerID)
	assert.Equal(t, task_logger.TaskWaitingStatus, tsk.Task.Status)
	dispatchPendingQueueEvents(&pool)
	assert.Equal(t, 1, state.QueueLen())
}

func TestFailTaskRunnerLost_HA(t *testing.T) {
	t.Run("DB row still running: task failed", func(t *testing.T) {
		setupReconcilerConfig(t)

		// CreateTestStore replaces util.Config, so HA is enabled after it.
		store := sql.CreateTestStore()
		util.Config.HA = &util.HAConfig{Enabled: true}
		state := NewMemoryTaskStateStore()
		pool := newReconcilerTestPool(store, state)

		now := time.Now()
		newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)

		tsk := &TaskRunner{Task: newTask, pool: &pool}
		state.SetRunning(tsk)

		pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")

		assert.Equal(t, task_logger.TaskFailStatus, tsk.Task.Status)
		assert.NotNil(t, tsk.Task.End)
	})

	t.Run("DB row already finished: no-op", func(t *testing.T) {
		setupReconcilerConfig(t)

		// CreateTestStore replaces util.Config, so HA is enabled after it.
		store := sql.CreateTestStore()
		util.Config.HA = &util.HAConfig{Enabled: true}
		state := NewMemoryTaskStateStore()
		pool := newReconcilerTestPool(store, state)

		now := time.Now()
		newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)

		// Another node already finished the task in the DB.
		finished := newTask
		finished.Status = task_logger.TaskSuccessStatus
		finished.End = &now
		require.NoError(t, store.UpdateTask(finished))

		// Stale in-memory copy still says "running".
		tsk := &TaskRunner{Task: newTask, pool: &pool}
		state.SetRunning(tsk)

		pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")

		// The DB refresh observed the terminal status: nothing was failed.
		assert.Equal(t, task_logger.TaskSuccessStatus, tsk.Task.Status)
		assert.Empty(t, pool.queueEvents)
	})
}

func TestRequeueUndispatchedTask(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	newTask := createReconcilerTestTaskNoRunner(t, store, task_logger.TaskStartingStatus)
	tsk := &TaskRunner{Task: newTask, pool: &pool}
	state.SetRunning(tsk)

	pool.requeueUndispatchedTask(tsk)

	// Reset to waiting; enqueue happens when the requeue event is processed.
	assert.Equal(t, task_logger.TaskWaitingStatus, tsk.Task.Status)
	assert.Nil(t, tsk.Task.RunnerID)
	assert.Equal(t, 0, state.QueueLen())

	select {
	case ev := <-pool.queueEvents:
		assert.Equal(t, EventTypeRequeued, ev.eventType)
		assert.Equal(t, tsk.Task.ID, ev.task.Task.ID)
		pool.dispatchQueueEvent(ev)
	default:
		t.Fatal("expected EventTypeRequeued in queueEvents")
	}

	assert.Equal(t, 1, state.QueueLen())

	row, err := store.GetTaskByID(newTask.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskWaitingStatus, row.Status)
}

func TestReconcileRunnerTasks_DispatchingTaskKept(t *testing.T) {
	setupReconcilerConfig(t)

	store := sql.CreateTestStore()
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	// A task with no runner that this process is actively dispatching must not
	// be requeued out from under its live goroutine.
	newTask := createReconcilerTestTaskNoRunner(t, store, task_logger.TaskStartingStatus)
	tsk := &TaskRunner{Task: newTask, pool: &pool}
	tsk.dispatching.Store(true)
	state.SetRunning(tsk)

	pool.reconcileRunnerTasks(time.Now())

	assert.Equal(t, task_logger.TaskStartingStatus, tsk.Task.Status)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestRequeueUndispatchedTask_NoopWhenRunnerAssignedConcurrently(t *testing.T) {
	setupReconcilerConfig(t)

	// CreateTestStore replaces util.Config, so HA is enabled after it.
	store := sql.CreateTestStore()
	util.Config.HA = &util.HAConfig{Enabled: true}
	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(store, state)

	// A concurrent dispatch assigned a runner and persisted it in the DB while
	// our in-memory stub still shows "starting" with no runner.
	persisted, runnerID := createReconcilerTestTask(t, store, task_logger.TaskStartingStatus, nil)

	stale := persisted
	stale.RunnerID = nil
	tsk := &TaskRunner{Task: stale, pool: &pool}
	state.SetRunning(tsk)

	pool.requeueUndispatchedTask(tsk)

	// The DB refresh observed the assigned runner: the requeue bailed out.
	assert.NotNil(t, tsk.Task.RunnerID)
	assert.Equal(t, runnerID, *tsk.Task.RunnerID)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func TestRequeueUndispatchedTask_FinalizeLockHeld(t *testing.T) {
	setupReconcilerConfig(t)

	state := NewMemoryTaskStateStore()
	pool := newReconcilerTestPool(sql.CreateTestStore(), state)

	tsk := &TaskRunner{
		Task: db.Task{ID: 11, Status: task_logger.TaskStartingStatus},
		pool: &pool,
	}

	// Another actor holds the finalize lock: requeue must bail out untouched.
	require.True(t, state.TryFinalize(tsk.Task.ID))
	defer state.DeleteFinalize(tsk.Task.ID)

	pool.requeueUndispatchedTask(tsk)

	assert.Equal(t, task_logger.TaskStartingStatus, tsk.Task.Status)
	assert.Equal(t, 0, state.QueueLen())
	assert.Empty(t, pool.queueEvents)
}

func intPtr(v int) *int { return &v }
