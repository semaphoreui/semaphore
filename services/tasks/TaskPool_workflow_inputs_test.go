package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainLogs reads the log records queued for the pool's log writer.
func drainLogs(p *TaskPool) (lines []string) {
	for {
		select {
		case rec := <-p.logger:
			lines = append(lines, rec.output)
		default:
			return
		}
	}
}

func TestTaskPool_AddFailedTask(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	runID, nodeID := 4, 9
	user, err := fixture.store.CreateUserWithoutPassword(db.User{Username: "denis", Email: "denis@example.com", Name: "Denis"})
	require.NoError(t, err)
	userID := user.ID

	task, err := fixture.pool.AddFailedTask(db.Task{
		ProjectID:      fixture.task.ProjectID,
		TemplateID:     fixture.template.ID,
		WorkflowRunID:  &runID,
		WorkflowNodeID: &nodeID,
		Environment:    `{"image_tag":""}`,
		Secret:         `{"token":"must not be stored"}`,
	}, &userID, "denis", []string{"Input \"image_tag\" is required but has no value", "Required input(s) without a value: image_tag"})
	require.NoError(t, err)

	assert.Equal(t, task_logger.TaskFailStatus, task.Status)
	require.NotNil(t, task.Start)
	require.NotNil(t, task.End)
	assert.Empty(t, task.Secret)
	assert.Equal(t, &userID, task.UserID)

	persisted, err := fixture.store.GetTaskByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, persisted.Status)
	assert.NotNil(t, persisted.End)
	require.NotNil(t, persisted.WorkflowRunID)
	assert.Equal(t, runID, *persisted.WorkflowRunID)
	assert.Equal(t, `{"image_tag":""}`, persisted.Environment)

	lines := drainLogs(&fixture.pool)
	assert.Equal(t, []string{"Input \"image_tag\" is required but has no value", "Required input(s) without a value: image_tag"}, lines)

	// Nothing was enqueued: the engine progresses the run from the stored status.
	assert.Empty(t, fixture.pool.state.QueueRange())
}

func TestTaskPool_AddFailedTask_UnknownTemplate(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)

	_, err := fixture.pool.AddFailedTask(db.Task{ProjectID: fixture.task.ProjectID, TemplateID: 99999}, nil, "", nil)

	assert.Error(t, err)
}

func TestTaskPool_LogTask(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)

	t.Run("task registered in the pool", func(t *testing.T) {
		tr := NewTaskRunner(fixture.task, &fixture.pool, "", nil)
		fixture.pool.StateStore().SetRunning(tr)
		t.Cleanup(func() { fixture.pool.StateStore().DeleteRunning(tr.Task.ID) })

		fixture.pool.LogTask(fixture.task, []string{"first", "second"})

		assert.Equal(t, []string{"first", "second"}, drainLogs(&fixture.pool))
	})

	t.Run("task not registered yet still gets its lines", func(t *testing.T) {
		fixture.pool.LogTask(fixture.task, []string{"early"})

		assert.Equal(t, []string{"early"}, drainLogs(&fixture.pool))
	})

	t.Run("no lines, no records", func(t *testing.T) {
		fixture.pool.LogTask(fixture.task, nil)

		assert.Empty(t, drainLogs(&fixture.pool))
	})
}
