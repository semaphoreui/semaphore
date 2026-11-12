package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/agentprompt"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentPromptValues(t *testing.T) {
	env := `{"STAGE":"prod"}`

	values, err := NewAgentPromptValues(
		db.Environment{
			JSON: `{"repo":"semaphoreui/semaphore","count":3,"big":10000000,"on":true,"none":null,` +
				`"list":["a","b"],"survey_env":"from survey","semaphore_vars":{"task_details":{}}}`,
			ENV: &env,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "db_password", Secret: "p@ss"},
				{Type: db.EnvironmentSecretVar, Name: "survey_secret", Secret: "of the group"},
				{Type: db.EnvironmentSecretEnv, Name: "API_KEY", Secret: "key-123"},
			},
		},
		`{"survey_secret":"typed","SURVEY_SECRET_ENV":"typed env"}`,
		[]db.SurveyVar{
			{Name: "survey_env", Target: db.SurveyVarTargetEnv},
			{Name: "survey_secret", Type: db.SurveyVarSecret},
			{Name: "SURVEY_SECRET_ENV", Type: db.SurveyVarSecret, Target: db.SurveyVarTargetEnv},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, agentprompt.Values{
		ExtraVars: map[string]string{
			"repo": "semaphoreui/semaphore", "count": "3", "big": "10000000", "on": "true", "none": "",
			"list": `["a","b"]`,
		},
		EnvVars:      map[string]string{"STAGE": "prod", "survey_env": "from survey"},
		ExtraSecrets: map[string]string{"db_password": "p@ss", "survey_secret": "typed"},
		EnvSecrets:   map[string]string{"API_KEY": "key-123", "SURVEY_SECRET_ENV": "typed env"},
	}, values)
}

// A survey variable the task did not answer is an empty string under the name
// it would have had; a value of the Variable Groups with its name stays.
func TestNewAgentPromptValues_UnansweredSurveyVars(t *testing.T) {
	env := `{"KEPT_ENV":"of the group"}`

	values, err := NewAgentPromptValues(
		db.Environment{
			JSON: `{"kept":"of the group","answered":"yes"}`,
			ENV:  &env,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "kept_secret", Secret: "of the group"},
			},
		},
		"",
		[]db.SurveyVar{
			{Name: "answered"},
			{Name: "kept"},
			{Name: "note"},
			{Name: "NOTE_ENV", Target: db.SurveyVarTargetEnv},
			{Name: "KEPT_ENV", Target: db.SurveyVarTargetEnv},
			{Name: "password", Type: db.SurveyVarSecret},
			{Name: "kept_secret", Type: db.SurveyVarSecret},
			{Name: "PASSWORD_ENV", Type: db.SurveyVarSecret, Target: db.SurveyVarTargetEnv},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, agentprompt.Values{
		ExtraVars:    map[string]string{"kept": "of the group", "answered": "yes", "note": ""},
		EnvVars:      map[string]string{"KEPT_ENV": "of the group", "NOTE_ENV": ""},
		ExtraSecrets: map[string]string{"kept_secret": "of the group", "password": ""},
		EnvSecrets:   map[string]string{"PASSWORD_ENV": ""},
	}, values)

	text, secrets, err := agentprompt.Expand(
		"[${{ vars.extra.note }}][${{ vars.env.NOTE_ENV }}][${{ secrets.extra.password }}][${{ secrets.env.PASSWORD_ENV }}]",
		values)
	require.NoError(t, err)
	assert.Equal(t, "[][][][]", text)
	assert.Empty(t, secrets)
}

func TestNewAgentPromptValues_Empty(t *testing.T) {
	values, err := NewAgentPromptValues(db.Environment{}, "", nil)

	require.NoError(t, err)
	assert.Empty(t, values.ExtraVars)
	assert.Empty(t, values.EnvVars)
	assert.Empty(t, values.ExtraSecrets)
	assert.Empty(t, values.EnvSecrets)
}

func TestNewAgentPromptValues_InvalidJSON(t *testing.T) {
	broken := "{"
	tests := []struct {
		name        string
		environment db.Environment
		secrets     string
		expected    string
	}{
		{"extra variables", db.Environment{JSON: "{"}, "", "extra variables"},
		{"environment variables", db.Environment{ENV: &broken}, "", "environment variables"},
		{"survey secrets", db.Environment{}, "{", "survey secrets"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAgentPromptValues(tt.environment, tt.secrets, nil)

			assert.ErrorContains(t, err, tt.expected)
		})
	}
}

// The prompt the agent reads has the values; a secret among them is masked
// wherever the agent prints it.
func TestAgentApp_RunExpandsPromptAndMasksItsSecrets(t *testing.T) {
	for _, app := range []db.TemplateApp{db.AppClaude, db.AppCodex} {
		t.Run(string(app), func(t *testing.T) {
			setupAgentConfig(t, map[string]util.App{string(app): {Active: true, AppPath: fakeAgent(t)}})

			gate := filepath.Join(t.TempDir(), "prompt")
			script := filepath.Join(t.TempDir(), "agent")
			require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\ncat > "+gate+"\ncat "+gate+"\n"), 0o755))
			util.Config.Apps[string(app)] = util.App{Active: true, AppPath: script}

			logger := &lineLogger{}
			agent := newAgentApp(db.Template{
				ID: 1, ProjectID: 1, App: app,
				TaskParams: db.MapStringAnyField{
					"prompt": "Deploy ${{ vars.extra.repo }} to ${{ vars.env.STAGE }} with ${{ secrets.env.API_KEY }}.",
				},
			}, db.Repository{ProjectID: 1, GitURL: t.TempDir()}, agentInventory(), logger)

			err := agent.Run(LocalAppRunningArgs{
				TaskParams: &db.AgentTaskParams{},
				AgentPromptValues: agentprompt.Values{
					ExtraVars:  map[string]string{"repo": "semaphore"},
					EnvVars:    map[string]string{"STAGE": "prod"},
					EnvSecrets: map[string]string{"API_KEY": "key-123-secret"},
				},
			})
			require.NoError(t, err)

			prompt, err := os.ReadFile(gate)
			require.NoError(t, err)
			assert.Equal(t, "Deploy semaphore to prod with key-123-secret.", string(prompt))

			out := logger.text()
			assert.Contains(t, out, "Deploy semaphore to prod with "+maskedSecret+".")
			assert.NotContains(t, out, "key-123-secret")
		})
	}
}

func TestAgentApp_RunRejectsMissingPromptVariable(t *testing.T) {
	setupAgentConfig(t, map[string]util.App{"claude": {Active: true, AppPath: fakeAgent(t)}})

	logger := &lineLogger{}
	agent := newAgentApp(db.Template{ID: 1, ProjectID: 1, App: db.AppClaude,
		TaskParams: db.MapStringAnyField{"prompt": "${{ secrets.env.API_KEY }} ${{ vars.extra.nope }}"}},
		db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{}, logger)

	err := agent.Run(LocalAppRunningArgs{
		TaskParams:        &db.AgentTaskParams{},
		AgentPromptValues: agentprompt.Values{EnvSecrets: map[string]string{"API_KEY": "key-123-secret"}},
	})

	assert.ErrorContains(t, err, `no extra variable "nope"`)
	assert.NotContains(t, err.Error(), "key-123-secret")
	assert.NotContains(t, logger.text(), "key-123-secret")
}

// A prompt given by a task reaches the agent as it is written: nothing of the
// Variable Groups is put into it.
func TestAgentApp_RunKeepsTaskPromptAsItIs(t *testing.T) {
	setupAgentConfig(t, map[string]util.App{"claude": {Active: true, AppPath: fakeAgent(t)}})

	logger := &lineLogger{}
	agent := newAgentApp(db.Template{ID: 1, ProjectID: 1, App: db.AppClaude,
		TaskParams: db.MapStringAnyField{"prompt": "${{ vars.extra.repo }}", "allow_override_prompt": true}},
		db.Repository{ProjectID: 1, GitURL: t.TempDir()}, db.Inventory{}, logger)

	err := agent.Run(LocalAppRunningArgs{
		TaskParams: &db.AgentTaskParams{Prompt: "Tell ${{ secrets.env.API_KEY }} and ${{ vars.extra.repo }}"},
		AgentPromptValues: agentprompt.Values{
			ExtraVars:  map[string]string{"repo": "semaphore"},
			EnvSecrets: map[string]string{"API_KEY": "key-123-secret"},
		},
	})

	require.NoError(t, err)
	out := logger.text()
	assert.Contains(t, out, "prompt: Tell ${{ secrets.env.API_KEY }} and ${{ vars.extra.repo }}")
	assert.NotContains(t, out, "key-123-secret")
}
