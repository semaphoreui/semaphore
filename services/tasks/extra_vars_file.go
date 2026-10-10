package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ansible_vault"
	"github.com/semaphoreui/semaphore/util"
	"gopkg.in/yaml.v3"
)

// Ansible evaluates a Jinja2 expression found inside a variable's value when
// the playbook uses the variable, whatever the source of the value: a survey
// answer of "{{ lookup('pipe', 'id') }}" runs a command on the host that
// executes the task. The extra vars Semaphore passes — variable groups and
// their secrets, survey answers, workflow outputs of earlier tasks, survey
// secrets and the semaphore_vars task details — are data, never code, so they
// are handed over as YAML files in which every string carries the !unsafe
// tag, which makes Ansible use it literally (the ALLOW_JINJA_IN_EXTRA_VARS=never
// behaviour of AWX). The files also keep the values out of the command line,
// and the secrets among them sit in a vault-encrypted file (see
// extraVarsFiles), so no secret is on disk in clear while the task runs.

// extraVarsYAML renders vars as a YAML document. Every string is tagged
// !unsafe, nested maps and lists included; numbers, booleans and nulls carry
// no expression and stay as they are. Keys are written in sorted order so the
// file is reproducible.
func extraVarsYAML(vars map[string]any) ([]byte, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		value := &yaml.Node{}
		if err := value.Encode(vars[k]); err != nil {
			return nil, fmt.Errorf("extra var %q: %w", k, err)
		}
		markUnsafe(value)
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
			value,
		)
	}

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{mapping}}
	return yaml.Marshal(doc)
}

// markUnsafe tags every string scalar under node with !unsafe. A string that
// yaml.v3 would write plain gets the double-quoted style, so the tag and the
// value round-trip through any YAML reader.
func markUnsafe(node *yaml.Node) {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Tag == "!!str" {
			node.Tag = "!unsafe"
			node.Style = yaml.DoubleQuotedStyle
		}
	case yaml.MappingNode:
		// Content alternates key, value: only the values carry data.
		for i := 1; i < len(node.Content); i += 2 {
			markUnsafe(node.Content[i])
		}
	case yaml.SequenceNode, yaml.DocumentNode:
		for _, child := range node.Content {
			markUnsafe(child)
		}
	}
}

// extraVarsFiles are the files the extra vars of an Ansible task travel in.
// Open holds every variable that is not a secret. Vault holds the secrets —
// survey answers of type secret and variable-group secrets of type var —
// encrypted in the Ansible Vault format with a one-off password; it is empty
// when the task has no secrets. Both live in the project's tmp directory,
// readable by the user the task runs as only, and are removed by Cleanup.
type extraVarsFiles struct {
	Open  string
	Vault string
	// VaultID names the vault in `--vault-id <VaultID>@prompt` and in the
	// prompt Ansible prints for it; VaultPassword is the answer. Both are
	// random per task and never written anywhere but the process's stdin.
	VaultID       string
	VaultPassword string
}

// writeExtraVarsFiles writes the task's extra vars — the merged environment,
// the survey answers, the semaphore_vars details — as a YAML file for
// `--extra-vars @file`, and the secrets as a second, vault-encrypted one.
func (t *LocalExecutor) writeExtraVarsFiles(username string, incomingVersion *string) (files extraVarsFiles, err error) {
	vars, err := t.getEnvironmentExtraVars(username, incomingVersion)
	if err != nil {
		return
	}
	secrets, err := t.takeSecretExtraVars(vars)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			t.destroyExtraVarsFiles()
		}
	}()

	content, err := extraVarsYAML(vars)
	if err != nil {
		return
	}
	if files.Open, err = t.writeTaskTmpFile("extra_vars", content); err != nil {
		return
	}
	t.extraVarsFile = files.Open

	if len(secrets) == 0 {
		return
	}

	plaintext, err := extraVarsYAML(secrets)
	if err != nil {
		return
	}
	vaultID, err := uuid.NewRandom()
	if err != nil {
		return
	}
	password, err := uuid.NewRandom()
	if err != nil {
		return
	}
	files.VaultID = vaultID.String()
	files.VaultPassword = password.String()

	vaulttext, err := ansible_vault.Encrypt(plaintext, files.VaultPassword)
	if err != nil {
		return
	}
	if files.Vault, err = t.writeTaskTmpFile("secret_vars", vaulttext); err != nil {
		return
	}
	t.secretVarsFile = files.Vault
	t.secretVaultID = files.VaultID
	t.secretVaultPassword = files.VaultPassword

	return
}

// takeSecretExtraVars removes the secrets from vars and returns them: the
// survey answers delivered through the Secret field (those with the env
// target are already gone from vars, getEnvironmentExtraVars drops them) and
// the variable-group secrets of type var. A secret of the same name as a
// variable of the group wins, as it did when both sat in one map.
func (t *LocalExecutor) takeSecretExtraVars(vars map[string]any) (secrets map[string]any, err error) {
	secrets = make(map[string]any)

	if t.Secret != "" {
		surveySecrets := make(map[string]any)
		if err = json.Unmarshal([]byte(t.Secret), &surveySecrets); err != nil {
			return
		}
		for name := range surveySecrets {
			if value, ok := vars[name]; ok {
				secrets[name] = value
				delete(vars, name)
			}
		}
	}

	for _, secret := range t.Environment.Secrets {
		if secret.Type == db.EnvironmentSecretVar {
			secrets[secret.Name] = secret.Secret
			delete(vars, secret.Name)
		}
	}

	return
}

// writeTaskTmpFile writes content to a new task_<id>_<kind>_*.yml in the
// project's tmp directory, readable by the task's user only.
func (t *LocalExecutor) writeTaskTmpFile(kind string, content []byte) (string, error) {
	dir := util.Config.GetProjectTmpDir(t.Template.ProjectID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, fmt.Sprintf("task_%d_%s_*.yml", t.Task.ID, kind))
	if err != nil {
		return "", err
	}
	path := f.Name()
	_ = f.Close()

	if err = os.WriteFile(path, content, 0o600); err == nil {
		err = util.ChownDir(path)
	}
	if err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return filepath.Clean(path), nil
}

// destroyExtraVarsFiles removes both files and forgets the vault password.
func (t *LocalExecutor) destroyExtraVarsFiles() {
	if t.extraVarsFile != "" {
		_ = os.Remove(t.extraVarsFile)
		t.extraVarsFile = ""
	}
	if t.secretVarsFile != "" {
		_ = os.Remove(t.secretVarsFile)
		t.secretVarsFile = ""
	}
	t.secretVaultID = ""
	t.secretVaultPassword = ""
}
