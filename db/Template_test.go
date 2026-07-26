package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSurveyVarTarget_JSONRoundTrip(t *testing.T) {
	v := SurveyVar{Name: "MY_VAR", Title: "My Var", Target: SurveyVarTargetEnv}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"target":"env"`)

	// Default target must be omitted for backward compatibility.
	v.Target = SurveyVarTargetDefault
	data, err = json.Marshal(v)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "target")

	var parsed SurveyVar
	require.NoError(t, json.Unmarshal([]byte(`{"name":"X","target":"env"}`), &parsed))
	assert.Equal(t, SurveyVarTargetEnv, parsed.Target)
}

func TestTemplateValidate_SurveyVarTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  SurveyVarTarget
		wantErr bool
	}{
		{"default target is valid", SurveyVarTargetDefault, false},
		{"env target is valid", SurveyVarTargetEnv, false},
		{"unknown target is rejected", SurveyVarTarget("bogus"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := Template{
				Name:     "test",
				Playbook: "playbook.yml",
				// App left empty ("") so Validate skips the util.Config.Apps whitelist check.
				SurveyVars: []SurveyVar{{Name: "V", Title: "V", Target: tt.target}},
			}
			err := tpl.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, "invalid survey variable target")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
