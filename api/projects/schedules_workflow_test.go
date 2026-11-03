package projects

import (
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/schedules"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddScheduleWorkflowAvailability(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			f := newResourceFixture(t)
			workflowID := f.workflow(t, "deploy")
			util.Config.Schedule = &util.ScheduleConfig{Timezone: "UTC"}
			pool := schedules.CreateSchedulePool(f.store, nil, nil, nil)
			t.Cleanup(pool.Destroy)
			c := &ScheduleController{pro: proFeatureGate{subscription: noLicense{}, getFeatures: func(*db.User, string) pro_interfaces.Features { return pro_interfaces.Features{Workflows: enabled} }}}
			r, rec := f.request(http.MethodPost, fmt.Sprintf(`{"name":"nightly","workflow_id":%d,"cron_format":"@daily","task_params":{"environment":"{\"region\":\"eu\"}"}}`, workflowID), map[string]any{"schedule_pool": pool}, nil)
			w := httptest.NewRecorder()
			c.AddSchedule(w, r)
			if !enabled {
				assert.Equal(t, http.StatusForbidden, w.Code)
				assert.Empty(t, rec.All())
				return
			}
			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			list, err := f.store.GetProjectSchedules(f.project.ID, true, false)
			require.NoError(t, err)
			require.Len(t, list, 1)
			assert.Equal(t, &workflowID, list[0].WorkflowID)
			assert.Equal(t, audit.ScheduleMetadata{WorkflowID: workflowID}, only(t, rec, audit.ResourceScheduleCreate).Event.Metadata)
		})
	}
}
