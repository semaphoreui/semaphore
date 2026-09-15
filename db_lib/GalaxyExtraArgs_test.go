package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			extra, err := galaxyExtraArgs(args, tt.reqType)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, extra)
		})
	}
}

// Role and collection args must never leak into each other's command:
// `ansible-galaxy role install --pre` is rejected by ansible.
func TestGalaxyExtraArgs_KeptSeparate(t *testing.T) {
	args := LocalAppInstallingArgs{TplParams: &db.AnsibleTemplateParams{
		GalaxyCollectionArgs: []string{"--pre"},
	}}

	roleArgs, err := galaxyExtraArgs(args, GalaxyRole)
	require.NoError(t, err)
	assert.Empty(t, roleArgs)

	collectionArgs, err := galaxyExtraArgs(args, GalaxyCollection)
	require.NoError(t, err)
	assert.Equal(t, []string{"--pre"}, collectionArgs)
}

func TestGalaxyExtraArgs_RejectsDisallowed(t *testing.T) {
	args := LocalAppInstallingArgs{TplParams: &db.AnsibleTemplateParams{
		GalaxyRoleArgs:       []string{"--pre"},
		GalaxyCollectionArgs: []string{"--token", "secret"},
	}}

	_, err := galaxyExtraArgs(args, GalaxyRole)
	assert.ErrorContains(t, err, `"--pre" is not allowed`)

	_, err = galaxyExtraArgs(args, GalaxyCollection)
	assert.ErrorContains(t, err, `"--token" is not allowed`)
}
