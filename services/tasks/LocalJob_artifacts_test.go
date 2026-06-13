package tasks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

func setupArtifactsConfig(t *testing.T) {
	t.Helper()
	prev := util.Config
	t.Cleanup(func() { util.Config = prev })
	util.Config = &util.ConfigType{}
}

func TestLocalJob_getEnvironmentExtraVarsInjectsWorkflowArtifacts(t *testing.T) {
	setupArtifactsConfig(t)
	job := &LocalJob{
		Task:     db.Task{ID: 7},
		Template: db.Template{App: db.AppAnsible, Type: db.TemplateTask},
		Environment: db.Environment{
			JSON: `{"existing":"value"}`,
		},
		WorkflowArtifacts: map[string]any{
			"deployed_version": "1.2.3",
			"ready":            true,
		},
	}

	extraVars, err := job.getEnvironmentExtraVars("alice", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if extraVars["existing"] != "value" {
		t.Fatalf("environment values must survive merging, got %v", extraVars["existing"])
	}
	if extraVars["deployed_version"] != "1.2.3" {
		t.Fatalf("expected deployed_version flattened at top level, got %v", extraVars["deployed_version"])
	}
	if extraVars["ready"] != true {
		t.Fatalf("expected ready flattened at top level, got %v", extraVars["ready"])
	}
	if _, ok := extraVars["semaphore_workflow_artifacts"]; !ok {
		t.Fatalf("expected namespaced semaphore_workflow_artifacts key")
	}
	if _, ok := extraVars["semaphore_vars"]; !ok {
		t.Fatalf("expected semaphore_vars to be preserved")
	}
}

func TestLocalJob_getEnvironmentExtraVarsJSONInjectsWorkflowArtifacts(t *testing.T) {
	setupArtifactsConfig(t)
	job := &LocalJob{
		Task:     db.Task{ID: 7},
		Template: db.Template{App: db.AppAnsible, Type: db.TemplateTask},
		Environment: db.Environment{
			JSON: `{"existing":"value"}`,
		},
		Secret: `{"secret_key":"shh"}`,
		WorkflowArtifacts: map[string]any{
			"deployed_version": "1.2.3",
		},
	}

	raw, err := job.getEnvironmentExtraVarsJSON("alice", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded["deployed_version"] != "1.2.3" {
		t.Fatalf("missing artifact key in JSON output: %v", decoded)
	}
	if decoded["existing"] != "value" {
		t.Fatalf("missing existing env var: %v", decoded)
	}
	if decoded["secret_key"] != "shh" {
		t.Fatalf("missing secret var: %v", decoded)
	}
}

func TestLocalJob_getShellEnvironmentExtraENVIncludesArtifacts(t *testing.T) {
	setupArtifactsConfig(t)
	job := &LocalJob{
		Task:     db.Task{ID: 7},
		Template: db.Template{App: db.AppBash, Type: db.TemplateTask},
		WorkflowArtifacts: map[string]any{
			"deployed_version": "1.2.3",
			"ready":            true,
			"nested":           map[string]any{"x": 1},
		},
	}

	env := job.getShellEnvironmentExtraENV("alice", nil)
	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "SEMAPHORE_WF_DEPLOYED_VERSION=") {
		t.Fatalf("expected SEMAPHORE_WF_DEPLOYED_VERSION env var, got: %s", joined)
	}
	if !strings.Contains(joined, "SEMAPHORE_WF_READY=") {
		t.Fatalf("expected SEMAPHORE_WF_READY env var, got: %s", joined)
	}
	if strings.Contains(joined, "SEMAPHORE_WF_NESTED=") {
		t.Fatalf("nested values must not be exported, got: %s", joined)
	}
}

func TestLocalJob_applyWorkflowArtifactsSkipsReservedKeys(t *testing.T) {
	job := &LocalJob{
		WorkflowArtifacts: map[string]any{
			"safe_key":      "ok",
			"semaphore_vars": "evil", // would clobber semaphore_vars; must be skipped
			"task_details":  "evil",
		},
	}
	out := map[string]any{}
	job.applyWorkflowArtifacts(out)
	if out["safe_key"] != "ok" {
		t.Fatalf("safe key was dropped")
	}
	if v, ok := out["semaphore_vars"]; ok && v == "evil" {
		t.Fatalf("reserved key semaphore_vars was overwritten by artifacts")
	}
	if out["task_details"] == "evil" {
		t.Fatalf("reserved key task_details was overwritten")
	}
}
