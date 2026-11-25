package projects

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	sqlstore "github.com/semaphoreui/semaphore/db/sql"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	projectService "github.com/semaphoreui/semaphore/services/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func workflowPermController(getFeatures func(*db.User, string) pro_interfaces.Features) *WorkflowPermController {
	return &WorkflowPermController{
		workflowRoles: &fakeWorkflowRoles{existing: db.WorkflowRolePerm{ID: 5, RoleID: 7, RoleSlug: "task_runner"}},
		pro:           proFeatureGate{subscription: noLicense{}, getFeatures: getFeatures},
	}
}

func TestWorkflowPerms_WritesAreEnterpriseOnly(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	builtinRole(t, store, db.ProjectTaskRunner)
	body := `{"role_slug":"task_runner","permissions":16}`
	tests := []struct {
		name        string
		getFeatures func(*db.User, string) pro_interfaces.Features
		method      string
		permID      string
		wantCode    int
	}{
		{"create without the feature", noFeatures, http.MethodPost, "", http.StatusForbidden},
		{"update without the feature", noFeatures, http.MethodPut, "5", http.StatusForbidden},
		{"delete without the feature", noFeatures, http.MethodDelete, "5", http.StatusForbidden},
		{"create with the feature", allFeatures, http.MethodPost, "", http.StatusCreated},
		{"update with the feature", allFeatures, http.MethodPut, "5", http.StatusNoContent},
		{"delete with the feature", allFeatures, http.MethodDelete, "5", http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, rec := workflowPermRequest(t, store, tt.method, body, tt.permID)
			w := httptest.NewRecorder()
			controller := workflowPermController(tt.getFeatures)
			switch tt.method {
			case http.MethodPost:
				controller.AddWorkflowPerm(w, r)
			case http.MethodPut:
				controller.UpdateWorkflowPerm(w, r)
			case http.MethodDelete:
				controller.DeleteWorkflowPerm(w, r)
			}
			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), workflowPermsEnterpriseOnly)
				assert.Empty(t, rec.All(), "a plan refusal is not an audit event")
			}
		})
	}
}

func TestGetWorkflowPerms_ReadsStayOpenWithoutTheFeature(t *testing.T) {
	store := sqlstore.InitConfigCreateTestStore()
	r, _ := workflowPermRequest(t, store, http.MethodGet, "", "5")
	w := httptest.NewRecorder()
	workflowPermController(noFeatures).GetWorkflowPerm(w, r)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func TestRestore_WorkflowGrantsAreEnterpriseOnly(t *testing.T) {
	tests := []struct {
		name    string
		roles   []projectService.BackupWorkflowRole
		refused bool
	}{
		{"a grant is refused without the feature", []projectService.BackupWorkflowRole{{Slug: "task_runner", Permissions: db.CanRunWorkflows}}, true},
		{"a workflow without grants passes", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			controller := &BackupController{
				workflowStore: proFactory.NewWorkflowStore(f.store),
				pro:           proFeatureGate{subscription: noLicense{}, getFeatures: noFeatures},
			}
			backup := projectService.BackupFormat{
				Version:   projectService.BackupFormatVersion,
				Meta:      projectService.BackupMeta{Project: db.Project{Name: "restored"}},
				Workflows: []projectService.BackupWorkflow{{WorkflowTemplate: db.WorkflowTemplate{Name: "wf"}, Roles: tt.roles}},
			}
			body, err := backup.Marshal()
			require.NoError(t, err)
			projectsBefore, err := f.store.GetAllProjects()
			require.NoError(t, err)

			r, _ := f.request(http.MethodPost, body, nil, nil)
			w := httptest.NewRecorder()
			controller.Restore(w, r)

			projectsAfter, err := f.store.GetAllProjects()
			require.NoError(t, err)
			if tt.refused {
				assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
				assert.Contains(t, w.Body.String(), workflowPermsEnterpriseOnly)
				assert.Len(t, projectsAfter, len(projectsBefore), "nothing is created")
				return
			}
			assert.NotEqual(t, http.StatusForbidden, w.Code, w.Body.String())
		})
	}
}
