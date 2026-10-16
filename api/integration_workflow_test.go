package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubWorkflowStore answers GetWorkflowTemplate from a map; every other
// method of db.WorkflowManager is left to the nil embedded interface, so a
// call the webhook must not make panics the test.
type stubWorkflowStore struct {
	db.WorkflowManager
	workflows map[int]db.WorkflowTemplate
}

func (s *stubWorkflowStore) GetWorkflowTemplate(projectID int, workflowID int) (db.WorkflowTemplate, error) {
	workflow, ok := s.workflows[workflowID]
	if !ok || workflow.ProjectID != projectID {
		return db.WorkflowTemplate{}, db.ErrNotFound
	}
	return workflow, nil
}

// stubWorkflowService records how the webhook started a run.
type stubWorkflowService struct {
	pro_interfaces.WorkflowService
	calls []stubWorkflowStart
	err   error
	run   db.WorkflowRun
}

type stubWorkflowStart struct {
	actor    audit.Actor
	workflow db.WorkflowTemplate
	params   db.WorkflowRunParams
	user     *db.User
}

func (s *stubWorkflowService) StartWorkflow(ctx context.Context, workflow db.WorkflowTemplate, params db.WorkflowRunParams, user *db.User) (db.WorkflowRun, error) {
	s.calls = append(s.calls, stubWorkflowStart{actor: audit.ActorFrom(ctx), workflow: workflow, params: params, user: user})
	if s.err != nil {
		return db.WorkflowRun{}, s.err
	}
	return s.run, nil
}

type workflowWebhookFixture struct {
	store       *sql.SqlDb
	project     db.Project
	integration db.Integration
	workflowID  int
	alias       string
}

// newWorkflowWebhookFixture creates a project with a workflow row (inserted
// by hand: the open-source workflow store persists nothing) and an
// integration that starts it through its own alias, with two extractors.
func newWorkflowWebhookFixture(t *testing.T) workflowWebhookFixture {
	t.Helper()
	store := sql.InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "workflow webhook project"})
	require.NoError(t, err)

	_, err = store.Sql().Exec(store.PrepareQuery(
		"insert into project__workflow_template (project_id, name) values (?, ?)"), project.ID, "deploy")
	require.NoError(t, err)
	workflowID64, err := store.Sql().SelectInt(store.PrepareQuery(
		"select id from project__workflow_template where project_id = ? and name = ?"), project.ID, "deploy")
	require.NoError(t, err)
	workflowID := int(workflowID64)

	integration, err := store.CreateIntegration(db.Integration{
		Name:       "deploy hook",
		ProjectID:  project.ID,
		WorkflowID: &workflowID,
		AuthMethod: db.IntegrationAuthNone,
	})
	require.NoError(t, err)

	for _, value := range []db.IntegrationExtractValue{
		{Name: "ref", IntegrationID: integration.ID, ValueSource: db.IntegrationExtractHeaderValue, Key: "X-Ref", Variable: "ref", VariableType: db.IntegrationVariableEnvironment},
		{Name: "tag", IntegrationID: integration.ID, ValueSource: db.IntegrationExtractBodyValue, BodyDataType: db.IntegrationBodyDataJSON, Key: "release.tag", Variable: "image_tag", VariableType: db.IntegrationVariableTaskParam},
	} {
		_, err = store.CreateIntegrationExtractValue(project.ID, value)
		require.NoError(t, err)
	}

	alias := "workflow-alias"
	_, err = store.CreateIntegrationAlias(db.IntegrationAlias{Alias: alias, ProjectID: project.ID, IntegrationID: &integration.ID})
	require.NoError(t, err)

	return workflowWebhookFixture{store: store, project: project, integration: integration, workflowID: workflowID, alias: alias}
}

func (f workflowWebhookFixture) receive(t *testing.T, ctrl *IntegrationController) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/"+f.alias, bytes.NewReader([]byte(`{"release":{"tag":"1.4.2"}}`)))
	req.Header.Set("X-Ref", "refs/heads/main")
	req = mux.SetURLVars(req, map[string]string{"integration_alias": f.alias})

	w := httptest.NewRecorder()
	ctrl.ReceiveIntegration(w, req)
	return w
}

func TestReceiveIntegration_StartsWorkflowRun(t *testing.T) {
	f := newWorkflowWebhookFixture(t)
	rootTaskID := 77
	svc := &stubWorkflowService{run: db.WorkflowRun{ID: 42, ProjectID: f.project.ID, WorkflowTemplateID: f.workflowID, RootTaskID: &rootTaskID}}
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{
		f.workflowID: {ID: f.workflowID, ProjectID: f.project.ID, Name: "deploy"},
	}}
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, wfStore, svc)

	w := f.receive(t, ctrl)

	assert.Equal(t, http.StatusNoContent, w.Code)
	require.Len(t, svc.calls, 1)
	call := svc.calls[0]
	assert.Equal(t, f.workflowID, call.workflow.ID)
	assert.Nil(t, call.user, "a webhook has no user")
	assert.Equal(t, audit.IntegrationActor(f.integration.ID, f.integration.Name), call.actor)
	require.NotNil(t, call.params.IntegrationID)
	assert.Equal(t, f.integration.ID, *call.params.IntegrationID)
	assert.Equal(t, db.MapStringAnyField{"ref": "refs/heads/main", "image_tag": "1.4.2"}, call.params.Environment,
		"both extractor variable types become survey answers")

	assert.Equal(t, "42", w.Header().Get("X-Semaphore-Workflow-Run-ID"))
	assert.Equal(t, "77", w.Header().Get("X-Semaphore-Task-ID"))
	assert.Empty(t, w.Header().Get("X-Semaphore-Template-ID"), "no template header for a workflow")
	assert.NotEmpty(t, w.Header().Get("X-Semaphore-Workflow-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Semaphore-Integration-ID"))
}

func TestReceiveIntegration_WorkflowStartRefused(t *testing.T) {
	f := newWorkflowWebhookFixture(t)
	svc := &stubWorkflowService{err: common_errors.NewValidationError(`workflow variable "env" is required`)}
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{
		f.workflowID: {ID: f.workflowID, ProjectID: f.project.ID, Name: "deploy"},
	}}
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, wfStore, svc)

	w := f.receive(t, ctrl)

	assert.Equal(t, http.StatusNoContent, w.Code, "a refused start is logged, the webhook still answers 204")
	assert.Len(t, svc.calls, 1)
	assert.Empty(t, w.Header().Get("X-Semaphore-Workflow-Run-ID"))
	assert.Empty(t, w.Header().Get("X-Semaphore-Task-ID"))
}

func TestReceiveIntegration_WorkflowUnavailableInOpenSourceBuild(t *testing.T) {
	f := newWorkflowWebhookFixture(t)
	svc := &stubWorkflowService{}
	// The open-source store answers with an empty template and no error.
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{f.workflowID: {}}}
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, wfStore, svc)

	w := f.receive(t, ctrl)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, svc.calls, "nothing to start without a workflow")
	assert.Empty(t, w.Header().Get("X-Semaphore-Workflow-Run-ID"))
}

func TestReceiveIntegration_WorkflowLookupError(t *testing.T) {
	f := newWorkflowWebhookFixture(t)
	svc := &stubWorkflowService{}
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{}}
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, wfStore, svc)

	w := f.receive(t, ctrl)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, svc.calls)
}

func TestGetWorkflowRunParams(t *testing.T) {
	integrationID := 5
	integration := db.Integration{ID: integrationID, ProjectID: 1, WorkflowID: new(9)}
	header := http.Header{}
	header.Set("X-Ref", "main")
	payload := []byte(`{"n":3}`)

	t.Run("values of both types under their variable names", func(t *testing.T) {
		params, err := GetWorkflowRunParams(integration, payload, header, func(int, int) ([]db.IntegrationExtractValue, error) {
			return []db.IntegrationExtractValue{
				{ValueSource: db.IntegrationExtractHeaderValue, Key: "X-Ref", Variable: "ref", VariableType: db.IntegrationVariableEnvironment},
				{ValueSource: db.IntegrationExtractBodyValue, BodyDataType: db.IntegrationBodyDataJSON, Key: "n", Variable: "count", VariableType: db.IntegrationVariableTaskParam},
			}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, &integrationID, params.IntegrationID)
		assert.Equal(t, "main", params.Environment["ref"])
		assert.EqualValues(t, 3, params.Environment["count"])
	})

	t.Run("no extractors means no answers", func(t *testing.T) {
		params, err := GetWorkflowRunParams(integration, payload, header, func(int, int) ([]db.IntegrationExtractValue, error) {
			return nil, nil
		})
		require.NoError(t, err)
		assert.Nil(t, params.Environment)
		assert.Equal(t, &integrationID, params.IntegrationID)
	})

	t.Run("extractor lookup error is returned", func(t *testing.T) {
		_, err := GetWorkflowRunParams(integration, payload, header, func(int, int) ([]db.IntegrationExtractValue, error) {
			return nil, errors.New("boom")
		})
		assert.ErrorContains(t, err, "boom")
	})
}
