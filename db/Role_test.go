package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		isNew   bool
		wantErr bool
	}{
		{"valid new role", Role{Slug: "ci_bot-2", Name: "CI"}, true, false},
		{"slug keeps its case", Role{Slug: "Deployer", Name: "D"}, true, false},
		{"longest slug", Role{Slug: strings.Repeat("a", 100), Name: "A"}, true, false},
		{"empty name", Role{Slug: "deployer"}, true, true},
		{"empty slug", Role{Name: "D"}, true, true},
		{"space in slug", Role{Slug: "with space", Name: "D"}, true, true},
		{"non-ASCII slug", Role{Slug: "dé", Name: "D"}, true, true},
		{"slash in slug", Role{Slug: "a/b", Name: "D"}, true, true},
		{"slug too long", Role{Slug: strings.Repeat("a", 101), Name: "A"}, true, true},
		{"reserved slug", Role{Slug: string(ProjectOwner), Name: "D"}, true, true},
		{"reserved slug in other case", Role{Slug: "Task_Runner", Name: "D"}, true, true},
		{"reserved slug all", Role{Slug: "ALL", Name: "D"}, true, true},
		{"unknown permission bit", Role{Slug: "d", Name: "D", Permissions: 16}, true, true},
		{"negative permissions", Role{Slug: "d", Name: "D", Permissions: -1}, true, true},
		{"old slug stays editable", Role{Slug: "my role", Name: "D"}, false, false},
		{"update still needs a name", Role{Slug: "my role"}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role, tt.isNew)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
