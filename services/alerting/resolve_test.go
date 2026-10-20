package alerting

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	alertingProject := db.Project{ID: 1, Alert: true}
	quietProject := db.Project{ID: 1, Alert: false}

	tests := []struct {
		name     string
		project  db.Project
		template db.Template
		schedule *db.Schedule
		defaults []int
		expected db.AlertSnapshot
	}{
		{
			name:     "legacy template keeps server channels and project defaults",
			project:  alertingProject,
			template: db.Template{},
			defaults: []int{5, 6},
			expected: db.AlertSnapshot{Instance: true, AlertIDs: []int{5, 6}, OnSuccess: true, OnError: true},
		},
		{
			name:     "project without alert flag gets no server channels",
			project:  quietProject,
			template: db.Template{AlertMode: db.AlertModeDefault},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: false, AlertIDs: []int{5}, OnSuccess: true, OnError: true},
		},
		{
			name:     "suppress flags are carried",
			project:  alertingProject,
			template: db.Template{SuppressSuccessAlerts: true, SuppressErrorAlerts: true},
			defaults: []int{},
			expected: db.AlertSnapshot{Instance: true, AlertIDs: []int{}, OnSuccess: false, OnError: false},
		},
		{
			name:     "ids mode ignores server channels and defaults",
			project:  alertingProject,
			template: db.Template{AlertMode: db.AlertModeIDs, AlertIDs: []int{9, 9, 3}},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: false, AlertIDs: []int{9, 3}, OnSuccess: true, OnError: true},
		},
		{
			name:     "ids mode with empty list is silent",
			project:  alertingProject,
			template: db.Template{AlertMode: db.AlertModeIDs},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: false, AlertIDs: []int{}, OnSuccess: true, OnError: true},
		},
		{
			name:     "alert ids without a mode imply ids mode",
			project:  alertingProject,
			template: db.Template{AlertIDs: []int{4}},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: false, AlertIDs: []int{4}, OnSuccess: true, OnError: true},
		},
		{
			name:     "inherit schedule keeps the template choice",
			project:  alertingProject,
			template: db.Template{},
			schedule: &db.Schedule{AlertMode: db.AlertModeInherit, AlertIDs: []int{8}},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: true, AlertIDs: []int{5}, OnSuccess: true, OnError: true},
		},
		{
			name:     "ids schedule replaces everything",
			project:  alertingProject,
			template: db.Template{},
			schedule: &db.Schedule{AlertMode: db.AlertModeIDs, AlertIDs: []int{8}},
			defaults: []int{5},
			expected: db.AlertSnapshot{Instance: false, AlertIDs: []int{8}, OnSuccess: true, OnError: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Resolve(tt.project, tt.template, tt.schedule, tt.defaults))
		})
	}
}
