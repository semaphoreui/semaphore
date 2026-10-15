package db

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func surveyDefault(values ...string) *SurveyVarDefaultValue {
	return &SurveyVarDefaultValue{Values: values}
}

func TestValidateWorkflowSurveyVars(t *testing.T) {
	tests := []struct {
		name    string
		vars    []SurveyVar
		wantErr string
	}{
		{"no survey", nil, ""},
		{"string and int", []SurveyVar{
			{Name: "image_tag", Title: "Image tag", Required: true},
			{Name: "replicas", Title: "Replicas", Type: SurveyVarInt, DefaultValue: surveyDefault("3")},
		}, ""},
		{"env target", []SurveyVar{{Name: "region", Title: "Region", Target: SurveyVarTargetEnv}}, ""},
		{"enum with valid default", []SurveyVar{{
			Name: "env", Title: "Env", Type: SurveyVarEnum,
			Values:       []SurveyVarEnumValue{{Name: "Prod", Value: "prod"}, {Name: "Dev", Value: "dev"}},
			DefaultValue: surveyDefault("dev"),
		}}, ""},
		{"empty name", []SurveyVar{{Title: "No name"}}, "name can not be empty"},
		{"duplicate name", []SurveyVar{
			{Name: "env", Title: "Env"},
			{Name: "env", Title: "Env again"},
		}, "declared more than once"},
		{"secret rejected", []SurveyVar{{Name: "token", Title: "Token", Type: SurveyVarSecret}}, "can not be a secret"},
		{"unknown target", []SurveyVar{{Name: "x", Title: "X", Target: SurveyVarTarget("file")}}, "invalid target"},
		{"enum default outside values", []SurveyVar{{
			Name: "env", Title: "Env", Type: SurveyVarEnum,
			Values:       []SurveyVarEnumValue{{Name: "Prod", Value: "prod"}},
			DefaultValue: surveyDefault("dev"),
		}}, "is not in values list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflowSurveyVars(tt.vars)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			var validation *common_errors.ValidationError
			assert.ErrorAs(t, err, &validation)
		})
	}
}

func TestWorkflowTemplate_SurveyVarByName(t *testing.T) {
	workflow := WorkflowTemplate{SurveyVars: []SurveyVar{
		{Name: "env", Title: "Env"},
		{Name: "image_tag", Title: "Image tag", Type: SurveyVarText},
	}}

	v, ok := workflow.SurveyVarByName("image_tag")
	require.True(t, ok)
	assert.Equal(t, SurveyVarText, v.Type)

	_, ok = workflow.SurveyVarByName("missing")
	assert.False(t, ok)

	_, ok = WorkflowTemplate{}.SurveyVarByName("env")
	assert.False(t, ok)
}

// The wire shape of the new fields: survey_vars is always present on a
// template (an empty list, never null, once the store fills it), a node
// carries survey_inputs only when it has bindings, the JSON mirrors of the
// DB columns never leak.
func TestWorkflowSurveyJSONShape(t *testing.T) {
	raw := `{"id":1,"kind":"task","template_id":5,"survey_inputs":[{"var":"image_tag","workflow_var":"tag"}],"position_x":0,"position_y":0}`
	var node WorkflowNode
	require.NoError(t, json.Unmarshal([]byte(raw), &node))
	require.Len(t, node.SurveyInputs, 1)
	assert.Equal(t, WorkflowNodeSurveyInput{Var: "image_tag", WorkflowVar: "tag"}, node.SurveyInputs[0])

	stored := "[]"
	node.SurveyInputsJSON = &stored
	out, err := json.Marshal(node)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"survey_inputs":[{"var":"image_tag","workflow_var":"tag"}]`)
	assert.NotContains(t, string(out), "SurveyInputsJSON")
	assert.NotContains(t, string(out), `"survey_inputs_json"`)

	bare, err := json.Marshal(WorkflowNode{ID: 2, Kind: WorkflowNodeApprovalKind})
	require.NoError(t, err)
	assert.NotContains(t, string(bare), "survey_inputs")

	tpl, err := json.Marshal(WorkflowTemplate{SurveyVars: []SurveyVar{}})
	require.NoError(t, err)
	assert.Contains(t, string(tpl), `"survey_vars":[]`)

	revision, err := json.Marshal(WorkflowRevision{SurveyVarsJSON: &stored})
	require.NoError(t, err)
	assert.NotContains(t, string(revision), "survey_vars")

	// The run's environment is an object on the wire and JSON text in the
	// database column.
	run, err := json.Marshal(WorkflowRun{Environment: MapStringAnyField{"tag": "1.2.3", "regions": []string{"eu"}}})
	require.NoError(t, err)
	assert.Contains(t, string(run), `"environment":{"regions":["eu"],"tag":"1.2.3"}`)

	bareRun, err := json.Marshal(WorkflowRun{})
	require.NoError(t, err)
	assert.NotContains(t, string(bareRun), "environment")

	column, err := WorkflowRun{Environment: MapStringAnyField{"tag": "1.2.3"}}.Environment.Value()
	require.NoError(t, err)
	assert.JSONEq(t, `{"tag":"1.2.3"}`, string(column.([]byte)))
	column, err = WorkflowRun{}.Environment.Value()
	require.NoError(t, err)
	assert.Nil(t, column, "no survey is NULL, not an empty object")

	var params WorkflowRunParams
	require.NoError(t, json.Unmarshal([]byte(`{"environment":{"tag":"x","count":2}}`), &params))
	assert.Equal(t, "x", params.Environment["tag"])
	assert.Equal(t, float64(2), params.Environment["count"])
	var empty WorkflowRunParams
	require.NoError(t, json.Unmarshal([]byte(`{}`), &empty))
	assert.Nil(t, empty.Environment)
}
