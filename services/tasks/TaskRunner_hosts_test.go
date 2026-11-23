package tasks

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hostsJob struct {
	run   func() error
	hosts *string
	async bool
}

func (j *hostsJob) Run(string, *string, string) error { return j.run() }
func (j *hostsJob) Kill()                             {}
func (j *hostsJob) IsKilled() bool                    { return false }
func (j *hostsJob) Async() bool                       { return j.async }
func (j *hostsJob) Hosts() *string                    { return j.hosts }

func TestTaskRunner_SavesHostsOfFailedTask(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	repo := &ansibleTaskRepoFake{}
	fixture.pool.ansibleTaskRepo = repo
	taskRunner := TaskRunner{
		Task:         fixture.task,
		Template:     fixture.template,
		pool:         &fixture.pool,
		keyInstaller: fixture.keyInstaller,
	}
	document := `{"version":1,"hosts":{"web-01":{"ok":2,"failed":1}}}`
	taskRunner.job = &hostsJob{
		run: func() error {
			taskRunner.SetStatus(task_logger.TaskRunningStatus)
			return errors.New("playbook failed")
		},
		hosts: &document,
	}

	taskRunner.run()

	assert.Equal(t, task_logger.TaskFailStatus, taskRunner.Task.Status)
	assert.Equal(t, 1, repo.calls)
	assert.Equal(t, fixture.task.ID, repo.taskID)
	require.Len(t, repo.hosts, 1)
	assert.Equal(t, "web-01", repo.hosts[0].Host)
}

func TestTaskRunner_InvalidHostsDoNotChangeStatus(t *testing.T) {
	fixture := newTaskRunnerRunFixture(t)
	repo := &ansibleTaskRepoFake{}
	fixture.pool.ansibleTaskRepo = repo
	taskRunner := TaskRunner{
		Task:         fixture.task,
		Template:     fixture.template,
		pool:         &fixture.pool,
		keyInstaller: fixture.keyInstaller,
	}
	document := `{`
	taskRunner.job = &hostsJob{
		run: func() error {
			taskRunner.SetStatus(task_logger.TaskRunningStatus)
			return nil
		},
		hosts: &document,
	}
	var output strings.Builder
	taskRunner.AddLogListener(func(_ time.Time, msg string) {
		output.WriteString(msg)
	})

	taskRunner.run()

	assert.Equal(t, task_logger.TaskSuccessStatus, taskRunner.Task.Status)
	assert.Equal(t, 0, repo.calls)
	assert.Contains(t, output.String(), "Host stats were not saved")
}

func TestTaskRunnerRun_RecordsInventoryBeforeJobRuns(t *testing.T) {
	tests := []struct {
		name  string
		async bool
	}{
		{"local job", false},
		{"remote job", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newTaskRunnerRunFixture(t)
			inventoryID := *fixture.template.InventoryID
			taskRunner := TaskRunner{
				Task:         fixture.task,
				Template:     fixture.template,
				Inventory:    db.Inventory{ID: inventoryID},
				pool:         &fixture.pool,
				keyInstaller: fixture.keyInstaller,
			}
			var recordedAtRun *int
			taskRunner.job = &hostsJob{
				run: func() error {
					persisted, err := fixture.store.GetTaskByID(fixture.task.ID)
					require.NoError(t, err)
					recordedAtRun = persisted.EffectiveInventoryID
					taskRunner.SetStatus(task_logger.TaskRunningStatus)
					return nil
				},
				async: tt.async,
			}

			taskRunner.run()

			require.NotNil(t, recordedAtRun)
			assert.Equal(t, inventoryID, *recordedAtRun)
			require.NotNil(t, taskRunner.Task.EffectiveInventoryID)
			assert.Equal(t, inventoryID, *taskRunner.Task.EffectiveInventoryID)
		})
	}
}
