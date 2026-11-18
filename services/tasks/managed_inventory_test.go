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

	text, err := buildManagedInventory("http_port: 8080", hosts, nil, nil)

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

	text, err := buildManagedInventory("", hosts, nil, nil)

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

	text, err := buildManagedInventory("flag: no", hosts, nil, nil)

	require.NoError(t, err)
	for _, line := range []string{"flag: no", "mode: 0755", "d: 2024-01-01", "on_off: on", `quoted: "yes"`} {
		assert.Contains(t, text, line)
	}
}

func TestBuildManagedInventory_JSONVariables(t *testing.T) {
	text, err := buildManagedInventory(`{"a": 1, "b": "x"}`, nil, nil, nil)

	require.NoError(t, err)
	var parsed map[string]map[string]map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, map[string]any{"a": 1, "b": "x"}, parsed["all"]["vars"])
}

func TestBuildManagedInventory_NoHosts(t *testing.T) {
	text, err := buildManagedInventory("", nil, nil, nil)

	require.NoError(t, err)
	var parsed map[string]map[string]map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	require.Contains(t, parsed, "all")
	require.Contains(t, parsed["all"], "hosts")
	assert.Empty(t, parsed["all"]["hosts"])
}

func TestBuildManagedInventory_HostErrorNamesHostOnly(t *testing.T) {
	hosts := []db.InventoryHost{{Name: "web1", Variables: "secret: [oops"}}

	_, err := buildManagedInventory("", hosts, nil, nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "host web1: ")
	assert.NotContains(t, err.Error(), "secret")
}

func TestBuildManagedInventory_RejectsAnchors(t *testing.T) {
	hosts := []db.InventoryHost{{Name: "web1", Variables: "base: &x {u: 1}\ncopy: *x"}}

	_, err := buildManagedInventory("", hosts, nil, nil)

	assert.ErrorContains(t, err, "host web1")
}

func TestBuildManagedInventory_Groups(t *testing.T) {
	hosts := []db.InventoryHost{
		{ID: 1, Name: "web-01", Variables: "http_port: 8081"},
		{ID: 2, Name: "db-01"},
	}
	groups := []db.InventoryGroup{
		{ID: 10, Name: "web", Variables: "http_port: 8080", Children: []int{11}},
		{ID: 11, Name: "web-eu"},
	}
	members := []db.InventoryGroupHost{{GroupID: 11, HostID: 1}}

	text, err := buildManagedInventory("http_port: 80", hosts, groups, members)

	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, map[string]any{
		"all": map[string]any{
			"vars": map[string]any{"http_port": 80},
			"hosts": map[string]any{
				"web-01": map[string]any{"http_port": 8081},
				"db-01":  map[string]any{},
			},
			"children": map[string]any{
				"web": map[string]any{
					"vars":     map[string]any{"http_port": 8080},
					"children": map[string]any{"web-eu": map[string]any{}},
				},
				"web-eu": map[string]any{
					"hosts": map[string]any{"web-01": map[string]any{}},
				},
			},
		},
	}, parsed)
}

func TestBuildManagedInventory_GroupKeepsOnlyEnabledHosts(t *testing.T) {
	groups := []db.InventoryGroup{{ID: 10, Name: "web"}}
	members := []db.InventoryGroupHost{{GroupID: 10, HostID: 1}}

	text, err := buildManagedInventory("", nil, groups, members)

	require.NoError(t, err)
	var parsed struct {
		All struct{ Children map[string]any }
	}
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, map[string]any{}, parsed.All.Children["web"])
}

func TestBuildManagedInventory_KeepsGroupNamesAsStrings(t *testing.T) {
	groups := []db.InventoryGroup{{ID: 1, Name: "true"}, {ID: 2, Name: "123"}, {ID: 3, Name: "null"}, {ID: 4, Name: "web-eu"}}

	text, err := buildManagedInventory("", nil, groups, nil)

	require.NoError(t, err)
	var parsed struct {
		All struct{ Children map[any]any }
	}
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	for _, name := range []string{"true", "123", "null", "web-eu"} {
		assert.Contains(t, parsed.All.Children, name)
	}
}

func TestBuildManagedInventory_GroupCycle(t *testing.T) {
	groups := []db.InventoryGroup{
		{ID: 1, Name: "web", Children: []int{2}},
		{ID: 2, Name: "web-eu", Children: []int{1}},
	}

	_, err := buildManagedInventory("", nil, groups, nil)

	assert.EqualError(t, err, "group cycle: web, web-eu, web")
}

func TestBuildManagedInventory_GroupErrorNamesGroup(t *testing.T) {
	groups := []db.InventoryGroup{{ID: 1, Name: "web", Variables: "[oops"}}

	_, err := buildManagedInventory("", nil, groups, nil)

	assert.ErrorContains(t, err, "group web: ")
}

func TestBuildManagedInventory_KeepsOnlyAllAtTheTop(t *testing.T) {
	hosts := []db.InventoryHost{{ID: 1, Name: "web-01"}}
	groups := []db.InventoryGroup{{ID: 1, Name: "plugin"}}
	members := []db.InventoryGroupHost{{GroupID: 1, HostID: 1}}

	text, err := buildManagedInventory("", hosts, groups, members)

	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(text), &parsed))
	assert.Len(t, parsed, 1)
	assert.Contains(t, parsed, "all")
}
