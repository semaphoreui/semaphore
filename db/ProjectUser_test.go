package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectUsers_RoleCan(t *testing.T) {
	assert.True(t, ProjectManager.Can(CanManageProjectResources))
	assert.False(t, ProjectManager.Can(CanUpdateProject))
}

func TestProjectUserRole_IsBuiltin(t *testing.T) {
	for _, role := range []ProjectUserRole{
		ProjectOwner,
		ProjectManager,
		ProjectTaskRunner,
		ProjectGuest,
	} {
		assert.True(t, role.IsBuiltin(), "role %q should be built-in", role)
	}

	assert.False(t, ProjectUserRole("deployer").IsBuiltin())
	assert.False(t, ProjectNone.IsBuiltin())
}
