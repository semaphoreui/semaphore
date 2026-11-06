package db

import (
	"testing"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplate_Validate_App(t *testing.T) {
	orig := util.Config
	defer func() { util.Config = orig }()

	util.Config = &util.ConfigType{
		Apps: map[string]util.App{"bash": {Active: true}},
	}

	// Arbitrary binary not in the whitelist must be rejected before persisting.
	tpl := &Template{Name: "rce", Playbook: "-c", App: "/bin/sh"}
	assert.EqualError(t, tpl.Validate(), "invalid app: /bin/sh")

	// Whitelisted app is accepted.
	ok := &Template{Name: "legit", Playbook: "script.sh", App: "bash"}
	assert.NoError(t, ok.Validate())

	// Empty legacy app runs no command and stays allowed.
	empty := &Template{Name: "legacy", Playbook: "site.yml", App: ""}
	assert.NoError(t, empty.Validate())
}

func TestTemplateApp_IsAgent(t *testing.T) {
	tests := []struct {
		app      TemplateApp
		expected bool
	}{
		{AppCodex, true},
		{AppClaude, true},
		{AppAnsible, false},
		{AppTerraform, false},
		{AppBash, false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.app), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.app.IsAgent())
		})
	}
}

func TestTemplateApp_AgentInventoryTypes(t *testing.T) {
	for _, app := range []TemplateApp{AppCodex, AppClaude} {
		t.Run(string(app), func(t *testing.T) {
			assert.Equal(t, []InventoryType{InventoryMcp}, app.InventoryTypes())
			assert.True(t, app.HasInventoryType(InventoryMcp))
			assert.False(t, app.HasInventoryType(InventoryStatic))
			assert.True(t, app.RepositoryFileFilter().OnlyDirectories)
		})
	}

	for _, app := range []TemplateApp{AppAnsible, AppTerraform, AppTofu, AppBash} {
		t.Run(string(app)+" has no mcp", func(t *testing.T) {
			assert.False(t, app.HasInventoryType(InventoryMcp))
		})
	}
}

func agentTemplate(app TemplateApp, params MapStringAnyField) *Template {
	return &Template{Name: "agent", App: app, TaskParams: params}
}

func TestTemplate_Validate_Agent(t *testing.T) {
	orig := util.Config
	defer func() { util.Config = orig }()

	util.Config = &util.ConfigType{
		Apps: map[string]util.App{"codex": {Active: true}, "claude": {Active: true}},
	}

	tests := []struct {
		name    string
		params  MapStringAnyField
		errText string
	}{
		{"prompt set, no playbook needed", MapStringAnyField{"prompt": "fix the tests"}, ""},
		{"empty prompt but override allowed", MapStringAnyField{"allow_override_prompt": true}, ""},
		{"no prompt", MapStringAnyField{}, "agent template prompt can not be empty"},
		{"blank prompt", MapStringAnyField{"prompt": "   "}, "agent template prompt can not be empty"},
		{"wrong prompt type", MapStringAnyField{"prompt": 42}, "invalid task params"},
	}
	for _, tt := range tests {
		for _, app := range []TemplateApp{AppCodex, AppClaude} {
			t.Run(string(app)+"/"+tt.name, func(t *testing.T) {
				err := agentTemplate(app, tt.params).Validate()
				if tt.errText == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, tt.errText)
				}
			})
		}
	}
}

func TestTemplate_AgentPrompt(t *testing.T) {
	tests := []struct {
		name       string
		tplParams  MapStringAnyField
		taskPrompt string
		expected   string
		errText    string
	}{
		{"template prompt", MapStringAnyField{"prompt": "tpl"}, "", "tpl", ""},
		{"task prompt allowed", MapStringAnyField{"prompt": "tpl", "allow_override_prompt": true}, "task", "task", ""},
		{"blank task prompt falls back", MapStringAnyField{"prompt": "tpl", "allow_override_prompt": true}, "  ", "tpl", ""},
		{"task prompt not allowed", MapStringAnyField{"prompt": "tpl"}, "task", "", "does not allow overriding the prompt"},
		{"nothing to run", MapStringAnyField{"allow_override_prompt": true}, "", "", "agent prompt can not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := agentTemplate(AppClaude, tt.tplParams)
			prompt, err := tpl.AgentPrompt(AgentTaskParams{Prompt: tt.taskPrompt})
			if tt.errText == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, prompt)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func TestTask_ValidateNewTask_AgentPrompt(t *testing.T) {
	tpl := agentTemplate(AppCodex, MapStringAnyField{"prompt": "tpl"})

	ok := Task{}
	assert.NoError(t, ok.ValidateNewTask(*tpl))

	overridden := Task{Params: MapStringAnyField{"prompt": "other"}}
	err := overridden.ValidateNewTask(*tpl)
	var validationErr *common_errors.ValidationError
	assert.ErrorAs(t, err, &validationErr)

	tpl.TaskParams["allow_override_prompt"] = true
	assert.NoError(t, overridden.ValidateNewTask(*tpl))
}
