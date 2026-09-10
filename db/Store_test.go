package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectToJSON(t *testing.T) {
	v := &SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	assert.NotNil(t, s)
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\"}", *s)
}

func TestObjectToJSON2(t *testing.T) {
	var v *SurveyVar = nil
	s := ObjectToJSON(v)
	assert.Nil(t, s)
}

func TestObjectToJSON3(t *testing.T) {
	v := SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	assert.NotNil(t, s)
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\"}", *s)
}

func TestGetAccessKeyOptionsValidate(t *testing.T) {
	id := 1

	tests := []struct {
		name    string
		options GetAccessKeyOptions
		wantErr string
	}{
		{"ignore owner skips checks", GetAccessKeyOptions{IgnoreOwner: true}, ""},
		{"shared owner needs no id", GetAccessKeyOptions{Owner: AccessKeyShared}, ""},
		{"variable owner with environment id", GetAccessKeyOptions{Owner: AccessKeyVariable, EnvironmentID: &id}, ""},
		{"variable owner without environment id", GetAccessKeyOptions{Owner: AccessKeyVariable}, "environment_id is required"},
		{"environment owner without environment id", GetAccessKeyOptions{Owner: AccessKeyEnvironment}, "environment_id is required"},
		{"storage owner with storage id", GetAccessKeyOptions{Owner: AccessKeySecretStorage, StorageID: &id}, ""},
		{"storage owner without storage id", GetAccessKeyOptions{Owner: AccessKeySecretStorage}, "storage_id is required"},
		{"task owner with task id", GetAccessKeyOptions{Owner: AccessKeyTaskSecret, TaskID: &id}, ""},
		{"task owner without task id", GetAccessKeyOptions{Owner: AccessKeyTaskSecret}, "task_id is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
