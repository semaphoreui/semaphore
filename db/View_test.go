package db

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
)

func TestViewsWithTemplateStats(t *testing.T) {
	status := func(s task_logger.TaskStatus) *task_logger.TaskStatus { return &s }
	view := func(id int) *int { return &id }

	views := []View{
		{ID: 1, Title: "All", Type: ViewTypeAll, ShowTemplateStats: true},
		{ID: 2, Title: "Build", Type: ViewTypeCustom, ShowTemplateStats: true},
		{ID: 3, Title: "Empty", Type: ViewTypeCustom, ShowTemplateStats: true},
		{ID: 4, Title: "Quiet", Type: ViewTypeCustom},
		{ID: 5, Title: "Quiet all", Type: ViewTypeAll},
	}

	counts := []ViewTemplateCount{
		{ViewID: view(2), Status: status(task_logger.TaskSuccessStatus), Templates: 3},
		{ViewID: view(2), Status: status(task_logger.TaskFailStatus), Templates: 2},
		{ViewID: view(2), Status: status(task_logger.TaskStoppedStatus), Templates: 1},
		{ViewID: view(2), Status: status(task_logger.TaskStoppingStatus), Templates: 1},
		{ViewID: view(2), Status: nil, Templates: 4},
		{ViewID: view(2), Status: status(task_logger.TaskWaitingConfirmation), Templates: 5},
		// queued or executed: all running
		{ViewID: view(2), Status: status(task_logger.TaskWaitingStatus), Templates: 1},
		{ViewID: view(2), Status: status(task_logger.TaskStartingStatus), Templates: 2},
		{ViewID: view(2), Status: status(task_logger.TaskRunningStatus), Templates: 3},
		{ViewID: view(2), Status: status(task_logger.TaskConfirmed), Templates: 4},
		{ViewID: view(2), Status: status(task_logger.TaskRejected), Templates: 5},
		// a status this code does not know is in no count
		{ViewID: view(2), Status: status("unknown"), Templates: 100},
		// assigned to no view and to a view that is not listed: only "All" has them
		{ViewID: nil, Status: status(task_logger.TaskSuccessStatus), Templates: 7},
		{ViewID: view(99), Status: status(task_logger.TaskFailStatus), Templates: 1},
		// counted for another reason: a view that does not show the counts stays without them
		{ViewID: view(4), Status: status(task_logger.TaskSuccessStatus), Templates: 1},
	}

	res := ViewsWithTemplateStats(views, counts)

	assert.Equal(t, []ViewWithTemplateStats{
		{View: views[0], TemplateStats: &ViewTemplateStats{Success: 11, Failed: 3, Stopped: 6, Confirming: 5, Running: 15}},
		{View: views[1], TemplateStats: &ViewTemplateStats{Success: 3, Failed: 2, Stopped: 6, Confirming: 5, Running: 15}},
		{View: views[2], TemplateStats: &ViewTemplateStats{}},
		{View: views[3]},
		{View: views[4]},
	}, res)
}

func TestViewTemplateCountScopeOf(t *testing.T) {
	tests := []struct {
		name     string
		views    []View
		expected ViewTemplateCountScope
	}{
		{"no views", nil, ViewTemplateCountScope{}},
		{"no view shows the counts", []View{{ID: 1, Type: ViewTypeAll}, {ID: 2}}, ViewTemplateCountScope{}},
		{"custom views that show them", []View{
			{ID: 1, Type: ViewTypeAll},
			{ID: 2, ShowTemplateStats: true},
			{ID: 3},
			{ID: 4, ShowTemplateStats: true},
		}, ViewTemplateCountScope{ViewIDs: []int{2, 4}}},
		{"an All view needs every template", []View{
			{ID: 1, Type: ViewTypeAll, ShowTemplateStats: true},
			{ID: 2, ShowTemplateStats: true},
		}, ViewTemplateCountScope{All: true, ViewIDs: []int{2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := ViewTemplateCountScopeOf(tt.views)
			assert.Equal(t, tt.expected, scope)
			assert.Equal(t, !tt.expected.All && len(tt.expected.ViewIDs) == 0, scope.Empty())
		})
	}
}

func TestViewsWithTemplateStats_NoViews(t *testing.T) {
	res := ViewsWithTemplateStats(nil, []ViewTemplateCount{{Templates: 1}})

	assert.NotNil(t, res, "an empty list must be encoded as [], not null")
	assert.Empty(t, res)
}

func TestView_ValidateFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  *MapStringAnyField
		wantErr string
	}{
		{"no filter", nil, ""},
		{"name filter", &MapStringAnyField{"name": "prod"}, ""},
		{"empty name", &MapStringAnyField{"name": ""}, ""},
		{"unknown key would silently do nothing", &MapStringAnyField{"app": "ansible"}, "unknown view filter: app"},
		{"name must be text", &MapStringAnyField{"name": 42}, "view filter name must be a string"},
		{"fits the column", &MapStringAnyField{"name": strings.Repeat("a", 980)}, ""},
		{"longer than the column", &MapStringAnyField{"name": strings.Repeat("a", 1000)}, "view filter is too long"},
		// json.Marshal writes < as <, six bytes, so the name alone undercounts.
		{"escaped characters count as encoded", &MapStringAnyField{"name": strings.Repeat("<", 200)}, "view filter is too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := View{Title: "v", Filter: tt.filter}
			err := view.Validate()

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestView_NameFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *MapStringAnyField
		expected string
	}{
		{"no filter", nil, ""},
		{"set", &MapStringAnyField{"name": "prod"}, "prod"},
		{"trimmed", &MapStringAnyField{"name": "  prod  "}, "prod"},
		{"not a string", &MapStringAnyField{"name": 42}, ""},
		{"other key only", &MapStringAnyField{"app": "x"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := View{Filter: tt.filter}
			assert.Equal(t, tt.expected, view.NameFilter())
		})
	}
}
