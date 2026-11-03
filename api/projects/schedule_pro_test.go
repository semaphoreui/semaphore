package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	projectService "github.com/semaphoreui/semaphore/services/project"
	"github.com/semaphoreui/semaphore/services/schedules"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scheduleController(getFeatures func(*db.User, string) pro_interfaces.Features) *ScheduleController {
	return &ScheduleController{pro: proFeatureGate{subscription: noLicense{}, getFeatures: getFeatures}}
}

// schedulePool is the pool the handlers refresh after a change.
func (f resourceFixture) schedulePool(t *testing.T) schedules.SchedulePool {
	t.Helper()
	util.Config.Schedule = &util.ScheduleConfig{Timezone: "UTC"}
	pool := schedules.CreateSchedulePool(f.store, nil, nil, nil)
	t.Cleanup(pool.Destroy)
	return pool
}

func TestAddSchedule_WorkflowTargetIsProOnly(t *testing.T) {
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
			pool := f.schedulePool(t)
			target := `"template_id":` + strconv.Itoa(f.template(t).ID)
			if tt.workflow {
				target = `"workflow_id":` + strconv.Itoa(f.workflow(t, "wf"))
			}
			body := `{"name":"nightly","project_id":` + strconv.Itoa(f.project.ID) + `,` + target + `,"cron_format":"0 0 * * *","active":true}`
			r, rec := f.request(http.MethodPost, body, map[string]any{"schedule_pool": pool}, nil)
			w := httptest.NewRecorder()

			scheduleController(tt.getFeatures).AddSchedule(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			stored, err := f.store.GetProjectSchedules(f.project.ID, false, true)
			require.NoError(t, err)
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), workflowSchedulesProOnly)
				assert.Empty(t, stored, "nothing is created")
				assert.Empty(t, rec.All(), "a plan refusal is not an audit deny event")
				return
			}
			require.Len(t, stored, 1)
			created := only(t, rec, audit.ResourceScheduleCreate)
			if tt.workflow {
				assert.Equal(t, "wf", stored[0].WorkflowName)
				assert.Empty(t, stored[0].TemplateName)
				assert.Equal(t, audit.ScheduleMetadata{WorkflowID: *stored[0].WorkflowID}, created.Event.Metadata)
			} else {
				assert.Equal(t, "deploy", stored[0].TemplateName)
				assert.Empty(t, stored[0].WorkflowName)
				assert.Equal(t, audit.ScheduleMetadata{TemplateID: *stored[0].TemplateID}, created.Event.Metadata)
			}
		})
	}
}

func TestAddSchedule_RejectsBothOrNoTarget(t *testing.T) {
	f := newResourceFixture(t)
	pool := f.schedulePool(t)
	template := f.template(t)
	workflowID := f.workflow(t, "wf")

	tests := []struct {
		name   string
		target string
	}{
		{"no target", `"template_id":null`},
		{"both targets", `"template_id":` + strconv.Itoa(template.ID) + `,"workflow_id":` + strconv.Itoa(workflowID)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"name":"nightly","project_id":` + strconv.Itoa(f.project.ID) + `,` + tt.target + `,"cron_format":"0 0 * * *"}`
			r, _ := f.request(http.MethodPost, body, map[string]any{"schedule_pool": pool}, nil)
			w := httptest.NewRecorder()

			scheduleController(allFeatures).AddSchedule(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			stored, err := f.store.GetProjectSchedules(f.project.ID, false, true)
			require.NoError(t, err)
			assert.Empty(t, stored)
		})
	}
}

func TestAddSchedule_SubscriptionFailureIs500(t *testing.T) {
	f := newResourceFixture(t)
	pool := f.schedulePool(t)
	controller := &ScheduleController{pro: proFeatureGate{subscription: brokenLicense{}, getFeatures: allFeatures}}
	body := `{"name":"nightly","project_id":` + strconv.Itoa(f.project.ID) + `,"workflow_id":` + strconv.Itoa(f.workflow(t, "wf")) + `,"cron_format":"0 0 * * *"}`
	r, _ := f.request(http.MethodPost, body, map[string]any{"schedule_pool": pool}, nil)
	w := httptest.NewRecorder()

	controller.AddSchedule(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
}

func TestUpdateSchedule_WorkflowTargetIsProOnly(t *testing.T) {
	tests := []struct {
		name         string
		getFeatures  func(*db.User, string) pro_interfaces.Features
		oldWorkflow  bool // the stored schedule runs workflow "wf"
		newTarget    string
		wantCode     int
		wantWorkflow string // name of the workflow the schedule runs afterwards, "" for the template
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
			pool := f.schedulePool(t)
			template := f.template(t)
			workflows := map[string]int{"wf": f.workflow(t, "wf"), "other": f.workflow(t, "other")}

			old := db.Schedule{Name: "nightly", ProjectID: f.project.ID, CronFormat: "0 0 * * *", Active: true}
			if tt.oldWorkflow {
				old.WorkflowID = new(workflows["wf"])
			} else {
				old.TemplateID = &template.ID
			}
			old, err := f.store.CreateSchedule(old)
			require.NoError(t, err)

			target := `"template_id":` + strconv.Itoa(template.ID)
			if tt.newTarget != "" {
				target = `"workflow_id":` + strconv.Itoa(workflows[tt.newTarget])
			}
			body := `{"id":` + strconv.Itoa(old.ID) + `,"name":"renamed","project_id":` + strconv.Itoa(f.project.ID) + `,` + target + `,"cron_format":"0 0 * * *","active":true}`
			r, rec := f.request(http.MethodPut, body, map[string]any{"schedule": old, "schedule_pool": pool}, nil)
			w := httptest.NewRecorder()

			scheduleController(tt.getFeatures).UpdateSchedule(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			stored, err := f.store.GetSchedule(f.project.ID, old.ID)
			require.NoError(t, err)
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), workflowSchedulesProOnly)
				assert.Equal(t, "nightly", stored.Name, "nothing is changed")
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

func TestRestore_WorkflowScheduleIsProOnly(t *testing.T) {
	f := newResourceFixture(t)
	controller := &BackupController{
		workflowStore: proFactory.NewWorkflowStore(f.store),
		pro:           proFeatureGate{subscription: noLicense{}, getFeatures: noFeatures},
	}
	backup := projectService.BackupFormat{
		Version:   projectService.BackupFormatVersion,
		Meta:      projectService.BackupMeta{Project: db.Project{Name: "restored"}},
		Workflows: []projectService.BackupWorkflow{{WorkflowTemplate: db.WorkflowTemplate{Name: "wf"}}},
		Schedules: []projectService.BackupSchedule{{Schedule: db.Schedule{Name: "nightly", CronFormat: "0 0 * * *"}, Workflow: new("wf")}},
	}
	body, err := backup.Marshal()
	require.NoError(t, err)
	projectsBefore, err := f.store.GetAllProjects()
	require.NoError(t, err)

	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	controller.Restore(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), workflowSchedulesProOnly)
	projectsAfter, err := f.store.GetAllProjects()
	require.NoError(t, err)
	assert.Len(t, projectsAfter, len(projectsBefore), "nothing is created")
}
