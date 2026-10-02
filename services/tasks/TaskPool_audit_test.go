package tasks

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCreateMetadata(t *testing.T) {
	id := func(v int) *int { return &v }
	tests := []struct {
		name    string
		trigger string
		task    db.Task
		want    audit.TaskCreateMetadata
	}{
		{"api", audit.TriggerAPI, db.Task{TemplateID: 3, UserID: id(7)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAPI, TemplateID: 3}},
		{"api ignores source IDs from the request", audit.TriggerAPI, db.Task{TemplateID: 3, UserID: id(7), ScheduleID: id(5), IntegrationID: id(6), WorkflowRunID: id(4)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAPI, TemplateID: 3}},
		{"deploy started by a user", audit.TriggerAPI, db.Task{TemplateID: 3, UserID: id(7), BuildTaskID: id(9), ScheduleID: id(5)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAPI, TemplateID: 3, ParentTaskID: 9}},
		{"schedule", audit.TriggerSchedule, db.Task{TemplateID: 3, ScheduleID: id(5)}, audit.TaskCreateMetadata{Trigger: audit.TriggerSchedule, TemplateID: 3, ScheduleID: 5}},
		{"integration", audit.TriggerIntegration, db.Task{TemplateID: 3, IntegrationID: id(6)}, audit.TaskCreateMetadata{Trigger: audit.TriggerIntegration, TemplateID: 3, IntegrationID: 6}},
		{"autorun", audit.TriggerAutorun, db.Task{TemplateID: 3, BuildTaskID: id(9)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAutorun, TemplateID: 3, ParentTaskID: 9}},
		{"workflow", audit.TriggerWorkflow, db.Task{TemplateID: 3, WorkflowRunID: id(4), BuildTaskID: id(9)}, audit.TaskCreateMetadata{Trigger: audit.TriggerWorkflow, TemplateID: 3, WorkflowRunID: 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, taskCreateMetadata(tt.trigger, tt.task))
		})
	}
}

func TestAddTaskFrom_RecordsTheCallerActor(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	fixture.pool.register = make(chan *TaskRunner, 1)
	ctx := audit.WithActor(context.Background(), audit.SystemActor(audit.ComponentScheduler))
	schedule, err := fixture.store.CreateSchedule(db.Schedule{ProjectID: fixture.template.ProjectID, TemplateID: fixture.template.ID, CronFormat: "* * * * *"})
	require.NoError(t, err)

	task, err := fixture.pool.AddTaskFrom(ctx, audit.TriggerSchedule, db.Task{TemplateID: fixture.template.ID, ScheduleID: &schedule.ID}, nil, "", fixture.template.ProjectID, false)
	require.NoError(t, err)

	got, err := rec.Only(audit.TaskExecutionCreate)
	require.NoError(t, err)
	assert.Equal(t, audit.SystemActor(audit.ComponentScheduler), got.Actor)
	assert.Equal(t, &audit.Target{Type: audit.TargetTask, ID: strconv.Itoa(task.ID), Name: fixture.template.Name}, got.Event.Target)
	assert.Equal(t, audit.TriggerSchedule, got.Event.Metadata.(audit.TaskCreateMetadata).Trigger)
}

func TestAddTask_ActsForTheWorkflowRunUser(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	fixture.pool.register = make(chan *TaskRunner, 1)
	user, err := fixture.store.CreateUserWithoutPassword(db.User{Username: "alice", Name: "Alice", Email: "alice@example.com"})
	require.NoError(t, err)

	_, err = fixture.pool.AddTask(db.Task{TemplateID: fixture.template.ID}, &user.ID, "alice", fixture.template.ProjectID, false)
	require.NoError(t, err)

	got, err := rec.Only(audit.TaskExecutionCreate)
	require.NoError(t, err)
	assert.Equal(t, audit.UserActor(user.ID, "alice", "", ""), got.Actor)
	assert.Equal(t, audit.TriggerWorkflow, got.Event.Metadata.(audit.TaskCreateMetadata).Trigger)
}

func TestTaskPool_WithoutRecorderRecordsNothing(t *testing.T) {
	pool := TaskPool{}
	assert.Equal(t, audit.Nop{}, pool.recorder())
}

func TestFinishRun_RecordsCompletionOnce(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	userID := 7
	fixture.task.UserID = &userID
	taskRunner := TaskRunner{Task: fixture.task, Template: fixture.template, pool: &fixture.pool, keyInstaller: fixture.keyInstaller}
	taskRunner.job = &successfulKilledJob{onRun: func() {
		taskRunner.SetStatus(task_logger.TaskStoppingStatus)
	}}

	taskRunner.run()

	got, err := rec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err)
	assert.Equal(t, audit.SystemActor(audit.ComponentTaskRunner), got.Actor)
	meta := got.Event.Metadata.(audit.TaskCompleteMetadata)
	assert.Equal(t, "stopped", meta.Result)
	assert.Equal(t, 7, meta.InitiatorID)
	assert.Equal(t, fixture.template.ID, meta.TemplateID)
}

func TestFailTaskRunnerLost_RecordsTheReconciler(t *testing.T) {
	setupReconcilerConfig(t)
	store := sql.InitConfigCreateTestStore()
	pool := newReconcilerTestPool(store, NewMemoryTaskStateStore())
	rec := &audittest.Recorder{}
	pool.SetAuditRecorder(rec)
	now := time.Now()
	newTask, _ := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)
	tsk := &TaskRunner{Task: newTask, pool: &pool}
	pool.state.SetRunning(tsk)

	pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")
	pool.failTaskRunnerLost(tsk, nil, "runner stopped responding")

	got, err := rec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err, "a second finalize records nothing")
	assert.Equal(t, audit.SystemActor(audit.ComponentReconciler), got.Actor)
	meta := got.Event.Metadata.(audit.TaskCompleteMetadata)
	assert.Equal(t, "error", meta.Result)
	assert.Equal(t, audit.EndReasonRunnerLost, meta.EndReason)
	assert.Equal(t, audit.ReasonNone, got.Event.Reason)
}

func TestFinalizeRemoteTask_RecordsTheReportingRunner(t *testing.T) {
	setupReconcilerConfig(t)
	store := sql.InitConfigCreateTestStore()
	pool := newReconcilerTestPool(store, NewMemoryTaskStateStore())
	rec := &audittest.Recorder{}
	pool.SetAuditRecorder(rec)
	now := time.Now()
	newTask, runnerID := createReconcilerTestTask(t, store, task_logger.TaskSuccessStatus, &now)
	tsk := &TaskRunner{Task: newTask, pool: &pool}

	pool.FinalizeRemoteTask(tsk, &db.Runner{ID: runnerID, Name: "r1"})

	got, err := rec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err)
	assert.Equal(t, audit.RunnerActor(runnerID, "r1"), got.Actor)
}

// failingSecretService makes populateDetails fail after the task row is stored.
type failingSecretService struct{ EncryptionServiceMock }

func (*failingSecretService) DeserializeSecret(*db.AccessKey) error { return errors.New("no key") }

func TestAddTaskFrom_FailedPreparationCompletesTheTask(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	fixture.pool.register = make(chan *TaskRunner, 1)
	fixture.pool.encryptionService = &failingSecretService{}

	_, err := fixture.pool.AddTaskFrom(context.Background(), audit.TriggerAPI, db.Task{TemplateID: fixture.template.ID}, nil, "", fixture.template.ProjectID, false)
	require.Error(t, err)

	kinds := []audit.Kind{}
	for _, got := range rec.All() {
		kinds = append(kinds, got.Event.Kind)
	}
	assert.Equal(t, []audit.Kind{audit.TaskExecutionCreate, audit.TaskExecutionComplete}, kinds)
	assert.Equal(t, "error", rec.All()[1].Event.Metadata.(audit.TaskCompleteMetadata).Result)
}

func TestSetStatus_WaitingConfirmationRequestsApproval(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	runner, err := fixture.store.CreateRunner(db.Runner{Name: "r"})
	require.NoError(t, err)
	fixture.task.RunnerID = &runner.ID
	taskRunner := TaskRunner{Task: fixture.task, Template: fixture.template, pool: &fixture.pool}

	taskRunner.SetStatus(task_logger.TaskWaitingConfirmation)
	taskRunner.SetStatus(task_logger.TaskWaitingConfirmation)

	got, err := rec.Only(audit.TaskApprovalRequest)
	require.NoError(t, err, "an unchanged status requests nothing")
	assert.Equal(t, audit.RunnerActor(runner.ID, ""), got.Actor)
}

func TestFinalizeRemoteTask_DispatchFailureIsNotTheRunner(t *testing.T) {
	setupReconcilerConfig(t)
	store := sql.InitConfigCreateTestStore()
	pool := newReconcilerTestPool(store, NewMemoryTaskStateStore())
	rec := &audittest.Recorder{}
	pool.SetAuditRecorder(rec)
	now := time.Now()
	newTask, runnerID := createReconcilerTestTask(t, store, task_logger.TaskRunningStatus, &now)
	tsk := &TaskRunner{Task: newTask, pool: &pool}

	tsk.FailDispatch()
	pool.FinalizeRemoteTask(tsk, &db.Runner{ID: runnerID, Name: "r1"})

	got, err := rec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err)
	assert.Equal(t, audit.SystemActor(audit.ComponentTaskRunner), got.Actor)
	assert.Empty(t, got.Event.Metadata.(audit.TaskCompleteMetadata).EndReason)
}

func TestStopTasksByTemplate_CompletesOnlyTasksOutsideTheQueue(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	rec := &audittest.Recorder{}
	fixture.pool.SetAuditRecorder(rec)
	starting, err := fixture.store.CreateTask(db.Task{ProjectID: fixture.template.ProjectID, TemplateID: fixture.template.ID, Status: task_logger.TaskStartingStatus}, 0)
	require.NoError(t, err)
	waiting, err := fixture.store.CreateTask(db.Task{ProjectID: fixture.template.ProjectID, TemplateID: fixture.template.ID, Status: task_logger.TaskWaitingStatus}, 0)
	require.NoError(t, err)
	fixture.pool.state.Enqueue(NewTaskRunner(waiting, &fixture.pool, "", nil))

	fixture.pool.StopTasksByTemplate(fixture.template.ProjectID, fixture.template.ID, false)

	got, err := rec.Only(audit.TaskExecutionComplete)
	require.NoError(t, err, "a queued task gets no complete")
	assert.Equal(t, strconv.Itoa(starting.ID), got.Event.Target.ID)
	assert.Equal(t, "stopped", got.Event.Metadata.(audit.TaskCompleteMetadata).Result)
}

func TestRecordComplete_EndReasonSetByTheTimeoutTimer(t *testing.T) {
	pool := TaskPool{}
	pool.SetAuditRecorder(&audittest.Recorder{})
	tr := &TaskRunner{pool: &pool}
	done := make(chan struct{})
	go func() {
		tr.endReason.Store(audit.EndReasonTimeout)
		close(done)
	}()
	tr.recordComplete(audit.SystemActor(audit.ComponentTaskRunner))
	<-done
}
