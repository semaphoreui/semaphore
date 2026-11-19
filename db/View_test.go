package db

import (
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
