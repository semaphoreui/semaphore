package tasks

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// The extra vars reach Ansible as a YAML file, so a string that looks like a
// YAML number, boolean, null or date must still arrive as a string. The
// webhook payload {"release":"2026.08"} is the reported case: the task's
// environment keeps the string, the playbook sees a float.
//
// Ansible >= 2.19 (ansible-core 2.19, bundled in ansible 11+) constructs an
// !unsafe scalar by resolving its value again as a *plain* scalar — the
// quoting style is ignored (construct_yaml_unsafe ->
// _resolve_and_construct_object -> resolver.resolve(..., implicit=(True, False))
// in ansible/_internal/_yaml/_constructor.py). The two tests below check the
// document from both sides: resolved by Ansible itself when an installation is
// found, and by a Go re-implementation of that one rule otherwise.

// typedStringVars are extra vars whose string values a YAML resolver would
// take for another type, next to values of those real types, at every
// nesting level the serializer handles.
func typedStringVars() map[string]any {
	return map[string]any{
		"release":    "2026.08",    // the reported webhook value
		"identifier": "00123",      // leading zeros: octal 83 to a YAML 1.1 resolver
		"flag":       "true",       // bool
		"nothing":    "null",       // null
		"day":        "2026-08-01", // date
		"expr":       "{{ lookup('pipe', 'id') }}",
		"header":     "#jinja2: variable_start_string:'[[', variable_end_string:']]'\n[[ 6*7 ]]",
		"nested": map[string]any{
			"release": "2026.08",
			"expr":    "{{ 6*7 }}",
		},
		"list": []any{"2026.08", "007", "{{ 6*7 }}"},
		// Real numbers, booleans and nulls must stay what they are.
		"replicas": float64(3),
		"enabled":  true,
		"absent":   nil,
	}
}

// stringPaths are the paths in typedStringVars whose value is a string, with
// the value expected at each.
var stringPaths = map[string]string{
	".release":        "2026.08",
	".identifier":     "00123",
	".flag":           "true",
	".nothing":        "null",
	".day":            "2026-08-01",
	".expr":           "{{ lookup('pipe', 'id') }}",
	".header":         "#jinja2: variable_start_string:'[[', variable_end_string:']]'\n[[ 6*7 ]]",
	".nested.release": "2026.08",
	".nested.expr":    "{{ 6*7 }}",
	".list[0]":        "2026.08",
	".list[1]":        "007",
	".list[2]":        "{{ 6*7 }}",
}

// resolveLikeAnsibleUnsafe reads document the way the Ansible >= 2.19 loader
// does: a scalar tagged !unsafe is resolved as if it were plain, whatever its
// quoting; everything else is read as written. It returns the value found at
// every scalar path.
func resolveLikeAnsibleUnsafe(t *testing.T, document []byte) map[string]any {
	t.Helper()
	var root yaml.Node
	require.NoError(t, yaml.Unmarshal(document, &root))

	values := map[string]any{}
	var walk func(n *yaml.Node, path string)
	walk = func(n *yaml.Node, path string) {
		switch n.Kind {
		case yaml.DocumentNode:
			walk(n.Content[0], path)
		case yaml.MappingNode:
			for i := 0; i < len(n.Content); i += 2 {
				walk(n.Content[i+1], path+"."+n.Content[i].Value)
			}
		case yaml.SequenceNode:
			for i, c := range n.Content {
				walk(c, path+"["+string(rune('0'+i))+"]")
			}
		case yaml.ScalarNode:
			scalar := *n
			if scalar.Tag == "!unsafe" {
				// Ansible: resolver.resolve(ScalarNode, node.value, implicit=(True, False))
				scalar.Tag = ""
				scalar.Style = 0
			}
			var v any
			require.NoError(t, scalar.Decode(&v), "%s", path)
			values[path] = v
		}
	}
	walk(&root, "")
	return values
}

func TestExtraVarsYAML_StringsSurviveUnsafeResolution(t *testing.T) {
	document, err := extraVarsYAML(typedStringVars())
	require.NoError(t, err)

	values := resolveLikeAnsibleUnsafe(t, document)

	t.Logf("document:\n%s", document)
	for path, want := range stringPaths {
		assert.Equal(t, want, values[path], "%s must stay the string %q", path, want)
	}
	assert.Equal(t, 3, values[".replicas"], "a real number stays a number")
	assert.Equal(t, true, values[".enabled"], "a real boolean stays a boolean")
	assert.Nil(t, values[".absent"], "a real null stays null")
}

// ansibleLoaderScript loads the YAML document on stdin with Ansible's own
// DataLoader and prints every leaf as {"path": {"value": ..., "type": ...}}.
const ansibleLoaderScript = `
import json, sys
from ansible.parsing.dataloader import DataLoader

def walk(value, path, out):
    if isinstance(value, dict):
        for k, v in value.items():
            walk(v, path + "." + str(k), out)
    elif isinstance(value, list):
        for i, v in enumerate(value):
            walk(v, path + "[" + str(i) + "]", out)
    else:
        if isinstance(value, str):
            kind = "str"
        elif isinstance(value, bool):
            kind = "bool"
        elif isinstance(value, int):
            kind = "int"
        elif isinstance(value, float):
            kind = "float"
        elif value is None:
            kind = "null"
        else:
            kind = type(value).__name__
        out[path] = {"value": value if kind in ("str", "bool", "int", "float", "null") else str(value), "type": kind}

out = {}
walk(DataLoader().load(sys.stdin.read()), "", out)
json.dump(out, sys.stdout)
`

type ansibleLeaf struct {
	Value any    `json:"value"`
	Type  string `json:"type"`
}

// ansiblePython finds a Python interpreter that imports ansible:
// $SEMAPHORE_TEST_ANSIBLE_PYTHON, the interpreter of the `ansible` script on
// PATH (a pipx/uv/venv install), then python3 itself. Without one the caller
// skips.
func ansiblePython(t *testing.T) string {
	t.Helper()
	var candidates []string
	if p := os.Getenv("SEMAPHORE_TEST_ANSIBLE_PYTHON"); p != "" {
		candidates = append(candidates, p)
	}
	if script, err := exec.LookPath("ansible"); err == nil {
		if f, err := os.Open(script); err == nil {
			if line, err := bufio.NewReader(f).ReadString('\n'); err == nil && strings.HasPrefix(line, "#!") {
				candidates = append(candidates, strings.Fields(strings.TrimPrefix(line, "#!"))[0])
			}
			_ = f.Close()
		}
	}
	candidates = append(candidates, "python3")

	for _, python := range candidates {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := exec.CommandContext(ctx, python, "-c", "import ansible").Run()
		cancel()
		if err == nil {
			return python
		}
	}
	t.Skipf("no Python with ansible found (tried %q); set SEMAPHORE_TEST_ANSIBLE_PYTHON", candidates)
	return ""
}

func TestExtraVarsYAML_AnsibleLoaderKeepsStringTypes(t *testing.T) {
	python := ansiblePython(t)

	document, err := extraVarsYAML(typedStringVars())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, "-c", ansibleLoaderScript)
	cmd.Stdin = bytes.NewReader(document)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	require.NoError(t, err, "ansible loader failed: %s", stderr.String())

	var leaves map[string]ansibleLeaf
	require.NoError(t, json.Unmarshal(stdout, &leaves), "%s", stdout)

	t.Logf("document:\n%s", document)
	for path, want := range stringPaths {
		leaf := leaves[path]
		assert.Equal(t, "str", leaf.Type, "%s: Ansible loaded %v as %s, not a string", path, leaf.Value, leaf.Type)
		assert.Equal(t, want, leaf.Value, "%s", path)
	}
	assert.Equal(t, ansibleLeaf{Value: float64(3), Type: "int"}, leaves[".replicas"])
	assert.Equal(t, ansibleLeaf{Value: true, Type: "bool"}, leaves[".enabled"])
	assert.Equal(t, ansibleLeaf{Value: nil, Type: "null"}, leaves[".absent"])
}
