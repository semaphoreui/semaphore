package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildManagedInventory(t *testing.T) {
	hosts := []db.InventoryHost{
		{Name: "db1"},
		{Name: "web1", Variables: "ansible_host: 10.0.0.5"},
	}

	text, err := buildManagedInventory("http_port: 8080", hosts)

	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, map[string]any{
		"all": map[string]any{
			"vars": map[string]any{"http_port": 8080},
			"hosts": map[string]any{
				"db1":  map[string]any{},
				"web1": map[string]any{"ansible_host": "10.0.0.5"},
			},
		},
	}, parsed)
}

func TestBuildManagedInventory_KeepsNamesAsStrings(t *testing.T) {
	hosts := []db.InventoryHost{{Name: "123"}, {Name: "null"}, {Name: "yes"}, {Name: "~"}}

	text, err := buildManagedInventory("", hosts)

	require.NoError(t, err)
	for _, key := range []string{`"123":`, `"null":`, `"yes":`, `"~":`} {
		assert.Contains(t, text, key)
	}
}

func TestBuildManagedInventory_KeepsValuesAsWritten(t *testing.T) {
	hosts := []db.InventoryHost{{
		Name:      "web1",
		Variables: "mode: 0755\nd: 2024-01-01\non_off: on\nquoted: \"yes\"",
	}}

	text, err := buildManagedInventory("flag: no", hosts)

	require.NoError(t, err)
	for _, line := range []string{"flag: no", "mode: 0755", "d: 2024-01-01", "on_off: on", `quoted: "yes"`} {
		assert.Contains(t, text, line)
	}
}

func TestBuildManagedInventory_JSONVariables(t *testing.T) {
	text, err := buildManagedInventory(`{"a": 1, "b": "x"}`, nil)

	require.NoError(t, err)
	var parsed map[string]map[string]map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, map[string]any{"a": 1, "b": "x"}, parsed["all"]["vars"])
}

func TestBuildManagedInventory_NoHosts(t *testing.T) {
	text, err := buildManagedInventory("", nil)

	require.NoError(t, err)
	var parsed map[string]map[string]map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	require.Contains(t, parsed, "all")
	require.Contains(t, parsed["all"], "hosts")
	assert.Empty(t, parsed["all"]["hosts"])
}

func TestBuildManagedInventory_HostErrorNamesHostOnly(t *testing.T) {
	hosts := []db.InventoryHost{{Name: "web1", Variables: "secret: [oops"}}

	_, err := buildManagedInventory("", hosts)

	require.Error(t, err)
	assert.ErrorContains(t, err, "host web1: ")
	assert.NotContains(t, err.Error(), "secret")
}

func TestBuildManagedInventory_RejectsAnchors(t *testing.T) {
	hosts := []db.InventoryHost{{Name: "web1", Variables: "base: &x {u: 1}\ncopy: *x"}}

	_, err := buildManagedInventory("", hosts)

	assert.ErrorContains(t, err, "host web1")
}
