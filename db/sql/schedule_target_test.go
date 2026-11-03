package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSchedule_WorkflowTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	created, err := store.CreateSchedule(db.Schedule{
		Name:       "nightly workflow",
		ProjectID:  project.ID,
		WorkflowID: &workflowID,
		CronFormat: "0 0 * * *",
		Active:     true,
		TaskParams: &db.TaskParams{Message: "ignored for workflows"},
	})
	require.NoError(t, err)
	assert.Nil(t, created.TemplateID)
	assert.Nil(t, created.TaskParamsID, "task params do not apply to a workflow target")

	loaded, err := store.GetSchedule(project.ID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.WorkflowID)
	assert.Equal(t, workflowID, *loaded.WorkflowID)
	assert.Nil(t, loaded.TemplateID)
	assert.Nil(t, loaded.TaskParams)
	assert.True(t, loaded.RunsWorkflow())

	// The cron table loads every schedule, whatever its target.
	all, err := store.GetSchedules()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, created.ID, all[0].ID)
}

func TestGetProjectSchedules_NamesEitherTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	_, err := store.CreateSchedule(db.Schedule{Name: "tpl", ProjectID: project.ID, TemplateID: &template.ID, CronFormat: "0 0 * * *"})
	require.NoError(t, err)
	_, err = store.CreateSchedule(db.Schedule{Name: "wf", ProjectID: project.ID, WorkflowID: &workflowID, CronFormat: "0 1 * * *"})
	require.NoError(t, err)

	list, err := store.GetProjectSchedules(project.ID, true, true)
	require.NoError(t, err)
	require.Len(t, list, 2)

	byName := map[string]db.ScheduleWithTpl{}
	for _, s := range list {
		byName[s.Name] = s
	}
	assert.Equal(t, template.Name, byName["tpl"].TemplateName)
	assert.Empty(t, byName["tpl"].WorkflowName, "a template schedule names only its template")
	assert.Equal(t, "wf", byName["wf"].WorkflowName)
	assert.Empty(t, byName["wf"].TemplateName)

	// GetTemplateSchedules keys on template_id and never sees a workflow schedule.
	tplSchedules, err := store.GetTemplateSchedules(project.ID, template.ID, false)
	require.NoError(t, err)
	require.Len(t, tplSchedules, 1)
	assert.Equal(t, "tpl", tplSchedules[0].Name)
}

func TestCreateSchedule_RejectsBothOrNoTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	_, err := store.CreateSchedule(db.Schedule{Name: "none", ProjectID: project.ID, CronFormat: "0 0 * * *"})
	assertValidationError(t, err, "Schedule must run a template or a workflow")

	_, err = store.CreateSchedule(db.Schedule{
		Name: "both", ProjectID: project.ID, TemplateID: &template.ID, WorkflowID: &workflowID, CronFormat: "0 0 * * *",
	})
	assertValidationError(t, err, "not both")

	repoID := template.RepositoryID
	_, err = store.CreateSchedule(db.Schedule{
		Name: "checker", ProjectID: project.ID, WorkflowID: &workflowID, RepositoryID: &repoID, CronFormat: "0 0 * * *",
	})
	assertValidationError(t, err, "cannot check a repository")
}

func TestCreateSchedule_TargetMustBelongToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	_, err = store.CreateSchedule(db.Schedule{Name: "foreign tpl", ProjectID: other.ID, TemplateID: &template.ID, CronFormat: "0 0 * * *"})
	assertValidationError(t, err, "Template does not exist in the project")

	_, err = store.CreateSchedule(db.Schedule{Name: "foreign wf", ProjectID: other.ID, WorkflowID: &workflowID, CronFormat: "0 0 * * *"})
	assertValidationError(t, err, "Workflow does not exist in the project")

	missing := workflowID + 1000
	_, err = store.CreateSchedule(db.Schedule{Name: "missing wf", ProjectID: project.ID, WorkflowID: &missing, CronFormat: "0 0 * * *"})
	assertValidationError(t, err, "Workflow does not exist in the project")
}

func TestUpdateSchedule_SwitchingTargetDropsTaskParams(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, template := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	created, err := store.CreateSchedule(db.Schedule{
		Name: "nightly", ProjectID: project.ID, TemplateID: &template.ID, CronFormat: "0 0 * * *",
		TaskParams: &db.TaskParams{Message: "deploy"},
	})
	require.NoError(t, err)
	require.NotNil(t, created.TaskParamsID)
	paramsID := *created.TaskParamsID

	// Template to workflow: the params row is unlinked and removed.
	err = store.UpdateSchedule(db.Schedule{
		ID: created.ID, Name: "nightly", ProjectID: project.ID, WorkflowID: &workflowID, CronFormat: "0 0 * * *",
		TaskParams: &db.TaskParams{Message: "ignored"},
	})
	require.NoError(t, err)

	updated, err := store.GetSchedule(project.ID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.TemplateID)
	require.NotNil(t, updated.WorkflowID)
	assert.Equal(t, workflowID, *updated.WorkflowID)
	assert.Nil(t, updated.TaskParamsID)
	count, err := store.Sql().SelectInt("select count(*) from project__task_params where id = ?", paramsID)
	require.NoError(t, err)
	assert.Zero(t, count, "the old task params row is deleted")

	// Workflow back to template, with new params.
	err = store.UpdateSchedule(db.Schedule{
		ID: created.ID, Name: "nightly", ProjectID: project.ID, TemplateID: &template.ID, CronFormat: "0 0 * * *",
		TaskParams: &db.TaskParams{Message: "again"},
	})
	require.NoError(t, err)
	back, err := store.GetSchedule(project.ID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, back.WorkflowID)
	require.NotNil(t, back.TemplateID)
	require.NotNil(t, back.TaskParams)
	assert.Equal(t, "again", back.TaskParams.Message)
}

func TestDeleteWorkflow_CascadesToItsSchedules(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestTemplate(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	created, err := store.CreateSchedule(db.Schedule{Name: "wf", ProjectID: project.ID, WorkflowID: &workflowID, CronFormat: "0 0 * * *"})
	require.NoError(t, err)

	_, err = store.Sql().Exec(store.PrepareQuery("delete from project__workflow_template where id = ?"), workflowID)
	require.NoError(t, err)

	_, err = store.GetSchedule(project.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}
