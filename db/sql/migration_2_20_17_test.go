package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_17 seeds a schedule with a task that refers to it on the
// pre-workflow schema and checks the upgrade (a table rebuild on SQLite)
// keeps those rows and lets a schedule point at a workflow.
func TestMigration_2_20_17(t *testing.T) {
	target := "2.20.16"
	store := InitConfigCreateTestStoreAt(&target)

	project, template := createTestTemplate(t, store)

	// The store already writes workflow_id, which the old schema lacks, so the
	// schedule is seeded by hand.
	_, err := store.Sql().Exec(store.PrepareQuery(
		"insert into project__schedule (project_id, template_id, cron_format, name, active, type, delete_after_run) values (?, ?, ?, ?, 1, '', 0)"),
		project.ID, template.ID, "0 0 * * *", "old schedule")
	require.NoError(t, err)
	scheduleID, err := store.Sql().SelectInt(store.PrepareQuery(
		"select id from project__schedule where project_id = ? and name = ?"), project.ID, "old schedule")
	require.NoError(t, err)
	oldID := int(scheduleID)

	task, err := store.CreateTask(db.Task{
		ProjectID: project.ID, TemplateID: template.ID, ScheduleID: &oldID, Status: "success",
	}, 0)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	migrated, err := store.GetSchedule(project.ID, oldID)
	require.NoError(t, err)
	require.NotNil(t, migrated.TemplateID)
	assert.Equal(t, template.ID, *migrated.TemplateID)
	assert.Nil(t, migrated.WorkflowID)
	assert.Equal(t, "old schedule", migrated.Name)
	assert.True(t, migrated.Active)

	migratedTask, err := store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, migratedTask.ScheduleID)
	assert.Equal(t, oldID, *migratedTask.ScheduleID)

	workflowID := createTestWorkflow(t, store, project.ID, "wf")
	workflowSchedule, err := store.CreateSchedule(db.Schedule{
		Name: "workflow schedule", ProjectID: project.ID, WorkflowID: &workflowID, CronFormat: "0 1 * * *",
	})
	require.NoError(t, err)
	assert.Nil(t, workflowSchedule.TemplateID)

	// The foreign key on schedule_id still sets null: deleting the schedule
	// leaves its tasks behind, as before the rebuild.
	require.NoError(t, store.DeleteSchedule(project.ID, oldID))
	orphanedTask, err := store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	assert.Nil(t, orphanedTask.ScheduleID, "task.schedule_id must be set null, not cascade-deleted")
}
