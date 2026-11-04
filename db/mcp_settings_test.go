package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOptions map[string]string

func (o fakeOptions) GetOption(key string) (string, error) { return o[key], nil }
func (o fakeOptions) SetOption(key string, value string) error {
	o[key] = value
	return nil
}

type failingOptions struct{}

func (failingOptions) GetOption(string) (string, error) { return "", errors.New("database is down") }

func TestReadMCPSettings(t *testing.T) {
	tests := []struct {
		name     string
		options  fakeOptions
		expected MCPSettings
	}{
		{"no rows: all off", fakeOptions{}, MCPSettings{}},
		{"both stored", fakeOptions{MCPAllowRunOption: "false", MCPAllowDestroyOption: "true"}, MCPSettings{AllowRun: false, AllowDestroy: true}},
		{"both on", fakeOptions{MCPAllowRunOption: "true", MCPAllowDestroyOption: "true"}, MCPSettings{AllowRun: true, AllowDestroy: true}},
		{"one stored, the other default", fakeOptions{MCPAllowDestroyOption: "true"}, MCPSettings{AllowRun: false, AllowDestroy: true}},
		{"not a boolean: the default", fakeOptions{MCPAllowRunOption: "yes", MCPAllowDestroyOption: "on"}, MCPSettings{AllowRun: false, AllowDestroy: false}},
		{"admin set stored", fakeOptions{MCPAdminAllowWriteOption: "true", MCPAdminAllowDestroyOption: "false"}, MCPSettings{AdminAllowWrite: true}},
		{"admin destroy alone", fakeOptions{MCPAdminAllowDestroyOption: "true"}, MCPSettings{AdminAllowDestroy: true}},
		{"all four on", fakeOptions{MCPAllowRunOption: "true", MCPAllowDestroyOption: "true", MCPAdminAllowWriteOption: "true", MCPAdminAllowDestroyOption: "true"},
			MCPSettings{AllowRun: true, AllowDestroy: true, AdminAllowWrite: true, AdminAllowDestroy: true}},
		{"empty value: the default", fakeOptions{MCPAllowRunOption: "", MCPAllowDestroyOption: ""}, MCPSettings{AllowRun: false, AllowDestroy: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := ReadMCPSettings(tt.options)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, settings)
		})
	}
}

func TestReadMCPSettings_StoreError(t *testing.T) {
	_, err := ReadMCPSettings(failingOptions{})
	assert.Error(t, err, "a failed read is not taken for the defaults")
}

func TestWriteMCPSettings_StoresBooleans(t *testing.T) {
	options := fakeOptions{}
	want := MCPSettings{AllowRun: false, AllowDestroy: true, AdminAllowWrite: true, AdminAllowDestroy: false}
	require.NoError(t, WriteMCPSettings(options, want))
	assert.Equal(t, fakeOptions{MCPAllowRunOption: "false", MCPAllowDestroyOption: "true",
		MCPAdminAllowWriteOption: "true", MCPAdminAllowDestroyOption: "false"}, options)

	settings, err := ReadMCPSettings(options)
	require.NoError(t, err)
	assert.Equal(t, want, settings)
}
