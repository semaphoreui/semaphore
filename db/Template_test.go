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

func TestTemplateNormalizedExecutorImage(t *testing.T) {
	t.Run("no override is stored as NULL", func(t *testing.T) {
		for _, image := range []*string{nil, new(""), new("   ")} {
			tpl := Template{ExecutorImage: image}
			assert.Nil(t, tpl.NormalizedExecutorImage())
		}
	})

	t.Run("override is stored trimmed", func(t *testing.T) {
		tpl := Template{ExecutorImage: new(" my-registry/job:1 ")}
		require.NotNil(t, tpl.NormalizedExecutorImage())
		assert.Equal(t, "my-registry/job:1", *tpl.NormalizedExecutorImage())
	})
}

func TestSurveyVarDefaultValue_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantArray bool
		wantVals  []string
		wantOut   string
	}{
		{"null", `null`, false, nil, `null`},
		{"scalar string", `"x"`, false, []string{"x"}, `"x"`},
		{"empty string", `""`, false, []string{""}, `""`},
		{"single-element array", `["x"]`, true, []string{"x"}, `["x"]`},
		{"multi-element array", `["x","y"]`, true, []string{"x", "y"}, `["x","y"]`},
		{"empty array", `[]`, true, []string{}, `[]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d SurveyVarDefaultValue
			require.NoError(t, json.Unmarshal([]byte(tt.input), &d))
			assert.Equal(t, tt.wantArray, d.IsArray())
			assert.Equal(t, tt.wantVals, d.Values)

			out, err := json.Marshal(d)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantOut, string(out))
		})
	}
}

func TestSurveyVarDefaultValue_JSONRoundTrip_RejectsNonString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"number", `42`},
		{"array of numbers", `[1,2,3]`},
		{"object", `{"k":"v"}`},
		{"bool", `true`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d SurveyVarDefaultValue
			assert.Error(t, json.Unmarshal([]byte(tt.input), &d))
		})
	}
}

func TestValidateSurveyVar(t *testing.T) {
	strVal := func(s string) *SurveyVarDefaultValue {
		return &SurveyVarDefaultValue{Values: []string{s}}
	}
	arrVal := func(s ...string) *SurveyVarDefaultValue {
		return &SurveyVarDefaultValue{Values: s, originalWasArray: true}
	}
	enumValues := []SurveyVarEnumValue{{Name: "A", Value: "a"}, {Name: "B", Value: "b"}}

	tests := []struct {
		name    string
		v       SurveyVar
		wantErr string // substring; empty = no error
	}{
		// --- select type ---
		{"select: nil default ok", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues}, ""},
		{"select: empty array ok", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: arrVal()}, ""},
		{"select: single value in list ok", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: arrVal("a")}, ""},
		{"select: multi values in list ok", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: arrVal("a", "b")}, ""},
		{"select: legacy scalar ok (backward compat)", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: strVal("a")}, ""},
		{"select: value not in list rejected", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: arrVal("zzz")}, "not in values list"},
		{"select: one of multi not in list rejected", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: arrVal("a", "zzz")}, "not in values list"},
		{"select: legacy multi-scalar rejected", SurveyVar{Name: "V", Type: SurveyVarSelect, Values: enumValues, DefaultValue: &SurveyVarDefaultValue{Values: []string{"a", "b"}}}, "must be an array"},

		// --- enum type ---
		{"enum: nil default ok", SurveyVar{Name: "V", Type: SurveyVarEnum, Values: enumValues}, ""},
		{"enum: scalar in list ok", SurveyVar{Name: "V", Type: SurveyVarEnum, Values: enumValues, DefaultValue: strVal("a")}, ""},
		{"enum: scalar not in list rejected", SurveyVar{Name: "V", Type: SurveyVarEnum, Values: enumValues, DefaultValue: strVal("zzz")}, "not in values list"},
		{"enum: array of one ok (lenient)", SurveyVar{Name: "V", Type: SurveyVarEnum, Values: enumValues, DefaultValue: arrVal("a")}, ""},
		{"enum: array of multi rejected", SurveyVar{Name: "V", Type: SurveyVarEnum, Values: enumValues, DefaultValue: arrVal("a", "b")}, "must be a string for enum"},

		// --- string / int / text / secret types ---
		{"string: nil default ok", SurveyVar{Name: "V", Type: SurveyVarStr}, ""},
		{"string: scalar ok", SurveyVar{Name: "V", Type: SurveyVarStr, DefaultValue: strVal("hello")}, ""},
		{"string: array of one ok (lenient)", SurveyVar{Name: "V", Type: SurveyVarStr, DefaultValue: arrVal("hello")}, ""},
		{"string: array of multi rejected", SurveyVar{Name: "V", Type: SurveyVarStr, DefaultValue: arrVal("a", "b")}, "must be a string for type"},
		{"int: scalar ok", SurveyVar{Name: "V", Type: SurveyVarInt, DefaultValue: strVal("42")}, ""},
		{"int: array of multi rejected", SurveyVar{Name: "V", Type: SurveyVarInt, DefaultValue: arrVal("1", "2")}, "must be a string for type"},
		{"text: scalar ok", SurveyVar{Name: "V", Type: SurveyVarText, DefaultValue: strVal("long text")}, ""},
		{"secret: scalar ok", SurveyVar{Name: "V", Type: SurveyVarText, DefaultValue: strVal("s3cr3t")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSurveyVar(tt.v)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestTemplateValidate_SurveyVarDefaultValue(t *testing.T) {
	// Validate() must surface ValidateSurveyVar errors end-to-end.
	tpl := Template{
		Name:       "test",
		Playbook:   "playbook.yml",
		SurveyVars: []SurveyVar{
			{
				Name:         "BAD",
				Title:        "Bad",
				Type:         SurveyVarStr,
				DefaultValue: &SurveyVarDefaultValue{Values: []string{"a", "b"}, originalWasArray: true},
			},
		},
	}
	err := tpl.Validate()
	assert.ErrorContains(t, err, "must be a string for type")
}

func strPtr(s string) *string {
	return &s
}
