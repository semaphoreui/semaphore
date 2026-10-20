package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestWorkflow inserts a workflow template row directly: the workflow
// store is a Pro feature and the open-source stub does not persist anything.
func createTestWorkflow(t *testing.T, store *SqlDb, projectID int, name string) int {
	t.Helper()
	_, err := store.Sql().Exec(
		store.PrepareQuery("insert into project__workflow_template (project_id, name) values (?, ?)"), projectID, name)
	require.NoError(t, err)
	id, err := store.Sql().SelectInt(
		store.PrepareQuery("select id from project__workflow_template where project_id = ? and name = ?"), projectID, name)
	require.NoError(t, err)
	return int(id)
}

func assertValidationError(t *testing.T, err error, msg string) {
	t.Helper()
	var validationErr *common_errors.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.ErrorContains(t, err, msg)
}

func TestCreateIntegration_WorkflowTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	created, err := store.CreateIntegration(db.Integration{
		Name:       "workflow hook",
		ProjectID:  project.ID,
		WorkflowID: &workflowID,
		TaskParams: &db.TaskParams{Message: "ignored for workflows"},
	})
	require.NoError(t, err)
	assert.Nil(t, created.TemplateID)
	assert.Nil(t, created.TaskParamsID, "task params do not apply to a workflow target")

	loaded, err := store.GetIntegration(project.ID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.WorkflowID)
	assert.Equal(t, workflowID, *loaded.WorkflowID)
	assert.Nil(t, loaded.TemplateID)
	assert.Nil(t, loaded.TaskParams)

	list, err := store.GetIntegrations(project.ID, db.RetrieveQueryParams{}, true)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestCreateIntegration_RejectsBothOrNoTarget(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, existing := createTestIntegration(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	_, err := store.CreateIntegration(db.Integration{Name: "none", ProjectID: project.ID})
	assertValidationError(t, err, "Integration must run a template or a workflow")

	_, err = store.CreateIntegration(db.Integration{
		Name: "both", ProjectID: project.ID, TemplateID: existing.TemplateID, WorkflowID: &workflowID,
	})
	assertValidationError(t, err, "not both")
}

func TestCreateIntegration_TargetMustBelongToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, existing := createTestIntegration(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	_, err = store.CreateIntegration(db.Integration{
		Name: "foreign template", ProjectID: other.ID, TemplateID: existing.TemplateID,
	})
	assertValidationError(t, err, "Template does not exist in the project")

	_, err = store.CreateIntegration(db.Integration{
		Name: "foreign workflow", ProjectID: other.ID, WorkflowID: &workflowID,
	})
	assertValidationError(t, err, "Workflow does not exist in the project")

	missing := workflowID + 1000
	_, err = store.CreateIntegration(db.Integration{
		Name: "missing workflow", ProjectID: project.ID, WorkflowID: &missing,
	})
	assertValidationError(t, err, "Workflow does not exist in the project")
}

func TestUpdateIntegration_SwitchTemplateToWorkflowDropsTaskParams(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	integration.TaskParams = &db.TaskParams{Message: "hello"}
	require.NoError(t, store.UpdateIntegration(integration))

	withParams, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	require.NotNil(t, withParams.TaskParamsID)
	paramsID := *withParams.TaskParamsID

	withParams.TemplateID = nil
	withParams.WorkflowID = &workflowID
	require.NoError(t, store.UpdateIntegration(withParams))

	switched, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	assert.Nil(t, switched.TemplateID)
	require.NotNil(t, switched.WorkflowID)
	assert.Equal(t, workflowID, *switched.WorkflowID)
	assert.Nil(t, switched.TaskParamsID)

	var orphan db.TaskParams
	err = store.getObject(project.ID, db.TaskParamsProps, paramsID, &orphan)
	assert.ErrorIs(t, err, db.ErrNotFound, "the old task params row must be removed")

	// And back to a template: the workflow link is cleared.
	switched.WorkflowID = nil
	switched.TemplateID = integration.TemplateID
	require.NoError(t, store.UpdateIntegration(switched))

	back, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	assert.Nil(t, back.WorkflowID)
	require.NotNil(t, back.TemplateID)
	assert.Equal(t, *integration.TemplateID, *back.TemplateID)
}

func TestUpdateIntegration_WithoutTaskParamsKeepsExistingOnes(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)

	integration.TaskParams = &db.TaskParams{Message: "keep me"}
	require.NoError(t, store.UpdateIntegration(integration))

	// An update that carries no task params (an API client sending only the
	// fields it changes) must not unlink the stored ones.
	renamed := integration
	renamed.Name = "renamed"
	renamed.TaskParams = nil
	renamed.TaskParamsID = nil
	require.NoError(t, store.UpdateIntegration(renamed))

	loaded, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", loaded.Name)
	require.NotNil(t, loaded.TaskParams)
	assert.Equal(t, "keep me", loaded.TaskParams.Message)
}

func TestUpdateIntegration_TargetMustBelongToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)

	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)
	foreignWorkflowID := createTestWorkflow(t, store, other.ID, "foreign wf")

	integration.TemplateID = nil
	integration.WorkflowID = &foreignWorkflowID
	err = store.UpdateIntegration(integration)
	assertValidationError(t, err, "Workflow does not exist in the project")

	loaded, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.WorkflowID, "a refused update must not change the row")
	assert.NotNil(t, loaded.TemplateID)
}

func TestDeleteWorkflow_CascadesToIntegration(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)
	workflowID := createTestWorkflow(t, store, project.ID, "wf")

	created, err := store.CreateIntegration(db.Integration{
		Name: "workflow hook", ProjectID: project.ID, WorkflowID: &workflowID,
	})
	require.NoError(t, err)

	_, err = store.Sql().Exec(store.PrepareQuery("delete from project__workflow_template where id = ?"), workflowID)
	require.NoError(t, err)

	_, err = store.GetIntegration(project.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}
