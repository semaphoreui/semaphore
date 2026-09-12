package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
)

func TestShouldSkipStatusAlert(t *testing.T) {
	tests := []struct {
		name                  string
		status                task_logger.TaskStatus
		suppressSuccessAlerts bool
		suppressErrorAlerts   bool
		expected              bool
	}{
		{
			name:                  "success with suppress success",
			status:                task_logger.TaskSuccessStatus,
			suppressSuccessAlerts: true,
			expected:              true,
		},
		{
			name:                "fail with suppress error",
			status:              task_logger.TaskFailStatus,
			suppressErrorAlerts: true,
			expected:            true,
		},
		{
			name:     "fail without flags",
			status:   task_logger.TaskFailStatus,
			expected: false,
		},
		{
			name:     "success without flags",
			status:   task_logger.TaskSuccessStatus,
			expected: false,
		},
		{
			name:                  "fail with only success flag",
			status:                task_logger.TaskFailStatus,
			suppressSuccessAlerts: true,
			expected:              false,
		},
		{
			name:                "success with only error flag",
			status:              task_logger.TaskSuccessStatus,
			suppressErrorAlerts: true,
			expected:            false,
		},
		{
			name:                  "fail with both flags",
			status:                task_logger.TaskFailStatus,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              true,
		},
		{
			name:                  "success with both flags",
			status:                task_logger.TaskSuccessStatus,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              true,
		},
		{
			name:                  "waiting confirmation with both flags",
			status:                task_logger.TaskWaitingConfirmation,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &TaskRunner{
				Task: db.Task{
					Status: tt.status,
				},
				Template: db.Template{
					SuppressSuccessAlerts: tt.suppressSuccessAlerts,
					SuppressErrorAlerts:   tt.suppressErrorAlerts,
				},
			}

			assert.Equal(t, tt.expected, runner.shouldSkipStatusAlert())
		})
	}
}
