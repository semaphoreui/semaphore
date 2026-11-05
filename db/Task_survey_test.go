package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTaskSurveyValues(t *testing.T) {
	values := []SurveyVarEnumValue{{Name: "A", Value: "a"}, {Name: "B", Value: "b"}}
	vars := []SurveyVar{
		{Name: "str"},
		{Name: "num", Type: SurveyVarInt},
		{Name: "env", Type: SurveyVarEnum, Values: values},
		{Name: "sel", Type: SurveyVarSelect, Values: values},
		{Name: "pwd", Type: SurveyVarSecret},
	}

	tests := []struct {
		name    string
		env     string
		secret  string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"valid answers", `{"str":"x","num":"-12","env":"a","sel":["a","b"]}`, `{"pwd":"p"}`, false},
		{"int as json number", `{"num":5}`, "", false},
		{"empty answers", `{"str":"","num":"","env":null,"sel":[],"pwd":""}`, "", false},
		{"legacy scalar select", `{"sel":"b"}`, "", false},
		{"undeclared var ignored", `{"other":[1,2]}`, "", false},
		{"invalid environment json", `[1]`, "", true},
		{"invalid secret json", "", `x`, true},
		{"int not a number", `{"num":"1; rm -rf /"}`, "", true},
		{"int fractional", `{"num":1.5}`, "", true},
		{"enum not in values", `{"env":"c"}`, "", true},
		{"enum not a string", `{"env":["a"]}`, "", true},
		{"select item not in values", `{"sel":["a","c"]}`, "", true},
		{"string is an object", `{"str":{"a":1}}`, "", true},
		{"secret in environment", `{"pwd":"p"}`, "", true},
		{"enum bypass via secret", "", `{"env":"c"}`, true},
		{"int bypass via secret", "", `{"num":"abc"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskSurveyValues(vars, tt.env, tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
