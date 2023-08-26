package db

import (
	"testing"
)

func TestProjectUsers_RoleCan(t *testing.T) {
	if !ProjectManager.Can(CanManageProjectResources) {
		t.Fatal()
	}

	if ProjectManager.Can(CanUpdateProject) {
		t.Fatal()
	}
}
