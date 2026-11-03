package schedules

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEncryptionService is a test implementation of AccessKeyEncryptionService
type mockEncryptionService struct{}

func (m *mockEncryptionService) RekeyAccessKeys(oldKey string) (err error) {
	return nil
}

func (m *mockEncryptionService) SerializeSecret(key *db.AccessKey) error {
	return nil
}

func (m *mockEncryptionService) DeserializeSecret(key *db.AccessKey) error {
	return nil
}

func (m *mockEncryptionService) FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error {
	return nil
}

func (m *mockEncryptionService) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (m *mockEncryptionService) CreateTaskSurveySecrets(projectID int, taskID int, secrets string, expireAt time.Time) error {
	return nil
}

func (m *mockEncryptionService) GetTaskSurveySecrets(projectID int, taskID int) (string, error) {
	return "", nil
}

func (m *mockEncryptionService) DeleteTaskSurveySecrets(projectID int, taskID int) error {
	return nil
}

func TestValidateCronFormat(t *testing.T) {
	err := ValidateCronFormat("* * * *")
	if err == nil {
		t.Fatal("")
	}

	err = ValidateCronFormat("* * 1 * *")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestOneTimeSchedule(t *testing.T) {
	future := time.Now().Add(time.Hour)
	schedule := oneTimeSchedule{runAt: future}

	if schedule.Next(time.Now()) != future {
		t.Fatalf("expected next run at %v", future)
	}

	if !schedule.Next(future).IsZero() {
		t.Fatalf("expected schedule to stop after run time")
	}
}

// mockDeduplicator is a test implementation of ScheduleDeduplicator
type mockDeduplicator struct {
	mu             sync.Mutex
	allowExecution map[int]bool
	lockAttempts   map[int]int
}

func newMockDeduplicator() *mockDeduplicator {
	return &mockDeduplicator{
		allowExecution: make(map[int]bool),
		lockAttempts:   make(map[int]int),
	}
}

func (m *mockDeduplicator) TryLockExecution(scheduleID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lockAttempts[scheduleID]++
	return m.allowExecution[scheduleID]
}

func (m *mockDeduplicator) setAllowExecution(scheduleID int, allow bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowExecution[scheduleID] = allow
}

func (m *mockDeduplicator) getLockAttempts(scheduleID int) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lockAttempts[scheduleID]
}

// mockAccessKeyInstaller is a test implementation of AccessKeyInstaller
type mockAccessKeyInstaller struct{}

func (m *mockAccessKeyInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error) {
	return ssh.AccessKeyInstallation{}, nil
}

func setupTestSchedulePool(t *testing.T) (*SchedulePool, db.Store) {
	store := sql.InitConfigCreateTestStore()

	// Store original config and restore after test
	originalSchedule := util.Config.Schedule
	t.Cleanup(func() {
		util.Config.Schedule = originalSchedule
	})

	// Ensure util.Config Schedule is set (CreateTestStore doesn't set this)
	util.Config.Schedule = &util.ScheduleConfig{
		Timezone: "UTC",
	}

	pool := CreateSchedulePool(
		store,
		&tasks.TaskPool{},
		&mockAccessKeyInstaller{},
		&mockEncryptionService{},
	)

	t.Cleanup(func() {
		pool.Destroy()
	})

	return &pool, store
}

// TestSetDeduplicator verifies that SetDeduplicator properly configures the deduplicator
func TestSetDeduplicator(t *testing.T) {
	pool, _ := setupTestSchedulePool(t)

	// Initially no deduplicator should be set
	assert.Nil(t, pool.dedup, "deduplicator should be nil initially")

	// Set a deduplicator
	dedup := newMockDeduplicator()
	pool.SetDeduplicator(dedup)

	// Verify it's set
	assert.NotNil(t, pool.dedup, "deduplicator should be set after calling SetDeduplicator")
	assert.Equal(t, dedup, pool.dedup, "deduplicator should be the one we set")
}

// TestScheduleExecutesNormallyWithoutDeduplicator verifies schedules execute when no deduplicator is set
func TestScheduleExecutesNormallyWithoutDeduplicator(t *testing.T) {
	pool, _ := setupTestSchedulePool(t)

	// Ensure no deduplicator is set
	pool.SetDeduplicator(nil)

	// Verify that the deduplicator is nil (schedule would execute normally)
	assert.Nil(t, pool.dedup, "deduplicator should be nil, allowing normal execution")
}

// TestScheduleSkippedWhenTryLockExecutionReturnsFalse verifies schedules are skipped when TryLockExecution returns false
func TestScheduleSkippedWhenTryLockExecutionReturnsFalse(t *testing.T) {
	pool, _ := setupTestSchedulePool(t)

	// Set up deduplicator to deny execution
	dedup := newMockDeduplicator()
	scheduleID := 123
	dedup.setAllowExecution(scheduleID, false)
	pool.SetDeduplicator(dedup)

	// Simulate the deduplication check that happens in ScheduleRunner.Run()
	shouldSkip := pool.dedup != nil && !pool.dedup.TryLockExecution(scheduleID)

	// Verify the deduplicator was called and returned false
	assert.True(t, shouldSkip, "schedule should be skipped when TryLockExecution returns false")
	assert.Equal(t, 1, dedup.getLockAttempts(scheduleID), "TryLockExecution should be called once")
}

// TestScheduleProceedsWhenTryLockExecutionReturnsTrue verifies schedules proceed when TryLockExecution returns true
func TestScheduleProceedsWhenTryLockExecutionReturnsTrue(t *testing.T) {
	pool, _ := setupTestSchedulePool(t)

	// Set up deduplicator to allow execution
	dedup := newMockDeduplicator()
	scheduleID := 456
	dedup.setAllowExecution(scheduleID, true)
	pool.SetDeduplicator(dedup)

	// Simulate the deduplication check that happens in ScheduleRunner.Run()
	shouldSkip := pool.dedup != nil && !pool.dedup.TryLockExecution(scheduleID)

	// Verify the deduplicator was called and returned true (schedule proceeds)
	assert.False(t, shouldSkip, "schedule should proceed when TryLockExecution returns true")
	assert.Equal(t, 1, dedup.getLockAttempts(scheduleID), "TryLockExecution should be called once")
}

// fakeWorkflowService records the runs the scheduler asks for.
type fakeWorkflowService struct {
	pro_interfaces.WorkflowService
	mu      sync.Mutex
	started []db.WorkflowRunParams
	actors  []audit.Actor
	err     error
}

func (f *fakeWorkflowService) StartWorkflow(ctx context.Context, workflow db.WorkflowTemplate, params db.WorkflowRunParams, user *db.User) (db.WorkflowRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, params)
	f.actors = append(f.actors, audit.ActorFrom(ctx))
	if f.err != nil {
		return db.WorkflowRun{}, f.err
	}
	return db.WorkflowRun{ID: 1, ProjectID: workflow.ProjectID, WorkflowTemplateID: workflow.ID}, nil
}

// fakeWorkflowRepo answers GetWorkflowTemplate with a stored template: the
// open-source workflow store returns an empty one, which the runner treats
// as "workflows unavailable".
type fakeWorkflowRepo struct {
	db.WorkflowManager
	templates map[int]db.WorkflowTemplate
}

func (f fakeWorkflowRepo) GetWorkflowTemplate(projectID int, workflowID int) (db.WorkflowTemplate, error) {
	tpl, ok := f.templates[workflowID]
	if !ok || tpl.ProjectID != projectID {
		return db.WorkflowTemplate{}, db.ErrNotFound
	}
	return tpl, nil
}

// createWorkflowSchedule inserts a workflow template row directly (the
// workflow store is a Pro feature) and a schedule that starts it.
func createWorkflowSchedule(t *testing.T, store db.Store) (db.Schedule, db.WorkflowTemplate) {
	t.Helper()
	sqlStore := store.(*sql.SqlDb)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	_, err = sqlStore.Sql().Exec(sqlStore.PrepareQuery(
		"insert into project__workflow_template (project_id, name) values (?, ?)"), project.ID, "wf")
	require.NoError(t, err)
	workflowID, err := sqlStore.Sql().SelectInt(sqlStore.PrepareQuery(
		"select id from project__workflow_template where project_id = ? and name = ?"), project.ID, "wf")
	require.NoError(t, err)
	id := int(workflowID)
	schedule, err := store.CreateSchedule(db.Schedule{
		Name: "nightly", ProjectID: project.ID, WorkflowID: &id, CronFormat: "0 0 * * *", Active: true,
	})
	require.NoError(t, err)
	return schedule, db.WorkflowTemplate{ID: id, ProjectID: project.ID, Name: "wf"}
}

func TestScheduleRunner_StartsWorkflow(t *testing.T) {
	pool, store := setupTestSchedulePool(t)
	schedule, workflow := createWorkflowSchedule(t, store)
	service := &fakeWorkflowService{}
	pool.SetWorkflowService(fakeWorkflowRepo{templates: map[int]db.WorkflowTemplate{workflow.ID: workflow}}, service)
	dedup := newMockDeduplicator()
	dedup.setAllowExecution(schedule.ID, true)
	pool.SetDeduplicator(dedup)

	CreateScheduleRunner(schedule.ProjectID, schedule.ID, pool, pool.encryptionService, pool.keyInstaller).Run()

	require.Len(t, service.started, 1, "one run is started")
	require.NotNil(t, service.started[0].ScheduleID)
	assert.Equal(t, schedule.ID, *service.started[0].ScheduleID)
	assert.Nil(t, service.started[0].IntegrationID)
	assert.Empty(t, service.started[0].Environment, "a schedule starts the workflow with the survey defaults")
	assert.Equal(t, audit.SystemActor(audit.ComponentScheduler), service.actors[0])
	assert.Equal(t, 1, dedup.getLockAttempts(schedule.ID), "the HA lock is taken before the start")
}

func TestScheduleRunner_WorkflowStartFailureIsLogged(t *testing.T) {
	pool, store := setupTestSchedulePool(t)
	schedule, workflow := createWorkflowSchedule(t, store)
	service := &fakeWorkflowService{err: errors.New("required survey variable has no value")}
	pool.SetWorkflowService(fakeWorkflowRepo{templates: map[int]db.WorkflowTemplate{workflow.ID: workflow}}, service)

	assert.NotPanics(t, func() {
		CreateScheduleRunner(schedule.ProjectID, schedule.ID, pool, pool.encryptionService, pool.keyInstaller).Run()
	})
	assert.Len(t, service.started, 1)
}

func TestScheduleRunner_WorkflowSkippedWithoutEngine(t *testing.T) {
	pool, store := setupTestSchedulePool(t)
	schedule, workflow := createWorkflowSchedule(t, store)

	t.Run("no service wired", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CreateScheduleRunner(schedule.ProjectID, schedule.ID, pool, pool.encryptionService, pool.keyInstaller).Run()
		})
	})

	t.Run("open-source store returns an empty template", func(t *testing.T) {
		service := &fakeWorkflowService{}
		pool.SetWorkflowService(fakeWorkflowRepo{templates: map[int]db.WorkflowTemplate{}}, service)
		CreateScheduleRunner(schedule.ProjectID, schedule.ID, pool, pool.encryptionService, pool.keyInstaller).Run()
		assert.Empty(t, service.started, "nothing is started without a workflow")
		_ = workflow
	})

	t.Run("skipped when another node holds the lock", func(t *testing.T) {
		service := &fakeWorkflowService{}
		pool.SetWorkflowService(fakeWorkflowRepo{templates: map[int]db.WorkflowTemplate{workflow.ID: workflow}}, service)
		dedup := newMockDeduplicator()
		dedup.setAllowExecution(schedule.ID, false)
		pool.SetDeduplicator(dedup)
		CreateScheduleRunner(schedule.ProjectID, schedule.ID, pool, pool.encryptionService, pool.keyInstaller).Run()
		assert.Empty(t, service.started)
	})
}
