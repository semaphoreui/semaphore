package tasks

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// RunnerID cleared, status reset to waiting, task back in the queue.
	assert.Nil(t, tsk.Task.RunnerID)
	assert.Equal(t, task_logger.TaskWaitingStatus, tsk.Task.Status)
	assert.Equal(t, 1, state.QueueLen())

	// EventTypeRequeued emitted so the pool releases running/active state.
	select {
	case ev := <-pool.queueEvents:
		assert.Equal(t, EventTypeRequeued, ev.eventType)
		assert.Equal(t, tsk.Task.ID, ev.task.Task.ID)
	default:
		t.Fatal("expected EventTypeRequeued in queueEvents")
	}

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
