package tasks

import (
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Start the WebSocket hub so that sockets.Message() doesn't block.
	go sockets.StartWS()
}

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

	store := sql.InitConfigCreateTestStore()
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

func TestTaskPool_FinalizeRemoteTask_HA_ReleasesStalePoolStateWhenEndPersisted(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()
	util.Config.MaxParallelTasks = 0
	util.Config.HA = &util.HAConfig{Enabled: true}
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Name:      "test-key",
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "test-repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:    proj.ID,
		Name:         "remote tpl",
		Playbook:     "pb.yml",
		RepositoryID: repo.ID,
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

	persisted, err := store.GetTaskByID(task.ID)
	require.NoError(t, err)
	require.NotNil(t, persisted.End, "test setup must persist task end timestamp in DB")

	state := NewMemoryTaskStateStore()
	pool := TaskPool{
		queueEvents: make(chan PoolEvent, 1),
		state:       state,
		store:       store,
	}

	tr := &TaskRunner{
		Task: db.Task{
			ID:         task.ID,
			ProjectID:  task.ProjectID,
			TemplateID: task.TemplateID,
			Status:     task_logger.TaskSuccessStatus,
			// End intentionally unset in memory: DB already has it from another node.
		},
		Template: tpl,
		Alias:    "stale-alias",
		job:      &RemoteJob{Task: task},
	}
	state.SetRunning(tr)
	state.AddActive(tr.Task.ProjectID, tr)
	state.SetAlias(tr.Alias, tr)

	pool.FinalizeRemoteTask(tr, nil)

	assert.Equal(t, 0, state.RunningCount(), "stale running entry must be cleared when End is already persisted")
	assert.Equal(t, 0, state.ActiveCount(tr.Task.ProjectID), "stale active entry must be cleared when End is already persisted")
	assert.Nil(t, state.GetByAlias(tr.Alias), "stale alias mapping must be cleared when End is already persisted")

	select {
	case ev := <-pool.queueEvents:
		t.Fatalf("expected no finish event when End already set, got %+v", ev)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestTaskPool_StopTasksByTemplate_DequeuesWaitingTasksByID(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{MaxParallelTasks: 0}

	store := sql.InitConfigCreateTestStore()
	proj, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	state := NewMemoryTaskStateStore()
	pool := TaskPool{
		state: state,
		store: store,
	}

	stopMe := &TaskRunner{
		Task: db.Task{ID: 1, ProjectID: proj.ID, TemplateID: 10, Status: task_logger.TaskWaitingStatus},
	}
	keepMe := &TaskRunner{
		Task: db.Task{ID: 2, ProjectID: proj.ID, TemplateID: 20, Status: task_logger.TaskWaitingStatus},
	}
	state.Enqueue(stopMe)
	state.Enqueue(keepMe)

	pool.StopTasksByTemplate(proj.ID, 10, true)

	assert.Equal(t, 1, state.QueueLen(), "only the targeted template's waiting task should be dequeued")
	assert.Equal(t, keepMe.Task.ID, state.QueueGet(0).Task.ID)
}

// stubLogWriteService is a no-op log write service for tests.
type stubLogWriteService struct{}

func (s *stubLogWriteService) WriteEventLog(_ pro_interfaces.EventLogRecord) error { return nil }
func (s *stubLogWriteService) WriteTaskLog(_ pro_interfaces.TaskLogRecord) error   { return nil }
func (s *stubLogWriteService) WriteResult(_ any) error                             { return nil }

// stubEncryptionService is a no-op encryption service for tests.
type stubEncryptionService struct{}

func (s *stubEncryptionService) SerializeSecret(_ *db.AccessKey) error { return nil }

func (s *stubEncryptionService) DeserializeSecret(_ *db.AccessKey) error { return nil }

func (s *stubEncryptionService) FillEnvironmentSecrets(_ *db.Environment, _ bool) error { return nil }

func (s *stubEncryptionService) DeleteSecret(_ *db.AccessKey) error { return nil }

func (s *stubEncryptionService) RekeyAccessKeys(_ string) error { return nil }

func (s *stubEncryptionService) CreateTaskSurveySecrets(_ int, _ int, _ string, _ time.Time) error {
	return nil
}

func (s *stubEncryptionService) GetTaskSurveySecrets(_ int, _ int) (string, error) { return "", nil }
func (s *stubEncryptionService) DeleteTaskSurveySecrets(_ int, _ int) error        { return nil }

// stubInventoryService returns a fixed inventory for tests.
type stubInventoryService struct {
	inv db.Inventory
	err error
}

func (s *stubInventoryService) GetInventory(_ int, _ int) (db.Inventory, error) {
	return s.inv, s.err
}

// createOrphanTestFixtures sets up minimal DB state for orphan recovery tests:
// a project, user, access key, repository, inventory, template, and a task in
// the given status. Returns the created task.
func createOrphanTestFixtures(t *testing.T, store *sql.SqlDb, status task_logger.TaskStatus) db.Task {
	t.Helper()

	proj, err := store.CreateProject(db.Project{Name: "orphan-test"})
	assert.NoError(t, err)

	user, err := store.CreateUser(db.UserWithPwd{
		User: db.User{Username: "testuser", Name: "Test", Email: "test@example.com"},
		Pwd:  "password123",
	})
	assert.NoError(t, err)

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj.ID,
		UserID:    user.ID,
		Role:      db.ProjectOwner,
	})
	assert.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "none-key",
		Type:      db.AccessKeyNone,
		ProjectID: &proj.ID,
	})
	assert.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		Name:      "test-inv",
		ProjectID: proj.ID,
		Type:      "static",
		Inventory: "localhost",
	})
	assert.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		Name:      "test-repo",
		ProjectID: proj.ID,
		GitURL:    "https://git.example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	assert.NoError(t, err)

	invID := inv.ID
	tpl, err := store.CreateTemplate(db.Template{
		Name:         "test-template",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &invID,
		Playbook:     "test.yml",
		App:          db.AppAnsible,
	})
	assert.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		ProjectID:  proj.ID,
		TemplateID: tpl.ID,
		Status:     status,
	}, 0)
	assert.NoError(t, err)

	return task
}

func TestRecoverOrphanedTasks_NoOrphans(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()

	pool := TaskPool{
		store:           store,
		state:           NewMemoryTaskStateStore(),
		logger:          make(chan logRecord, 10000),
		logWriteService: &stubLogWriteService{},
	}

	// Should be a no-op when there are no orphaned tasks.
	pool.RecoverOrphanedTasks()
}

func TestRecoverOrphanedTasks_LocalTaskMarkedFailed(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	store := sql.InitConfigCreateTestStore()
	task := createOrphanTestFixtures(t, store, task_logger.TaskRunningStatus)

	pool := TaskPool{
		store:             store,
		state:             NewMemoryTaskStateStore(),
		logger:            make(chan logRecord, 10000),
		queueEvents:       make(chan PoolEvent, 1),
		encryptionService: &stubEncryptionService{},
		inventoryService:  &stubInventoryService{},
		logWriteService:   &stubLogWriteService{},
	}

	pool.RecoverOrphanedTasks()

	// Verify the task was marked as failed in the database.
	updated, err := store.GetTask(task.ProjectID, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, updated.Status)
	assert.NotNil(t, updated.End, "task End time should be set")

	// The task went through the normal completion path.
	require.Len(t, pool.queueEvents, 1)
	ev := <-pool.queueEvents
	assert.Equal(t, EventTypeFinished, ev.eventType)
	assert.Equal(t, task.ID, ev.task.Task.ID)
}

func TestRecoverOrphanedTasks_UnassignedTaskFailedWithRemoteRunners(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	store := sql.InitConfigCreateTestStore()
	// Remote runners are configured, but no runner was ever assigned, so no
	// runner can report the task.
	util.Config.UseRemoteRunner = true
	task := createOrphanTestFixtures(t, store, task_logger.TaskStartingStatus)

	pool := TaskPool{
		store:             store,
		state:             NewMemoryTaskStateStore(),
		logger:            make(chan logRecord, 10000),
		queueEvents:       make(chan PoolEvent, 1),
		encryptionService: &stubEncryptionService{},
		inventoryService:  &stubInventoryService{},
		logWriteService:   &stubLogWriteService{},
	}

	pool.RecoverOrphanedTasks()

	updated, err := store.GetTask(task.ProjectID, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, updated.Status)
	assert.NotNil(t, updated.End)
	assert.Equal(t, 0, pool.state.RunningCount())
}

func TestRecoverOrphanedTasks_AssignedTaskReadopted(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	store := sql.InitConfigCreateTestStore()
	task := createOrphanTestFixtures(t, store, task_logger.TaskRunningStatus)

	runner, err := store.CreateRunner(db.Runner{Name: "orphan-runner"})
	require.NoError(t, err)
	task.RunnerID = &runner.ID
	require.NoError(t, store.UpdateTask(task))

	pool := TaskPool{
		store:             store,
		state:             NewMemoryTaskStateStore(),
		logger:            make(chan logRecord, 10000),
		encryptionService: &stubEncryptionService{},
		inventoryService:  &stubInventoryService{},
		logWriteService:   &stubLogWriteService{},
	}

	// Runner settings no longer select a remote runner, but the task was
	// assigned to one before the restart.
	pool.RecoverOrphanedTasks()

	updated, err := store.GetTask(task.ProjectID, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task_logger.TaskRunningStatus, updated.Status)
	assert.Nil(t, updated.End)

	running := pool.state.RunningRange()
	require.Len(t, running, 1)
	assert.Equal(t, task.ID, running[0].Task.ID)
	require.NotNil(t, running[0].Task.RunnerID)
	assert.Equal(t, runner.ID, *running[0].Task.RunnerID)
}

func TestRecoverOrphanedTasks_SkippedInHAMode(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	store := sql.InitConfigCreateTestStore()
	// In HA the unfinished task may belong to another live node.
	util.Config.HA = &util.HAConfig{Enabled: true}
	task := createOrphanTestFixtures(t, store, task_logger.TaskRunningStatus)

	pool := TaskPool{
		store:             store,
		state:             NewMemoryTaskStateStore(),
		logger:            make(chan logRecord, 10000),
		encryptionService: &stubEncryptionService{},
		inventoryService:  &stubInventoryService{},
		logWriteService:   &stubLogWriteService{},
	}

	pool.RecoverOrphanedTasks()

	updated, err := store.GetTask(task.ProjectID, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task_logger.TaskRunningStatus, updated.Status)
	assert.Nil(t, updated.End)
}

func TestRecoverOrphanedTasks_HydrationFailureStillFails(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	store := sql.InitConfigCreateTestStore()

	// Report the inventory as deleted so hydration fails.
	task := createOrphanTestFixtures(t, store, task_logger.TaskRunningStatus)

	pool := TaskPool{
		store:             store,
		state:             NewMemoryTaskStateStore(),
		logger:            make(chan logRecord, 10000),
		encryptionService: &stubEncryptionService{},
		inventoryService:  &stubInventoryService{err: db.ErrNotFound},
	}

	pool.RecoverOrphanedTasks()

	// Even though hydration failed, the task should be marked as failed.
	updated, err := store.GetTask(task.ProjectID, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, updated.Status)
	assert.NotNil(t, updated.End, "task End time should be set")
}
