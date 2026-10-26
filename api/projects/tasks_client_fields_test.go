package projects

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

// A client may not bind its task to a workflow run, pre-fill outputs or pick
// a runner: those fields are set by the engine, the task's own run and the
// pool. Everything else of the request survives.
func TestSanitizeClientTask(t *testing.T) {
	runID, nodeID, runnerID := 7, 3, 2
	outputs := `{"values":{"image_tag":"evil"}}`
	branch := "main"

	got := sanitizeClientTask(db.Task{
		TemplateID:     5,
		Environment:    `{"image_tag":"1.0"}`,
		GitBranch:      &branch,
		Message:        "hello",
		WorkflowRunID:  &runID,
		WorkflowNodeID: &nodeID,
		Artifacts:      &outputs,
		RunnerID:       &runnerID,
	})

	assert.Nil(t, got.WorkflowRunID)
	assert.Nil(t, got.WorkflowNodeID)
	assert.Nil(t, got.Artifacts)
	assert.Nil(t, got.RunnerID)
	assert.Equal(t, 5, got.TemplateID)
	assert.Equal(t, `{"image_tag":"1.0"}`, got.Environment)
	assert.Equal(t, &branch, got.GitBranch)
	assert.Equal(t, "hello", got.Message)
}
