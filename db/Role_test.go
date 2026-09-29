package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCustomRole(t *testing.T) {
	projectID := 1
	builtinKey := BuiltinRoleOwner

	tests := []struct {
		name    string
		role    Role
		wantErr bool
	}{
		{
			name: "valid custom role",
			role: Role{
				Name:      "Deployer",
				ProjectID: &projectID,
			},
		},
		{
			name: "empty name",
			role: Role{
				ProjectID: &projectID,
			},
			wantErr: true,
		},
		{
			name: "custom role with built-in key",
			role: Role{
				Name:       "Deployer",
				ProjectID:  &projectID,
				BuiltinKey: &builtinKey,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCustomRole(tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
