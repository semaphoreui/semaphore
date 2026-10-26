package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EnvironmentValidate_EmptyName_ReturnsError(t *testing.T) {
	env := &Environment{
		Name: "",
		JSON: "{}",
		ENV:  nil,
	}
	err := env.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Environment name can not be empty", err.Error())
}

func Test_EnvironmentValidate_InvalidJSON_ReturnsError(t *testing.T) {
	env := &Environment{
		Name: "TestEnv",
		JSON: "{invalid_json}",
		ENV:  nil,
	}
	err := env.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Extra variables must be valid JSON", err.Error())
}

func Test_EnvironmentValidate_ValidJSON_ReturnsNoError(t *testing.T) {
	env := &Environment{
		Name: "TestEnv",
		JSON: `{"key": "value"}`,
		ENV:  nil,
	}
	err := env.Validate()
	assert.NoError(t, err)
}

func Test_EnvironmentValidate_InvalidEnvJSON_ReturnsError(t *testing.T) {
	envVar := "{invalid_json}"
	env := &Environment{
		Name: "TestEnv",
		JSON: `{"key": "value"}`,
		ENV:  &envVar,
	}
	err := env.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Environment variables must be valid JSON", err.Error())
}

func Test_EnvironmentValidate_EmptyJsonName_ReturnsError(t *testing.T) {
	env := &Environment{
		Name: "TestEnv",
		JSON: `{"": "value"}`,
		ENV:  nil,
	}
	err := env.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Extra variables key can not be empty", err.Error())
}

func Test_EnvironmentValidate_NonScalarEnvValues_ReturnsError(t *testing.T) {
	envVar := `{"key": {"nested": "value"}}`
	env := &Environment{
		Name: "TestEnv",
		JSON: `{"key": "value"}`,
		ENV:  &envVar,
	}
	err := env.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Environment variables values must be scalar", err.Error())
}

func Test_EnvironmentValidate_ValidEnvJSON_ReturnsNoError(t *testing.T) {
	envVar := `{"key": "value"}`
	env := &Environment{
		Name: "TestEnv",
		JSON: `{"key": "value"}`,
		ENV:  &envVar,
	}
	err := env.Validate()
	assert.NoError(t, err)
}

func TestEnvironmentSecret_Validate(t *testing.T) {
	tests := []struct {
		name    string
		secret  EnvironmentSecret
		wantErr string
	}{
		{"create with value", EnvironmentSecret{Type: EnvironmentSecretVar, Name: "a", Secret: "v", Operation: EnvironmentSecretCreate}, ""},
		{"create with key", EnvironmentSecret{Type: EnvironmentSecretEnv, Name: "A", KeyID: 3, Operation: EnvironmentSecretCreate}, ""},
		{"create with neither", EnvironmentSecret{Type: EnvironmentSecretVar, Name: "a", Operation: EnvironmentSecretCreate}, "either a value or a key"},
		{"create with both", EnvironmentSecret{Type: EnvironmentSecretVar, Name: "a", Secret: "v", KeyID: 3, Operation: EnvironmentSecretCreate}, "can not be set together"},
		{"create without name", EnvironmentSecret{Type: EnvironmentSecretVar, Secret: "v", Operation: EnvironmentSecretCreate}, "name can not be empty"},
		{"update keeps value", EnvironmentSecret{ID: 1, Type: EnvironmentSecretVar, Name: "a", Operation: EnvironmentSecretUpdate}, ""},
		{"update with both", EnvironmentSecret{ID: 1, Type: EnvironmentSecretVar, Name: "a", Secret: "v", KeyID: 3, Operation: EnvironmentSecretUpdate}, "can not be set together"},
		{"delete needs nothing", EnvironmentSecret{ID: 1, Type: EnvironmentSecretVar, Operation: EnvironmentSecretDelete}, ""},
		{"bad type", EnvironmentSecret{Type: "file", Name: "a", Secret: "v", Operation: EnvironmentSecretCreate}, "invalid environment secret type"},
		{"bad operation", EnvironmentSecret{Type: EnvironmentSecretVar, Name: "a", Secret: "v", Operation: "upsert"}, "invalid environment secret operation"},
		{"no operation", EnvironmentSecret{Type: EnvironmentSecretVar, Name: "a", Secret: "v"}, "invalid environment secret operation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.secret.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateEnvironmentSecretKey(t *testing.T) {
	taskID := 7
	tests := []struct {
		name    string
		key     AccessKey
		wantErr string
	}{
		{"shared string", AccessKey{Name: "k", Type: AccessKeyString}, ""},
		{"owned by a group (var)", AccessKey{Name: "k", Type: AccessKeyString, Owner: AccessKeyVariable}, ""},
		{"owned by a group (env)", AccessKey{Name: "k", Type: AccessKeyString, Owner: AccessKeyEnvironment}, ""},
		{"ssh", AccessKey{Name: "k", Type: AccessKeySSH}, "only string keys"},
		{"login_password", AccessKey{Name: "k", Type: AccessKeyLoginPassword}, "only string keys"},
		{"none", AccessKey{Name: "k", Type: AccessKeyNone}, "only string keys"},
		{"task survey secret", AccessKey{Name: "k", Type: AccessKeyString, Owner: AccessKeyTaskSecret, TaskID: &taskID}, "can not be used"},
		{"secret storage key", AccessKey{Name: "k", Type: AccessKeyString, Owner: AccessKeySecretStorage}, "can not be used"},
		{"string bound to a task", AccessKey{Name: "k", Type: AccessKeyString, TaskID: &taskID}, "belongs to a task"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironmentSecretKey(tt.key)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
