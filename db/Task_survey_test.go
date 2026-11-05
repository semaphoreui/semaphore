package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTaskSurveyValues(t *testing.T) {
	values := []SurveyVarEnumValue{{Name: "A", Value: "a"}, {Name: "B", Value: "b"}}
	vars := []SurveyVar{
		{Name: "str"},
		{Name: "num", Type: SurveyVarType(SurveyVarInt)},
		{Name: "env", Type: SurveyVarType(SurveyVarEnum), Values: values},
		{Name: "pwd", Type: surveyVarSecret},
	}

	tests := []struct {
		name    string
		env     string
		secret  string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"valid answers", `{"str":"x","num":"-12","env":"a"}`, `{"pwd":"p"}`, false},
		{"int as json number", `{"num":5}`, "", false},
		{"empty answers", `{"str":"","num":"","env":null,"pwd":""}`, "", false},
		{"undeclared var in environment", `{"other":"x"}`, "", true},
		{"undeclared var in secret", "", `{"ansible_ssh_common_args":"-o ProxyCommand=id"}`, true},
		{"semaphore_vars is undeclared", `{"semaphore_vars":{}}`, "", true},
		{"invalid environment json", `[1]`, "", true},
		{"invalid secret json", "", `x`, true},
		{"int not a number", `{"num":"1; rm -rf /"}`, "", true},
		{"int fractional", `{"num":1.5}`, "", true},
		{"enum not in values", `{"env":"c"}`, "", true},
		{"enum not a string", `{"env":["a"]}`, "", true},
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
