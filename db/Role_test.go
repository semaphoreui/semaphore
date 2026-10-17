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
		{"global role with global permissions", Role{Slug: "ops", Name: "Ops", GlobalPermissions: AllGlobalPermissions}, true, false},
		{"unknown global permission bit", Role{Slug: "ops", Name: "Ops", GlobalPermissions: 16}, true, true},
		{"negative global permissions", Role{Slug: "ops", Name: "Ops", GlobalPermissions: -1}, true, true},
		{"project role cannot hold global permissions", Role{Slug: "ops", Name: "Ops", ProjectID: new(int), GlobalPermissions: CanReadGlobalAudit}, true, true},
		{"project role update cannot gain global permissions", Role{Slug: "ops", Name: "Ops", ProjectID: new(int), GlobalPermissions: CanReadGlobalAudit}, false, true},
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

func TestRoleAssignments_Impact(t *testing.T) {
	assigned := RoleAssignments{
		Members: []RoleMember{{ProjectID: 1}, {ProjectID: 1}, {ProjectID: 2}},
		Grants:  []TemplateRolePerm{{ProjectID: 3}},
	}

	assert.Equal(t, RoleImpact{Members: 3, Projects: 3, TemplateGrants: 1}, assigned.Impact())
	assert.Equal(t, RoleImpact{}, RoleAssignments{}.Impact())

	assigned.GlobalMembers = []GlobalRoleAssignment{{UserID: 1}, {UserID: 2}}
	assert.Equal(t, RoleImpact{Members: 5, Projects: 3, TemplateGrants: 1}, assigned.Impact(), "global members count as members of no project")
}

func TestGlobalPermission_Can(t *testing.T) {
	mask := CanManageGlobalUsers | CanReadGlobalAudit
	assert.True(t, mask.Can(CanManageGlobalUsers))
	assert.True(t, mask.Can(CanManageGlobalUsers|CanReadGlobalAudit))
	assert.False(t, mask.Can(CanManageGlobalRoles))
	assert.False(t, mask.Can(CanManageGlobalUsers|CanManageGlobalRoles))
	assert.True(t, AllGlobalPermissions.Can(CanManageGlobalSystem))
	assert.Equal(t, GlobalPermission(15), AllGlobalPermissions, "the bit values are part of the API")
}

func TestGlobalPermissionCatalog_CoversEveryBit(t *testing.T) {
	var covered GlobalPermission
	ids := map[string]bool{}
	for _, definition := range GlobalPermissionCatalog() {
		covered |= definition.Permission
		assert.False(t, ids[definition.ID], "duplicate id %s", definition.ID)
		ids[definition.ID] = true
		assert.NotEmpty(t, definition.Description)
	}
	assert.Equal(t, AllGlobalPermissions, covered)
}

func TestExplainEffectiveGlobalPermissions(t *testing.T) {
	assignments := []GlobalRoleAssignment{
		{RoleID: 3, RoleName: "Auditor", GlobalPermissions: CanReadGlobalAudit},
		{RoleID: 4, RoleName: "Ops", GlobalPermissions: CanReadGlobalAudit | CanManageGlobalSystem},
	}

	explained := ExplainEffectiveGlobalPermissions(false, assignments)
	assert.Equal(t, CanReadGlobalAudit|CanManageGlobalSystem, explained.Permissions)
	byID := map[string]GlobalPermissionDecision{}
	for _, decision := range explained.Decisions {
		byID[decision.Permission] = decision
	}
	assert.Len(t, byID, len(GlobalPermissionCatalog()))
	assert.Equal(t, GlobalPermissionDecision{Permission: "global.audit.read", Allowed: true, RoleID: 3, RoleName: "Auditor"}, byID["global.audit.read"], "the first role that grants it")
	assert.Equal(t, GlobalPermissionDecision{Permission: "global.system.manage", Allowed: true, RoleID: 4, RoleName: "Ops"}, byID["global.system.manage"])
	assert.Equal(t, GlobalPermissionDecision{Permission: "global.users.manage"}, byID["global.users.manage"])

	admin := ExplainEffectiveGlobalPermissions(true, nil)
	assert.Equal(t, AllGlobalPermissions, admin.Permissions)
	for _, decision := range admin.Decisions {
		assert.True(t, decision.Allowed)
		assert.Equal(t, "Administrator", decision.RoleName)
	}

	nobody := ExplainEffectiveGlobalPermissions(false, nil)
	assert.Equal(t, GlobalPermission(0), nobody.Permissions)
	assert.Len(t, nobody.Decisions, len(GlobalPermissionCatalog()))
}

func TestErrLastGlobalAdministrator_IsAConflict(t *testing.T) {
	assert.ErrorIs(t, ErrLastGlobalAdministrator, ErrInvalidOperation)
	assert.ErrorIs(t, ErrGlobalRoleAssigned, ErrInvalidOperation)
	assert.NotErrorIs(t, ErrLastGlobalAdministrator, ErrGlobalRoleAssigned)
	assert.NotErrorIs(t, ErrInvalidOperation, ErrLastGlobalAdministrator)
}
