package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBTimeout = errors.New("dial tcp: i/o timeout")

// timeoutStore wraps a real store and fails selected calls the way a database
// that stopped answering does.
type timeoutStore struct {
	db.Store
	failGetProject bool
	failUpdateTask bool
}

func (s *timeoutStore) GetProject(projectID int) (db.Project, error) {
	if s.failGetProject {
		return db.Project{}, errDBTimeout
	}
	return s.Store.GetProject(projectID)
}

func (s *timeoutStore) UpdateTask(task db.Task) error {
	if s.failUpdateTask {
		return errDBTimeout
	}
	return s.Store.UpdateTask(task)
}

// A database timeout while a new task is being prepared used to panic after
// the row was stored as "waiting": the task was never queued and never
// marked failed, so it stayed "waiting" forever.
func TestAddTaskFrom_DBTimeoutDuringPreparationFailsTheTask(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	fixture.pool.register = make(chan *TaskRunner, 1)
	fixture.pool.store = &timeoutStore{Store: fixture.store, failGetProject: true}

	var newTask db.Task
	var err error
	require.NotPanics(t, func() {
		newTask, err = fixture.pool.AddTaskFrom(context.Background(), audit.TriggerAPI,
			db.Task{TemplateID: fixture.template.ID}, nil, "", fixture.template.ProjectID, false)
	})
	require.ErrorIs(t, err, errDBTimeout)
	require.NotZero(t, newTask.ID)

	stored, err := fixture.store.GetTask(fixture.template.ProjectID, newTask.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskFailStatus, stored.Status)
	assert.NotNil(t, stored.End)
	assert.Empty(t, fixture.pool.register, "a failed task is not queued")
}

// When the database also rejects the status update, the request must still
// return an error instead of panicking.
func TestAddTaskFrom_DBTimeoutOnStatusUpdateReturnsAnError(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	fixture.pool.register = make(chan *TaskRunner, 1)
	fixture.pool.store = &timeoutStore{Store: fixture.store, failGetProject: true, failUpdateTask: true}

	var err error
	require.NotPanics(t, func() {
		_, err = fixture.pool.AddTaskFrom(context.Background(), audit.TriggerAPI,
			db.Task{TemplateID: fixture.template.ID}, nil, "", fixture.template.ProjectID, false)
	})
	require.ErrorIs(t, err, errDBTimeout)
	assert.Empty(t, fixture.pool.register)
}

// A status write that times out must not take the pool goroutine down.
func TestSaveStatus_DBTimeoutDoesNotPanic(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	fixture.pool.store = &timeoutStore{Store: fixture.store, failUpdateTask: true}
	taskRunner := TaskRunner{Task: fixture.task, Template: fixture.template, pool: &fixture.pool}

	assert.NotPanics(t, taskRunner.saveStatus)
}
