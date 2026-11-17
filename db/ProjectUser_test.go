package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLiftPermissions(t *testing.T) {
	tests := []struct {
		name string
		mask ProjectUserPermission
		want ProjectUserPermission
	}{
		{"nothing", 0, 0},
		{"run tasks gives run and approve workflows", 1, 1 | CanRunWorkflows | CanApproveWorkflows},
		{"manage resources gives edit workflows", 4, 4 | CanEditWorkflows},
		{"manager", 5, 5 | CanRunWorkflows | CanEditWorkflows | CanApproveWorkflows},
		{"owner", 15, 15 | CanRunWorkflows | CanEditWorkflows | CanApproveWorkflows},
		{"already lifted", 1 | CanRunWorkflows | CanApproveWorkflows, 1 | CanRunWorkflows | CanApproveWorkflows},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, LiftPermissions(tt.mask))
		})
	}
}
