package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ansible_vault"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
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

// vaultArgs returns the --vault-id arguments and, for the one of the secrets
// file, the password found in inputs under the prompt Ansible prints for it.
func vaultArgs(t *testing.T, executor *LocalExecutor, args []string, inputs map[string]string) (vaultIDs []string, password string) {
	t.Helper()
	for _, arg := range args {
		if strings.HasPrefix(arg, "--vault-id=") {
			vaultIDs = append(vaultIDs, strings.TrimPrefix(arg, "--vault-id="))
		}
	}
	if executor.secretVaultID != "" {
		password = inputs["Vault password ("+executor.secretVaultID+"):"]
	}
	return
}

func TestLocalExecutor_getPlaybookArgs_ExtraVarsFile(t *testing.T) {
	executor := newExtraVarsExecutor(t)

	args, inputs, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)

	tmpDir := regexpQuote(util.Config.GetProjectTmpDir(3))
	joined := strings.Join(args, " ")
	require.Regexp(t, `--extra-vars @`+tmpDir+`/task_42_extra_vars_[^ ]+\.yml --extra-vars @`+tmpDir+`/task_42_secret_vars_[^ ]+\.yml`, joined,
		"the open file first, the secrets file last so a secret wins over a group variable of the same name")
	assert.Equal(t, 2, strings.Count(joined, "--extra-vars"), "the two files are the only --extra-vars")
	assert.NotContains(t, joined, "{{", "no expression reaches the command line")
	assert.NotContains(t, joined, "p@ss", "no secret reaches the command line")
	assert.NotContains(t, joined, "secret_expr", "no survey secret reaches the command line")

	// The one-off vault: a random id on the command line, the password only in
	// the answers to the prompt.
	vaultIDs, password := vaultArgs(t, executor, args, inputs)
	require.Len(t, vaultIDs, 1)
	assert.Regexp(t, `^[0-9a-f-]{36}@prompt$`, vaultIDs[0])
	assert.Equal(t, executor.secretVaultID+"@prompt", vaultIDs[0])
	require.Regexp(t, `^[0-9a-f-]{36}$`, password, "the password is a GUID answered on the prompt")
	assert.NotEqual(t, executor.secretVaultID, password)
	assert.NotContains(t, joined, password, "the password is not an argument")
	assert.Len(t, inputs, 1)

	// The open file: no secret keys, every value !unsafe.
	openPath := executor.extraVarsFile
	info, err := os.Stat(openPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	assert.Equal(t, filepath.Clean(util.Config.GetProjectTmpDir(3)), filepath.Dir(openPath))

	openContent, err := os.ReadFile(openPath)
	require.NoError(t, err)
	openTags := tagsOf(t, openContent)
	assert.Equal(t, "!unsafe", openTags[".image_tag"], "survey answer")
	assert.Equal(t, "!unsafe", openTags[".group_var"], "variable group value")
	assert.Equal(t, "!unsafe", openTags[".semaphore_vars.task_details.username"])
	assert.NotContains(t, openTags, ".token", "a survey secret is not in the open file")
	assert.NotContains(t, openTags, ".db_password", "a variable group secret is not in the open file")
	assert.NotContains(t, openTags, ".ENV_ONLY", "an env secret is not an extra var")
	assert.NotContains(t, string(openContent), "p@ss")
	assert.NotContains(t, string(openContent), "secret_expr")

	// The secrets file: vault-encrypted, decrypts with the prompt's password to
	// a YAML document of !unsafe secrets and nothing else.
	vaultPath := executor.secretVarsFile
	info, err = os.Stat(vaultPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	assert.Equal(t, filepath.Dir(openPath), filepath.Dir(vaultPath))

	vaulttext, err := os.ReadFile(vaultPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(vaulttext), "$ANSIBLE_VAULT;1.1;AES256\n"), "%s", vaulttext)
	assert.NotContains(t, string(vaulttext), "p@ss")
	assert.NotContains(t, string(vaulttext), "token")

	plaintext, err := ansible_vault.Decrypt(vaulttext, password)
	require.NoError(t, err)
	secretTags := tagsOf(t, plaintext)
	assert.Equal(t, map[string]string{".token": "!unsafe", ".db_password": "!unsafe"}, secretTags,
		"the secrets, each !unsafe, and nothing else")

	var secrets map[string]any
	require.NoError(t, yaml.Unmarshal(plaintext, &secrets))
	assert.Equal(t, "p@ss {{ 1+1 }}", secrets["db_password"], "a variable group secret, expression kept literal")
	assert.Equal(t, "{{ secret_expr }}", secrets["token"], "a survey secret, expression kept literal")

	executor.Cleanup()
	_, err = os.Stat(openPath)
	assert.True(t, os.IsNotExist(err), "Cleanup removes the open file")
	_, err = os.Stat(vaultPath)
	assert.True(t, os.IsNotExist(err), "Cleanup removes the secrets file")
	assert.Empty(t, executor.extraVarsFile)
	assert.Empty(t, executor.secretVarsFile)
	assert.Empty(t, executor.secretVaultID)
	assert.Empty(t, executor.secretVaultPassword, "Cleanup forgets the password")
}

func TestLocalExecutor_getPlaybookArgs_SecretsWithTemplateVault(t *testing.T) {
	executor := newExtraVarsExecutor(t)
	executor.vaultFileInstallations = map[string]ssh.AccessKeyInstallation{
		"default": {Password: "template-vault-pass"},
	}

	args, inputs, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)

	vaultIDs, password := vaultArgs(t, executor, args, inputs)
	assert.ElementsMatch(t, []string{"default@prompt", executor.secretVaultID + "@prompt"}, vaultIDs,
		"the template's vault and the one-off vault of the secrets")
	assert.Equal(t, "template-vault-pass", inputs["Vault password (default):"])
	assert.Equal(t, executor.secretVaultPassword, password)
	assert.Len(t, inputs, 2, "both prompts are answered")
	assert.NotContains(t, strings.Join(args, " "), "template-vault-pass")

	executor.Cleanup()
}

func TestLocalExecutor_getPlaybookArgs_SurveySecretWithEnvTarget(t *testing.T) {
	executor := newExtraVarsExecutor(t)
	executor.Template.SurveyVars = []db.SurveyVar{
		{Name: "api_key", Type: db.SurveyVarSecret, Target: db.SurveyVarTargetEnv},
	}
	executor.Secret = `{"token":"{{ secret_expr }}","api_key":"env-only-secret"}`

	args, inputs, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)
	defer executor.Cleanup()

	openContent, err := os.ReadFile(executor.extraVarsFile)
	require.NoError(t, err)
	assert.NotContains(t, string(openContent), "api_key")
	assert.NotContains(t, string(openContent), "env-only-secret")

	vaulttext, err := os.ReadFile(executor.secretVarsFile)
	require.NoError(t, err)
	_, password := vaultArgs(t, executor, args, inputs)
	plaintext, err := ansible_vault.Decrypt(vaulttext, password)
	require.NoError(t, err)
	var secrets map[string]any
	require.NoError(t, yaml.Unmarshal(plaintext, &secrets))
	assert.Equal(t, map[string]any{"token": "{{ secret_expr }}", "db_password": "p@ss {{ 1+1 }}"}, secrets,
		"a secret with the env target is delivered through the environment, not through a file")

	env, err := executor.getSurveyEnvVars()
	require.NoError(t, err)
	assert.Equal(t, []string{"api_key=env-only-secret"}, env)
}

func TestLocalExecutor_getPlaybookArgs_NoSecrets(t *testing.T) {
	executor := newExtraVarsExecutor(t)
	executor.Secret = ""
	executor.Environment.Secrets = []db.EnvironmentSecret{
		{Type: db.EnvironmentSecretEnv, Name: "ENV_ONLY", Secret: "not-a-var"},
	}

	args, inputs, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)
	defer executor.Cleanup()

	joined := strings.Join(args, " ")
	assert.Equal(t, 1, strings.Count(joined, "--extra-vars"), "no secrets: the open file only")
	assert.NotContains(t, joined, "--vault-id", "no secrets: no vault")
	assert.NotContains(t, joined, "secret_vars")
	assert.Empty(t, inputs)
	assert.Empty(t, executor.secretVarsFile)
	assert.Empty(t, executor.secretVaultID)
	assert.Empty(t, executor.secretVaultPassword)
	assert.FileExists(t, executor.extraVarsFile)
}

func TestLocalExecutor_takeSecretExtraVars_SecretWinsOverGroupVariable(t *testing.T) {
	executor := newExtraVarsExecutor(t)
	// The group defines db_password in clear and as a secret; the survey secret
	// token shadows a survey answer of the same name.
	executor.Environment.JSON = `{"db_password":"from-group-json","token":"from-survey-answer"}`
	executor.Task.Environment = executor.Environment.JSON

	vars, err := executor.getEnvironmentExtraVars("denis", nil)
	require.NoError(t, err)
	secrets, err := executor.takeSecretExtraVars(vars)
	require.NoError(t, err)

	assert.Equal(t, map[string]any{"token": "{{ secret_expr }}", "db_password": "p@ss {{ 1+1 }}"}, secrets)
	assert.NotContains(t, vars, "token")
	assert.NotContains(t, vars, "db_password")
	assert.Contains(t, vars, "semaphore_vars")
}

// The DEBUG entries of the extra_vars namespace tell the operator what was
// written and removed, by variable name and file, and must never carry a
// value or the vault password.
func TestLocalExecutor_extraVarsDebugLog(t *testing.T) {
	executor := newExtraVarsExecutor(t)

	logger := log.StandardLogger()
	level, formatter := logger.GetLevel(), logger.Formatter
	logger.SetLevel(log.DebugLevel)
	logger.SetFormatter(&log.TextFormatter{})
	hook := logtest.NewLocal(logger)
	t.Cleanup(func() {
		hook.Reset()
		logger.ReplaceHooks(make(log.LevelHooks))
		logger.SetLevel(level)
		logger.SetFormatter(formatter)
	})

	args, inputs, err := executor.getPlaybookArgs("denis", nil)
	require.NoError(t, err)
	openPath, vaultPath := executor.extraVarsFile, executor.secretVarsFile
	_, password := vaultArgs(t, executor, args, inputs)
	executor.Cleanup()

	var messages []string
	var rendered strings.Builder
	for _, entry := range hook.AllEntries() {
		if entry.Data["context"] != "extra_vars" {
			continue
		}
		assert.Equal(t, log.DebugLevel, entry.Level)
		assert.Equal(t, 42, entry.Data["task_id"])
		assert.Equal(t, 3, entry.Data["project_id"])
		messages = append(messages, entry.Message)
		line, err := entry.String()
		require.NoError(t, err)
		rendered.WriteString(line)
	}

	assert.Equal(t, []string{
		"Extra vars split into open and secret",
		"Open extra vars file written",
		"Secret extra vars file written, vault-encrypted",
		"Secret extra vars vault added to the ansible-playbook arguments, password answered on the prompt",
		"Extra vars file removed",
		"Extra vars file removed",
		"Secret extra vars vault password forgotten",
	}, messages)

	text := rendered.String()
	for _, name := range []string{"image_tag", "group_var", "semaphore_vars", "token", "db_password"} {
		assert.Contains(t, text, name, "variable names are logged")
	}
	assert.Contains(t, text, openPath)
	assert.Contains(t, text, vaultPath)
	assert.Contains(t, text, executor.extraVarsFile+"", "paths are logged")

	for _, secret := range []string{"p@ss", "secret_expr", "lookup('pipe'", "{{ ok }}", password} {
		assert.NotContains(t, text, secret, "no value and no password in the debug log")
	}
}

func regexpQuote(s string) string {
	return strings.NewReplacer(`.`, `\.`, `/`, `/`, `-`, `\-`).Replace(s)
}
