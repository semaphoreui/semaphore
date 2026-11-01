package projects

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	sqlstore "github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
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
