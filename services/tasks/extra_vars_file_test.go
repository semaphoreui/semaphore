package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// tagsOf collects the YAML tag of every scalar in a document, keyed by its path.
func tagsOf(t *testing.T, document []byte) map[string]string {
	t.Helper()
	var root yaml.Node
	require.NoError(t, yaml.Unmarshal(document, &root))
	tags := map[string]string{}
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
			tags[path] = n.Tag
		}
	}
	walk(&root, "")
	return tags
}

func TestExtraVarsYAML(t *testing.T) {
	vars := map[string]any{
		"image_tag": "{{ lookup('pipe', 'id') }}",
		"replicas":  float64(3),
		"enabled":   true,
		"nothing":   nil,
		"subnets":   []any{"subnet-1", "{{ 6*7 }}"},
		"config":    map[string]any{"region": "{{ region }}", "depth": map[string]any{"leaf": "x"}},
		"group_var": "{{ from_variable_group }}",
		"semaphore_vars": map[string]any{
			"task_details": map[string]any{"commit_message": "{{ evil }}", "id": 7},
		},
	}

	document, err := extraVarsYAML(vars)
	require.NoError(t, err)
	tags := tagsOf(t, document)

	assert.Equal(t, "!unsafe", tags[".image_tag"])
	assert.Equal(t, "!unsafe", tags[".group_var"], "variable groups are data too")
	assert.Equal(t, "!unsafe", tags[".subnets[0]"])
	assert.Equal(t, "!unsafe", tags[".subnets[1]"])
	assert.Equal(t, "!unsafe", tags[".config.region"])
	assert.Equal(t, "!unsafe", tags[".config.depth.leaf"])
	assert.Equal(t, "!unsafe", tags[".semaphore_vars.task_details.commit_message"])
	assert.Equal(t, "!!int", tags[".semaphore_vars.task_details.id"])
	assert.Equal(t, "!!int", tags[".replicas"], "a number carries no expression")
	assert.Equal(t, "!!bool", tags[".enabled"])
	assert.Equal(t, "!!null", tags[".nothing"])

	// Keys are sorted and every string is double-quoted.
	text := string(document)
	assert.Less(t, strings.Index(text, "config:"), strings.Index(text, "image_tag:"))
	assert.Contains(t, text, `image_tag: !unsafe "{{ lookup('pipe', 'id') }}"`)
	assert.Contains(t, text, `group_var: !unsafe "{{ from_variable_group }}"`)
}

func newExtraVarsExecutor(t *testing.T) *LocalExecutor {
	t.Helper()
	util.Config = &util.ConfigType{
		TmpPath: t.TempDir(),
		Process: &util.ConfigProcess{},
	}
	inventoryID := 1
	executor := &LocalExecutor{
		Task:     db.Task{ID: 42, Environment: `{"image_tag":"{{ lookup('pipe', 'id') }}"}`},
		Template: db.Template{ProjectID: 3, Playbook: "test.yml"},
		Inventory: db.Inventory{Type: db.InventoryStatic, SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{ID: 1, Type: db.AccessKeySSH}},
		Environment: db.Environment{
			JSON: `{"image_tag":"{{ lookup('pipe', 'id') }}","group_var":"{{ ok }}"}`,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "db_password", Secret: "p@ss {{ 1+1 }}"},
				{Type: db.EnvironmentSecretEnv, Name: "ENV_ONLY", Secret: "not-a-var"},
			},
		},
		Secret: `{"token":"{{ secret_expr }}"}`,
	}
	executor.Logger = &TaskRunner{Task: executor.Task}
	return executor
}

func TestLocalExecutor_getPlaybookArgs_ExtraVarsFile(t *testing.T) {
	executor := newExtraVarsExecutor(t)

	args, _, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)

	joined := strings.Join(args, " ")
	require.Regexp(t, `--extra-vars @`+regexpQuote(util.Config.GetProjectTmpDir(3))+`/task_42_extra_vars_[^ ]+\.yml`, joined)
	assert.Equal(t, 1, strings.Count(joined, "--extra-vars"), "the file is the only --extra-vars")
	assert.NotContains(t, joined, "{{", "no expression reaches the command line")
	assert.NotContains(t, joined, "p@ss", "no secret reaches the command line")

	path := executor.extraVarsFile
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	assert.Equal(t, filepath.Clean(util.Config.GetProjectTmpDir(3)), filepath.Dir(path))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	tags := tagsOf(t, content)
	assert.Equal(t, "!unsafe", tags[".image_tag"], "survey answer")
	assert.Equal(t, "!unsafe", tags[".group_var"], "variable group value")
	assert.Equal(t, "!unsafe", tags[".token"], "survey secret")
	assert.Equal(t, "!unsafe", tags[".db_password"], "variable group secret of type var")
	assert.NotContains(t, tags, ".ENV_ONLY", "an env secret is not an extra var")
	assert.Equal(t, "!unsafe", tags[".semaphore_vars.task_details.username"])

	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal(content, &parsed))
	assert.Equal(t, "p@ss {{ 1+1 }}", parsed["db_password"])

	executor.Cleanup()
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "Cleanup removes the file")
	assert.Empty(t, executor.extraVarsFile)
}

func regexpQuote(s string) string {
	return strings.NewReplacer(`.`, `\.`, `/`, `/`, `-`, `\-`).Replace(s)
}
