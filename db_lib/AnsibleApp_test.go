package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestAnsibleApp_skipGalaxyInstall(t *testing.T) {
	tests := []struct {
		name     string
		tpl      *db.AnsibleTemplateParams
		params   *db.AnsibleTaskParams
		expected bool
	}{
		{
			name:     "no template params",
			tpl:      nil,
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: false,
		},
		{
			name:     "template skip enabled, override disabled",
			tpl:      &db.AnsibleTemplateParams{SkipGalaxyInstall: true},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: false},
			expected: true,
		},
		{
			name:     "template skip disabled, override disabled, task wants skip",
			tpl:      &db.AnsibleTemplateParams{SkipGalaxyInstall: false},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: false,
		},
		{
			name: "override enabled, task disables skip",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              true,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: false},
			expected: false,
		},
		{
			name: "override enabled, task enables skip",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              false,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: true,
		},
		{
			name: "override enabled, nil task params falls back to template",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              true,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &AnsibleApp{}

			args := LocalAppInstallingArgs{}
			if tt.tpl != nil {
				args.TplParams = tt.tpl
			}
			if tt.params != nil {
				args.Params = tt.params
			}

			assert.Equal(t, tt.expected, app.skipGalaxyInstall(args))
		})
	}
}

func TestAnsibleApp_galaxyInstallKeys(t *testing.T) {
	repoKey := db.AccessKey{ID: 1, Name: "repo", Type: db.AccessKeySSH}

	tests := []struct {
		name         string
		repoKey      *db.AccessKey
		templateKeys []db.AccessKey
		expected     []string
	}{
		{
			name:         "repository key only",
			templateKeys: nil,
			expected:     []string{"repo"},
		},
		{
			name: "repository key is tried before the template keys",
			templateKeys: []db.AccessKey{
				{ID: 2, Name: "role-a", Type: db.AccessKeySSH},
				{ID: 3, Name: "role-b", Type: db.AccessKeySSH},
			},
			expected: []string{"repo", "role-a", "role-b"},
		},
		{
			// A public repository, or one reached over https with a
			// login/password key, has no key an agent can serve. The install
			// must still run, with the environment of the task.
			name:         "a repository without an ssh key contributes none",
			repoKey:      &db.AccessKey{ID: 1, Name: "none", Type: db.AccessKeyNone},
			templateKeys: nil,
			expected:     nil,
		},
		{
			name:         "template keys are used when the repository has none",
			repoKey:      &db.AccessKey{ID: 1, Name: "login", Type: db.AccessKeyLoginPassword},
			templateKeys: []db.AccessKey{{ID: 2, Name: "role-a", Type: db.AccessKeySSH}},
			expected:     []string{"role-a"},
		},
		{
			name: "keys which cannot be used by an SSH agent are skipped",
			templateKeys: []db.AccessKey{
				{ID: 2, Name: "login", Type: db.AccessKeyLoginPassword},
				{ID: 3, Name: "none", Type: db.AccessKeyNone},
				{ID: 4, Name: "role-a", Type: db.AccessKeySSH},
			},
			expected: []string{"repo", "role-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := repoKey
			if tt.repoKey != nil {
				key = *tt.repoKey
			}

			app := &AnsibleApp{
				Repository: db.Repository{SSHKey: key},
				Template:   db.Template{Keys: tt.templateKeys},
			}

			var names []string
			for _, key := range app.galaxyInstallKeys() {
				names = append(names, key.Name)
			}

			assert.Equal(t, tt.expected, names)
		})
	}
}

// TestGalaxyInstallArgs covers when ansible-galaxy is told to re-fetch roles.
//
// The keys of a template are tried one at a time, and each attempt installs
// what its key can reach. ansible-galaxy --force removes a role before fetching
// it again, so forcing on a retry deletes the roles the previous key installed
// when the current key cannot reach the same repository. Only the first attempt
// may force.
func TestGalaxyInstallArgs(t *testing.T) {
	const requirements = "/repo/requirements.yml"

	t.Run("the first attempt forces a fresh install", func(t *testing.T) {
		args := galaxyInstallArgs(GalaxyRole, requirements, true)

		assert.Equal(t,
			[]string{"role", "install", "-r", requirements, "--force"},
			args)
	})

	t.Run("a retry must not force", func(t *testing.T) {
		args := galaxyInstallArgs(GalaxyRole, requirements, false)

		assert.Equal(t, []string{"role", "install", "-r", requirements}, args)
		assert.NotContains(t, args, "--force",
			"forcing on a retry removes the roles installed by the previous key")
	})

	t.Run("collections use the same rule", func(t *testing.T) {
		assert.NotContains(t, galaxyInstallArgs(GalaxyCollection, requirements, false), "--force")
		assert.Contains(t, galaxyInstallArgs(GalaxyCollection, requirements, true), "--force")
	})
}
