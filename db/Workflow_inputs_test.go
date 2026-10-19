package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidWorkflowOutputName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"image_tag", true},
		{"_private", true},
		{"env-name", true},
		{"A1", true},
		{"", false},
		{"1st", false},
		{"-x", false},
		{"bad name", false},
		{"with.dot", false},
		{"юникод", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidWorkflowOutputName(tt.name))
		})
	}
}

func TestWorkflowEdge_EffectiveInputMode(t *testing.T) {
	assert.Equal(t, WorkflowEdgeInputByName, WorkflowEdge{}.EffectiveInputMode())
	assert.Equal(t, WorkflowEdgeInputExplicit, WorkflowEdge{InputMode: WorkflowEdgeInputExplicit}.EffectiveInputMode())
	assert.NoError(t, WorkflowEdgeInputByName.Validate())
	assert.NoError(t, WorkflowEdgeInputExplicit.Validate())
	assert.Error(t, WorkflowEdgeInputMode("implicit").Validate())
}
