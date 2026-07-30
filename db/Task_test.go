package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func surveyTemplate() Template {
	return Template{
		SurveyVars: []SurveyVar{
			{Name: "app_version", Type: SurveyVarStr},
			{Name: "replicas", Type: SurveyVarInt},
			{Name: "stage", Type: SurveyVarEnum, Values: []SurveyVarEnumValue{
				{Name: "Dev", Value: "dev"},
				{Name: "Prod", Value: "prod"},
			}},
			{Name: "notes", Type: SurveyVarText},
			{Name: "token", Type: SurveyVarSecret},
		},
	}
}

func TestTask_ValidateSurveyVars(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		secret      string
		errContains string
	}{
		{name: "no variables"},
		{name: "declared variables", environment: `{"app_version":"1.2.3","replicas":3,"stage":"prod","notes":"hi"}`},
		{name: "declared secret", secret: `{"token":"s3cret"}`},
		{name: "integer as string", environment: `{"replicas":"3"}`},
		{name: "empty value of typed var", environment: `{"replicas":"","stage":""}`},
		{name: "null value of typed var", environment: `{"replicas":null,"stage":null}`},

		{
			name:        "undeclared variable",
			environment: `{"ansible_ssh_common_args":"-o ProxyCommand=/bin/sh"}`,
			errContains: "ansible_ssh_common_args is not declared",
		},
		{
			name:        "undeclared variable next to a declared one",
			environment: `{"app_version":"1.2.3","ansible_connection":"local"}`,
			errContains: "ansible_connection is not declared",
		},
		{
			name:        "undeclared secret",
			secret:      `{"ansible_ssh_common_args":"-o ProxyCommand=/bin/sh"}`,
			errContains: "ansible_ssh_common_args is not declared",
		},
		{
			name:        "declared non secret variable sent as secret",
			secret:      `{"app_version":"1.2.3"}`,
			errContains: "app_version is not declared",
		},
		{
			name:        "secret variable sent as environment",
			environment: `{"token":"s3cret"}`,
			errContains: "token must be sent in the task secret",
		},
		{
			name:        "non integer value",
			environment: `{"replicas":"3; rm -rf /"}`,
			errContains: "replicas must be an integer",
		},
		{
			name:        "fractional value of integer var",
			environment: `{"replicas":1.5}`,
			errContains: "replicas must be an integer",
		},
		{
			name:        "value outside of enum",
			environment: `{"stage":"staging"}`,
			errContains: "stage has a value which is not allowed",
		},
		{
			name:        "environment is not an object",
			environment: `["app_version"]`,
			errContains: "task environment must be a JSON object",
		},
		{
			name:        "secret is not an object",
			secret:      `"token"`,
			errContains: "task secret must be a JSON object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Environment: tt.environment, Secret: tt.secret}

			err := task.ValidateSurveyVars(surveyTemplate())

			if tt.errContains == "" {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.errContains)
		})
	}
}

func TestTask_ValidateSurveyVars_TemplateWithoutSurvey(t *testing.T) {
	task := Task{Environment: `{"anything":"value"}`}

	err := task.ValidateSurveyVars(Template{})

	assert.ErrorContains(t, err, "anything is not declared")
}

func TestTask_ValidateSurveyVars_AllowAnyVarsInTask(t *testing.T) {
	task := Task{
		Environment: `{"ansible_ssh_common_args":"-o ProxyCommand=/bin/sh","replicas":"many"}`,
		Secret:      `{"undeclared":"value"}`,
	}

	tpl := surveyTemplate()
	tpl.AllowAnyVarsInTask = true

	assert.NoError(t, task.ValidateSurveyVars(tpl))

	// The very same task is rejected while the template does not opt in.
	tpl.AllowAnyVarsInTask = false
	assert.Error(t, task.ValidateSurveyVars(tpl))
}

func TestTemplate_GetSurveyVar(t *testing.T) {
	tpl := surveyTemplate()

	v := tpl.GetSurveyVar("stage")

	require.NotNil(t, v)
	assert.Equal(t, SurveyVarEnum, v.Type)
	assert.Nil(t, tpl.GetSurveyVar("STAGE"))
	assert.Nil(t, tpl.GetSurveyVar("unknown"))
}
