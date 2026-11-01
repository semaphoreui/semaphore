package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_Validate(t *testing.T) {
	templateID := 7
	workflowID := 9

	tests := []struct {
		name        string
		integration Integration
		wantErr     string
	}{
		{"template target", Integration{Name: "hook", GroupID: 1, TemplateID: &templateID}, ""},
		{"workflow target", Integration{Name: "hook", GroupID: 1, WorkflowID: &workflowID}, ""},
		{"no name", Integration{GroupID: 1, TemplateID: &templateID}, "No Name set for integration"},
		{"no group", Integration{Name: "hook", TemplateID: &templateID}, "Integration must belong to an integration group"},
		{"no target", Integration{Name: "hook", GroupID: 1}, "Integration must run a template or a workflow"},
		{"both targets", Integration{Name: "hook", GroupID: 1, TemplateID: &templateID, WorkflowID: &workflowID}, "Integration can run either a template or a workflow, not both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.integration.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIntegration_RunsWorkflow(t *testing.T) {
	id := 1
	assert.False(t, (&Integration{TemplateID: &id}).RunsWorkflow())
	assert.True(t, (&Integration{WorkflowID: &id}).RunsWorkflow())
}
