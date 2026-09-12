package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRole(t *testing.T) {
	projectID := 1

	tests := []struct {
		name    string
		role    Role
		wantErr bool
	}{
		{
			name:    "valid custom role",
			role:    Role{Slug: "deployer", Name: "Deployer", ProjectID: &projectID},
			wantErr: false,
		},
		{
			name:    "empty name",
			role:    Role{Slug: "deployer", Name: "", ProjectID: &projectID},
			wantErr: true,
		},
		{
			name:    "empty slug",
			role:    Role{Slug: "", Name: "Deployer", ProjectID: &projectID},
			wantErr: true,
		},
		{
			name:    "reserved slug owner",
			role:    Role{Slug: string(ProjectOwner), Name: "pwn", ProjectID: &projectID},
			wantErr: true,
		},
		{
			name:    "reserved slug manager",
			role:    Role{Slug: string(ProjectManager), Name: "pwn", ProjectID: &projectID},
			wantErr: true,
		},
		{
			name:    "reserved slug task_runner",
			role:    Role{Slug: string(ProjectTaskRunner), Name: "pwn", ProjectID: &projectID},
			wantErr: true,
		},
		{
			name:    "reserved slug guest",
			role:    Role{Slug: string(ProjectGuest), Name: "pwn", ProjectID: &projectID},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRole_ReservedSlugsMatchBuiltins guards against a built-in role
// being added later without also reserving its slug in ValidateRole.
func TestValidateRole_ReservedSlugsMatchBuiltins(t *testing.T) {
	for role := range rolePermissions {
		err := ValidateRole(Role{Slug: string(role), Name: "custom"})
		assert.Error(t, err, "built-in slug %q must be rejected by ValidateRole", role)
	}
}
