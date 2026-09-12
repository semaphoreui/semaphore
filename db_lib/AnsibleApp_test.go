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
