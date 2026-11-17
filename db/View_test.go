package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestView_ValidateFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  *MapStringAnyField
		wantErr string
	}{
		{"no filter", nil, ""},
		{"name filter", &MapStringAnyField{"name": "prod"}, ""},
		{"empty name", &MapStringAnyField{"name": ""}, ""},
		{"unknown key would silently do nothing", &MapStringAnyField{"app": "ansible"}, "unknown view filter: app"},
		{"name must be text", &MapStringAnyField{"name": 42}, "view filter name must be a string"},
		{"fits the column", &MapStringAnyField{"name": strings.Repeat("a", 980)}, ""},
		{"longer than the column", &MapStringAnyField{"name": strings.Repeat("a", 1000)}, "view filter is too long"},
		// json.Marshal writes < as <, six bytes, so the name alone undercounts.
		{"escaped characters count as encoded", &MapStringAnyField{"name": strings.Repeat("<", 200)}, "view filter is too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := View{Title: "v", Filter: tt.filter}
			err := view.Validate()

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestView_NameFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *MapStringAnyField
		expected string
	}{
		{"no filter", nil, ""},
		{"set", &MapStringAnyField{"name": "prod"}, "prod"},
		{"trimmed", &MapStringAnyField{"name": "  prod  "}, "prod"},
		{"not a string", &MapStringAnyField{"name": 42}, ""},
		{"other key only", &MapStringAnyField{"app": "x"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := View{Filter: tt.filter}
			assert.Equal(t, tt.expected, view.NameFilter())
		})
	}
}
