package tasks

import (
	"context"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCreateMetadata(t *testing.T) {
	id := func(v int) *int { return &v }
	tests := []struct {
		name string
		task db.Task
		want audit.TaskCreateMetadata
	}{
		{"api", db.Task{TemplateID: 3, UserID: id(7)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAPI, TemplateID: 3}},
		{"schedule", db.Task{TemplateID: 3, ScheduleID: id(5)}, audit.TaskCreateMetadata{Trigger: audit.TriggerSchedule, TemplateID: 3, ScheduleID: 5}},
		{"integration", db.Task{TemplateID: 3, IntegrationID: id(6)}, audit.TaskCreateMetadata{Trigger: audit.TriggerIntegration, TemplateID: 3, IntegrationID: 6}},
		{"autorun", db.Task{TemplateID: 3, BuildTaskID: id(9)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAutorun, TemplateID: 3, ParentTaskID: 9}},
		{"deploy started by a user", db.Task{TemplateID: 3, UserID: id(7), BuildTaskID: id(9)}, audit.TaskCreateMetadata{Trigger: audit.TriggerAPI, TemplateID: 3, ParentTaskID: 9}},
		{"workflow", db.Task{TemplateID: 3, WorkflowRunID: id(4), BuildTaskID: id(9)}, audit.TaskCreateMetadata{Trigger: audit.TriggerWorkflow, TemplateID: 3, WorkflowRunID: 4, ParentTaskID: 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, taskCreateMetadata(tt.task))
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

	task, err := fixture.pool.AddTaskFrom(ctx, db.Task{TemplateID: fixture.template.ID, ScheduleID: &schedule.ID}, nil, "", fixture.template.ProjectID, false)
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
}

func TestTaskPool_WithoutRecorderRecordsNothing(t *testing.T) {
	pool := TaskPool{}
	assert.Equal(t, audit.Nop{}, pool.recorder())
}
