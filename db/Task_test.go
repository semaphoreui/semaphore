package db

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_GetWorkflowUrl(t *testing.T) {
	runID := 42
	workflowID := 7

	tests := []struct {
		name     string
		webHost  string
		task     Task
		expected *string
	}{
		{
			name:     "workflow task",
			webHost:  "https://semaphore.example.com",
			task:     Task{ID: 1, ProjectID: 3, WorkflowRunID: &runID, WorkflowTemplateID: &workflowID},
			expected: new("https://semaphore.example.com/project/3/workflows/7/runs/42"),
		},
		{
			name:     "no web host",
			webHost:  "",
			task:     Task{ID: 1, ProjectID: 3, WorkflowRunID: &runID, WorkflowTemplateID: &workflowID},
			expected: nil,
		},
		{
			name:     "not a workflow task",
			webHost:  "https://semaphore.example.com",
			task:     Task{ID: 1, ProjectID: 3},
			expected: nil,
		},
		{
			name:     "workflow template not resolved",
			webHost:  "https://semaphore.example.com",
			task:     Task{ID: 1, ProjectID: 3, WorkflowRunID: &runID},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousConfig := util.Config
			t.Cleanup(func() { util.Config = previousConfig })
			util.Config = &util.ConfigType{WebHost: tt.webHost}

			actual := tt.task.GetWorkflowUrl()

			if tt.expected == nil {
				assert.Nil(t, actual)
				return
			}
			require.NotNil(t, actual)
			assert.Equal(t, *tt.expected, *actual)
		})
	}
}
