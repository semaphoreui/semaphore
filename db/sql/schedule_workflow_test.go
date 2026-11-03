package sql

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScheduleWorkflowTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "deploy")
	s, err := store.CreateSchedule(db.Schedule{ProjectID: project.ID, Name: "nightly", WorkflowID: &workflowID, CronFormat: "@daily", Active: true, TaskParams: &db.TaskParams{Environment: `{"region":"eu"}`}})
	require.NoError(t, err)
	loaded, err := store.GetSchedule(project.ID, s.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.TemplateID)
	assert.Equal(t, &workflowID, loaded.WorkflowID)
	assert.Equal(t, s.TaskParams.Environment, loaded.TaskParams.Environment)
	list, err := store.GetProjectSchedules(project.ID, true, false)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "deploy", list[0].TemplateName)
	s.WorkflowID, s.TemplateID = nil, &template.ID
	require.NoError(t, store.UpdateSchedule(s))
	loaded, err = store.GetSchedule(project.ID, s.ID)
	require.NoError(t, err)
	assert.Equal(t, &template.ID, loaded.TemplateID)
	assert.Nil(t, loaded.WorkflowID)
	s.TemplateID, s.WorkflowID = nil, &workflowID
	require.NoError(t, store.UpdateSchedule(s))
	_, err = store.Sql().Exec(store.PrepareQuery("delete from project__workflow_template where id=?"), workflowID)
	require.NoError(t, err)
	_, err = store.GetSchedule(project.ID, s.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestScheduleRejectsInvalidTargets(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	workflowID := createTestWorkflow(t, store, other.ID, "foreign")
	for _, tc := range []struct {
		name     string
		schedule db.Schedule
	}{
		{"no target", db.Schedule{ProjectID: project.ID}},
		{"both targets", db.Schedule{ProjectID: project.ID, TemplateID: &template.ID, WorkflowID: &workflowID}},
		{"foreign workflow", db.Schedule{ProjectID: project.ID, WorkflowID: &workflowID}},
		{"foreign template", db.Schedule{ProjectID: other.ID, TemplateID: &template.ID}},
		{"commit checker", db.Schedule{ProjectID: other.ID, WorkflowID: &workflowID, RepositoryID: new(1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.CreateSchedule(tc.schedule)
			assert.Error(t, err)
			assert.Error(t, store.UpdateSchedule(tc.schedule))
		})
	}
}

func TestMigration_2_20_17_Schedules(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)
	project, template := createTestTemplate(t, store)
	_, err := store.Sql().Exec(store.PrepareQuery("insert into project__schedule (project_id, template_id, name, cron_format) values (?, ?, 'old', '@daily')"), project.ID, template.ID)
	require.NoError(t, err)
	id, err := store.Sql().SelectInt("select id from project__schedule where name='old'")
	require.NoError(t, err)
	scheduleID := int(id)
	task, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: template.ID, ScheduleID: &scheduleID, Status: "success"}, 0)
	require.NoError(t, err)
	require.NoError(t, db.Migrate(store, nil))
	s, err := store.GetSchedule(project.ID, scheduleID)
	require.NoError(t, err)
	assert.Equal(t, &template.ID, s.TemplateID)
	loaded, err := store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	assert.Equal(t, &scheduleID, loaded.ScheduleID)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")
	_, err = store.CreateSchedule(db.Schedule{ProjectID: project.ID, WorkflowID: &workflowID, CronFormat: "@daily"})
	require.NoError(t, err)
	require.NoError(t, db.Rollback(store, before))
	count, err := store.Sql().SelectInt("select count(*) from project__schedule")
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	loaded, err = store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	assert.Equal(t, &scheduleID, loaded.ScheduleID)
}
