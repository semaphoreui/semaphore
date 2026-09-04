package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestGalaxyExtraArgs(t *testing.T) {
	params := &db.AnsibleTemplateParams{
		GalaxyRoleArgs:       []string{"--ignore-errors"},
		GalaxyCollectionArgs: []string{"--pre", "--no-deps"},
	}

	tests := []struct {
		name      string
		tplParams any
		reqType   GalaxyRequirementsType
		expected  []string
	}{
		{"roles", params, GalaxyRole, []string{"--ignore-errors"}},
		{"collections", params, GalaxyCollection, []string{"--pre", "--no-deps"}},
		{"no params", nil, GalaxyRole, nil},
		{"wrong params type", &db.TerraformTemplateParams{}, GalaxyRole, nil},
		{"nil ansible params", (*db.AnsibleTemplateParams)(nil), GalaxyCollection, nil},
		{"unset", &db.AnsibleTemplateParams{}, GalaxyRole, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := LocalAppInstallingArgs{TplParams: tt.tplParams}
			assert.Equal(t, tt.expected, galaxyExtraArgs(args, tt.reqType))
		})
	}
}

// Role and collection args must never leak into each other's command:
// `ansible-galaxy role install --pre` is rejected by ansible.
func TestGalaxyExtraArgs_KeptSeparate(t *testing.T) {
	args := LocalAppInstallingArgs{TplParams: &db.AnsibleTemplateParams{
		GalaxyCollectionArgs: []string{"--pre"},
	}}

	assert.Empty(t, galaxyExtraArgs(args, GalaxyRole))
	assert.Equal(t, []string{"--pre"}, galaxyExtraArgs(args, GalaxyCollection))
}
