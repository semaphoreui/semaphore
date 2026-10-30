package projects

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	projectService "github.com/semaphoreui/semaphore/services/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noLicense is a subscription service without a license: the plan is "" and every feature is
// whatever the build's feature table gives for it.
type noLicense struct {
	pro_interfaces.SubscriptionService
}

func (noLicense) GetToken() (pro_interfaces.SubscriptionToken, error) {
	return pro_interfaces.SubscriptionToken{}, db.ErrNotFound
}

// brokenLicense is a subscription store that cannot be read.
type brokenLicense struct {
	pro_interfaces.SubscriptionService
}

func (brokenLicense) GetToken() (pro_interfaces.SubscriptionToken, error) {
	return pro_interfaces.SubscriptionToken{}, errors.New("subscription store is down")
}

// allFeatures and noFeatures stand in for the build's feature table, so the tests hold for the
// open-source stub and for Pro alike.
func allFeatures(*db.User, string) pro_interfaces.Features {
	return pro_interfaces.Features{Workflows: true, WorkflowIntegrationsManagement: true}
}

func noFeatures(*db.User, string) pro_interfaces.Features {
	return pro_interfaces.Features{Workflows: true}
}

func integrationController(getFeatures func(*db.User, string) pro_interfaces.Features) *IntegrationController {
	return &IntegrationController{pro: proFeatureGate{subscription: noLicense{}, getFeatures: getFeatures}}
}

// workflow inserts a workflow template directly: the workflow store is a Pro feature, the table
// is in the community schema.
func (f resourceFixture) workflow(t *testing.T, name string) int {
	t.Helper()
	_, err := f.store.Sql().Exec(f.store.PrepareQuery(
		"insert into project__workflow_template (project_id, name) values (?, ?)"), f.project.ID, name)
	require.NoError(t, err)
	id, err := f.store.Sql().SelectInt(f.store.PrepareQuery(
		"select id from project__workflow_template where project_id = ? and name = ?"), f.project.ID, name)
	require.NoError(t, err)
	return int(id)
}

func TestAddIntegration_WorkflowTargetIsProOnly(t *testing.T) {
	tests := []struct {
		name        string
		getFeatures func(*db.User, string) pro_interfaces.Features
		workflow    bool
		wantCode    int
	}{
		{"template target without the feature", noFeatures, false, http.StatusCreated},
		{"workflow target without the feature", noFeatures, true, http.StatusForbidden},
		{"workflow target with the feature", allFeatures, true, http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			target := `"template_id":` + strconv.Itoa(f.template(t).ID)
			if tt.workflow {
				target = `"workflow_id":` + strconv.Itoa(f.workflow(t, "wf"))
			}
			r, rec := f.request(http.MethodPost, `{"name":"hook","project_id":`+strconv.Itoa(f.project.ID)+`,`+target+`,"auth_method":"none"}`, nil, nil)
			w := httptest.NewRecorder()

			integrationController(tt.getFeatures).AddIntegration(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			integrations, err := f.store.GetIntegrations(f.project.ID, db.RetrieveQueryParams{}, false)
			require.NoError(t, err)
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), workflowIntegrationsProOnly)
				assert.Empty(t, integrations, "nothing is created")
				assert.Empty(t, rec.All(), "a plan refusal is not an audit deny event")
				return
			}
			assert.Len(t, integrations, 1)
		})
	}
}

func TestAddIntegration_SubscriptionFailureIs500(t *testing.T) {
	f := newResourceFixture(t)
	controller := &IntegrationController{pro: proFeatureGate{subscription: brokenLicense{}, getFeatures: allFeatures}}
	r, _ := f.request(http.MethodPost, `{"name":"hook","project_id":`+strconv.Itoa(f.project.ID)+`,"workflow_id":`+strconv.Itoa(f.workflow(t, "wf"))+`,"auth_method":"none"}`, nil, nil)
	w := httptest.NewRecorder()

	controller.AddIntegration(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
}

func TestUpdateIntegration_WorkflowTargetIsProOnly(t *testing.T) {
	tests := []struct {
		name         string
		getFeatures  func(*db.User, string) pro_interfaces.Features
		oldWorkflow  bool // the stored integration runs workflow "wf"
		newTarget    string
		wantCode     int
		wantWorkflow string // name of the workflow the integration runs afterwards, "" for the template
	}{
		{"template to workflow without the feature", noFeatures, false, "wf", http.StatusForbidden, ""},
		{"template to workflow with the feature", allFeatures, false, "wf", http.StatusNoContent, "wf"},
		{"same workflow without the feature", noFeatures, true, "wf", http.StatusNoContent, "wf"},
		{"another workflow without the feature", noFeatures, true, "other", http.StatusForbidden, "wf"},
		{"another workflow with the feature", allFeatures, true, "other", http.StatusNoContent, "other"},
		{"workflow back to template without the feature", noFeatures, true, "", http.StatusNoContent, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			template := f.template(t)
			workflows := map[string]int{"wf": f.workflow(t, "wf"), "other": f.workflow(t, "other")}

			old := db.Integration{Name: "hook", ProjectID: f.project.ID, AuthMethod: db.IntegrationAuthNone}
			if tt.oldWorkflow {
				old.WorkflowID = new(workflows["wf"])
			} else {
				old.TemplateID = &template.ID
			}
			old, err := f.store.CreateIntegration(old)
			require.NoError(t, err)

			target := `"template_id":` + strconv.Itoa(template.ID)
			if tt.newTarget != "" {
				target = `"workflow_id":` + strconv.Itoa(workflows[tt.newTarget])
			}
			body := `{"id":` + strconv.Itoa(old.ID) + `,"name":"renamed","project_id":` + strconv.Itoa(f.project.ID) + `,` + target + `,"auth_method":"none"}`
			r, rec := f.request(http.MethodPut, body, map[string]any{"integration": old}, nil)
			w := httptest.NewRecorder()

			integrationController(tt.getFeatures).UpdateIntegration(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			stored, err := f.store.GetIntegration(f.project.ID, old.ID)
			require.NoError(t, err)
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), workflowIntegrationsProOnly)
				assert.Equal(t, "hook", stored.Name, "nothing is changed")
				assert.Empty(t, rec.All())
			} else {
				assert.Equal(t, "renamed", stored.Name)
				assert.Len(t, rec.All(), 1)
			}
			if tt.wantWorkflow == "" {
				assert.Nil(t, stored.WorkflowID)
				assert.Equal(t, template.ID, *stored.TemplateID)
			} else {
				require.NotNil(t, stored.WorkflowID)
				assert.Equal(t, workflows[tt.wantWorkflow], *stored.WorkflowID)
			}
		})
	}
}

func TestRestore_WorkflowIntegrationIsProOnly(t *testing.T) {
	f := newResourceFixture(t)
	controller := &BackupController{
		workflowStore: proFactory.NewWorkflowStore(f.store),
		pro:           proFeatureGate{subscription: noLicense{}, getFeatures: noFeatures},
	}
	backup := projectService.BackupFormat{
		Version:     projectService.BackupFormatVersion,
		Meta:        projectService.BackupMeta{Project: db.Project{Name: "restored"}},
		Workflows:   []projectService.BackupWorkflow{{WorkflowTemplate: db.WorkflowTemplate{Name: "wf"}}},
		Integration: []projectService.BackupIntegration{{Integration: db.Integration{Name: "hook"}, Workflow: new("wf")}},
	}
	body, err := backup.Marshal()
	require.NoError(t, err)
	projectsBefore, err := f.store.GetAllProjects()
	require.NoError(t, err)

	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	controller.Restore(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), workflowIntegrationsProOnly)
	projectsAfter, err := f.store.GetAllProjects()
	require.NoError(t, err)
	assert.Len(t, projectsAfter, len(projectsBefore), "nothing is created")
	assert.Empty(t, rec.All(), "no restore event, no deny event")
}
