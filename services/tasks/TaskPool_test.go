package tasks

import (
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingWorkflowService struct {
	mu        sync.Mutex
	completed []db.Task
}

func (s *recordingWorkflowService) StartWorkflow(workflow db.WorkflowTemplate, user *db.User) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (s *recordingWorkflowService) ProgressWorkflowRun(projectID int, runID int, user *db.User) error {
	return nil
}

func (s *recordingWorkflowService) StopWorkflowRun(projectID int, runID int, user *db.User) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (s *recordingWorkflowService) ResolveWorkflowApproval(projectID int, workflowID int, runID int, nodeID int, status db.WorkflowApprovalStatus, user *db.User) (db.WorkflowApproval, error) {
	return db.WorkflowApproval{}, nil
}

func (s *recordingWorkflowService) HandleWorkflowTaskCompletion(task db.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.completed = append(s.completed, task)
	return nil
}

func (s *recordingWorkflowService) GetWorkflowRunArtifacts(projectID int, runID int, currentTaskID *int) (map[string]any, error) {
	return map[string]any{}, nil
}

var _ pro_interfaces.WorkflowService = (*recordingWorkflowService)(nil)

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

func TestTaskPool_StopTasksByTemplate_NotifiesWorkflowOnWaitingTasks(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{MaxParallelTasks: 0}

	store := sql.CreateTestStore()
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

	tpl, err := store.CreateTemplate(db.Template{
		Name:         "workflow tpl",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
	})
	require.NoError(t, err)

	runID := 42
	task, err := store.CreateTask(db.Task{
		ProjectID:     proj.ID,
		TemplateID:    tpl.ID,
		Status:        task_logger.TaskWaitingStatus,
		WorkflowRunID: &runID,
	}, 0)
	require.NoError(t, err)

	state := NewMemoryTaskStateStore()
	workflowSvc := &recordingWorkflowService{}
	pool := TaskPool{
		state:            state,
		store:            store,
		workflowService:  workflowSvc,
	}

	pool.StopTasksByTemplate(proj.ID, tpl.ID, true)

	stopped, err := store.GetTask(proj.ID, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskStoppedStatus, stopped.Status)
	require.NotNil(t, stopped.End)

	workflowSvc.mu.Lock()
	defer workflowSvc.mu.Unlock()
	require.Len(t, workflowSvc.completed, 1)
	assert.Equal(t, task.ID, workflowSvc.completed[0].ID)
	assert.Equal(t, runID, *workflowSvc.completed[0].WorkflowRunID)
	assert.Equal(t, task_logger.TaskStoppedStatus, workflowSvc.completed[0].Status)
}
