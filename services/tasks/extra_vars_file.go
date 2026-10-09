package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"gopkg.in/yaml.v3"
)

// Ansible evaluates a Jinja2 expression found inside a variable's value when
// the playbook uses the variable, whatever the source of the value: a survey
// answer of "{{ lookup('pipe', 'id') }}" runs a command on the host that
// executes the task. The extra vars Semaphore passes — variable groups and
// their secrets, survey answers, workflow outputs of earlier tasks, survey
// secrets and the semaphore_vars task details — are data, never code, so they
// are handed over as a YAML file in which every string carries the !unsafe
// tag, which makes Ansible use it literally (the ALLOW_JINJA_IN_EXTRA_VARS=never
// behaviour of AWX). The file also keeps the values out of the command line.

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

// writeExtraVarsFile writes the task's extra vars — the merged environment,
// the survey secrets, the variable-group secrets of type "var" and the
// semaphore_vars details — as a YAML file for `--extra-vars @file` and returns
// its path. The file lives in the project's tmp directory, readable by the
// user the task runs as only, and is removed by Cleanup.
func (t *LocalExecutor) writeExtraVarsFile(username string, incomingVersion *string) (string, error) {
	vars, err := t.getEnvironmentExtraVars(username, incomingVersion)
	if err != nil {
		return "", err
	}
	for _, secret := range t.Environment.Secrets {
		if secret.Type == db.EnvironmentSecretVar {
			vars[secret.Name] = secret.Secret
		}
	}

	content, err := extraVarsYAML(vars)
	if err != nil {
		return "", err
	}

	dir := util.Config.GetProjectTmpDir(t.Template.ProjectID)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, fmt.Sprintf("task_%d_extra_vars_*.yml", t.Task.ID))
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

	t.extraVarsFile = filepath.Clean(path)
	return t.extraVarsFile, nil
}

func (t *LocalExecutor) destroyExtraVarsFile() {
	if t.extraVarsFile == "" {
		return
	}
	_ = os.Remove(t.extraVarsFile)
	t.extraVarsFile = ""
}
