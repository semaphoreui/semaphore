package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectUsers_RoleCan(t *testing.T) {
	assert.True(t, ProjectManager.Can(CanManageProjectResources))
	assert.False(t, ProjectManager.Can(CanUpdateProject))
}
