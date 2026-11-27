package server

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ansibleInventoryOutput = `{
  "_meta": {"hostvars": {
    "web-01": {"ansible_host": "10.0.0.1", "ports": [80, 443], "secret": {"__ansible_vault": "$ANSIBLE_VAULT;1.1;AES256\n6162"}},
    "fe80::1": {"role": "gateway"},
    "web-02": {"weight": 3, "memory": 1073741824, "files": 1000000, "ratio": 3.5, "serial": 12345678901234567890, "big": 2000000.0, "one": 1.0, "huge": 18446744073709551617}
  }},
  "all": {"children": ["ungrouped", "web"], "vars": {"ntp": "pool.ntp.org"}},
  "web": {"hosts": ["web-02", "web-01"], "children": ["canary"], "vars": {"http_port": 8080}},
  "canary": {"hosts": ["web-02"]},
  "ungrouped": {"hosts": ["fe80::1", "plain"], "vars": {"ansible_user": "deploy", "role": "edge"}}
}`

func TestParseAnsibleInventory(t *testing.T) {
	doc, err := ParseAnsibleInventory([]byte(ansibleInventoryOutput))
	require.NoError(t, err)

	names := []string{}
	for _, host := range doc.Hosts {
		names = append(names, host.Name)
	}
	require.Equal(t, []string{"fe80::1", "plain", "web-01", "web-02"}, names)

	assert.Equal(t, map[string]any{"role": "gateway", "ansible_user": "deploy"}, doc.Hosts[0].Variables)
	assert.Equal(t, map[string]any{"role": "edge", "ansible_user": "deploy"}, doc.Hosts[1].Variables)
	assert.Equal(t, "10.0.0.1", doc.Hosts[2].Variables["ansible_host"])
	assert.Equal(t, []any{int64(80), int64(443)}, doc.Hosts[2].Variables["ports"])
	assert.Equal(t, map[string]any{"__ansible_vault": "$ANSIBLE_VAULT;1.1;AES256\n6162"}, doc.Hosts[2].Variables["secret"])
	huge, _ := new(big.Int).SetString("18446744073709551617", 10)
	assert.Equal(t, map[string]any{
		"weight": int64(3), "memory": int64(1073741824), "files": int64(1000000), "ratio": 3.5, "serial": uint64(12345678901234567890),
		"big": 2000000.0, "one": 1.0, "huge": huge,
	}, doc.Hosts[3].Variables)

	text, err := db.InventoryVariablesText(doc.Hosts[3].Variables)
	require.NoError(t, err)
	assert.Equal(t, "big: 2000000.0\nfiles: 1000000\nhuge: 18446744073709551617\nmemory: 1073741824\none: 1.0\nratio: 3.5\nserial: 12345678901234567890\nweight: 3\n", text)

	assert.Equal(t, []db.InventoryDocumentGroup{
		{Name: "canary", Hosts: []string{"web-02"}},
		{Name: "web", Variables: map[string]any{"http_port": int64(8080)}, Hosts: []string{"web-01", "web-02"}, Children: []string{"canary"}},
	}, doc.Groups)
	assert.Equal(t, map[string]any{"ntp": "pool.ntp.org"}, doc.Variables)
}

func TestParseAnsibleInventory_Empty(t *testing.T) {
	doc, err := ParseAnsibleInventory([]byte(`{"_meta":{"hostvars":{}},"all":{"children":["ungrouped"]}}`))
	require.NoError(t, err)

	assert.Empty(t, doc.Hosts)
	assert.Empty(t, doc.Groups)
	assert.Empty(t, doc.Variables)
}

func TestParseAnsibleInventory_EmptyGroup(t *testing.T) {
	doc, err := ParseAnsibleInventory([]byte(`{"_meta":{"hostvars":{}},"all":{"children":["ungrouped","web","spare"]},"web":{"hosts":["a"],"children":["empty"]},"ungrouped":{"hosts":[]}}`))
	require.NoError(t, err)

	assert.Equal(t, []db.InventoryDocumentGroup{
		{Name: "empty"},
		{Name: "spare"},
		{Name: "web", Hosts: []string{"a"}, Children: []string{"empty"}},
	}, doc.Groups)
}

func TestParseAnsibleInventory_UngroupedVariablesTooLargeToCopy(t *testing.T) {
	hosts := make([]string, 200)
	for i := range hosts {
		hosts[i] = fmt.Sprintf("h%d", i)
	}
	output, err := json.Marshal(map[string]any{
		"all":       map[string]any{"children": []string{"ungrouped"}},
		"ungrouped": map[string]any{"hosts": hosts, "vars": map[string]any{"blob": strings.Repeat("a", 256<<10)}},
	})
	require.NoError(t, err)

	_, err = ParseAnsibleInventory(output)

	assert.EqualError(t, err, "the variables of the group ungrouped are too large to copy to its 200 hosts")
}
