package project

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type scheduleRestoreStore struct {
	db.Store
	restored db.Schedule
}

func (s *scheduleRestoreStore) CreateSchedule(schedule db.Schedule) (db.Schedule, error) {
	s.restored = schedule
	return schedule, nil
}

func TestBackupWorkflowSchedule(t *testing.T) {
	source := BackupDB{
		workflows: []db.WorkflowTemplate{{ID: 5, Name: "deploy"}},
		schedules: []db.Schedule{{ID: 7, WorkflowID: new(5), Name: "nightly", CronFormat: "@daily", TaskParams: &db.TaskParams{Environment: `{"region":"eu"}`}}},
	}
	backup, err := source.format()
	require.NoError(t, err)
	require.Len(t, backup.Schedules, 1)
	assert.Equal(t, new("deploy"), backup.Schedules[0].Workflow)
	store := &scheduleRestoreStore{}
	destination := BackupDB{meta: db.Project{ID: 20}, workflows: []db.WorkflowTemplate{{ID: 30, Name: "deploy"}}, store: store}
	require.NoError(t, backup.Schedules[0].Restore(&destination))
	assert.Equal(t, new(30), store.restored.WorkflowID)
	assert.Nil(t, store.restored.TemplateID)
	assert.Equal(t, 20, store.restored.ProjectID)
	assert.Equal(t, source.schedules[0].TaskParams.Environment, store.restored.TaskParams.Environment)
}
