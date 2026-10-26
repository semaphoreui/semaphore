package server

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFillSecretsService(t *testing.T, links []db.EnvironmentSecretWithKey) *accessKeyEncryptionServiceImpl {
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })

	repo := &mockEnvironmentManager{
		GetEnvironmentSecretKeysFn: func(projectID int, environmentID int) ([]db.EnvironmentSecretWithKey, error) {
			assert.Equal(t, 10, projectID)
			assert.Equal(t, 1, environmentID)
			return links, nil
		},
	}
	return &accessKeyEncryptionServiceImpl{environmentRepo: repo}
}

// storedStringKey is a string key as it sits in the database: the value encrypted into Secret.
func storedStringKey(t *testing.T, svc *accessKeyEncryptionServiceImpl, id int, name string, value string, envID *int) db.AccessKey {
	projectID := 10
	key := db.AccessKey{ID: id, Name: name, Type: db.AccessKeyString, ProjectID: &projectID, String: value, EnvironmentID: envID}
	if envID != nil {
		key.Owner = db.AccessKeyEnvironment
	}
	require.NoError(t, svc.SerializeSecret(&key))
	key.String = ""
	return key
}

func TestFillEnvironmentSecrets_WithoutValuesShowsTheLinks(t *testing.T) {
	envID := 1
	svc := newFillSecretsService(t, nil)
	owned := storedStringKey(t, svc, 5, "DB_PASSWORD", "p1", &envID)
	shared := storedStringKey(t, svc, 6, "shared", "p2", nil)
	svc.environmentRepo = &mockEnvironmentManager{
		GetEnvironmentSecretKeysFn: func(int, int) ([]db.EnvironmentSecretWithKey, error) {
			return []db.EnvironmentSecretWithKey{
				{EnvironmentSecret: db.EnvironmentSecret{ID: 1, ProjectID: 10, EnvironmentID: envID, Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", KeyID: 5, KeyName: "DB_PASSWORD", Owned: true}, Key: owned},
				{EnvironmentSecret: db.EnvironmentSecret{ID: 2, ProjectID: 10, EnvironmentID: envID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: 6, KeyName: "shared"}, Key: shared},
			}, nil
		},
	}
	env := db.Environment{ID: envID, ProjectID: 10, Name: "prod"}

	require.NoError(t, svc.FillEnvironmentSecrets(&env, false))

	assert.Equal(t, []db.EnvironmentSecret{
		{ID: 1, ProjectID: 10, EnvironmentID: envID, Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", KeyID: 5, KeyName: "DB_PASSWORD", Owned: true},
		{ID: 2, ProjectID: 10, EnvironmentID: envID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: 6, KeyName: "shared"},
	}, env.Secrets, "no value leaves the service")
}

func TestFillEnvironmentSecrets_WithValuesDecryptsOwnedAndLinkedKeys(t *testing.T) {
	envID := 1
	svc := newFillSecretsService(t, nil)
	owned := storedStringKey(t, svc, 5, "DB_PASSWORD", "p1", &envID)
	shared := storedStringKey(t, svc, 6, "shared", "p2", nil)
	svc.environmentRepo = &mockEnvironmentManager{
		GetEnvironmentSecretKeysFn: func(int, int) ([]db.EnvironmentSecretWithKey, error) {
			return []db.EnvironmentSecretWithKey{
				{EnvironmentSecret: db.EnvironmentSecret{ID: 1, Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", KeyID: 5}, Key: owned},
				{EnvironmentSecret: db.EnvironmentSecret{ID: 2, Type: db.EnvironmentSecretVar, Name: "token", KeyID: 6}, Key: shared},
				{EnvironmentSecret: db.EnvironmentSecret{ID: 3, Type: db.EnvironmentSecretEnv, Name: "TOKEN", KeyID: 6}, Key: shared},
			}, nil
		},
	}
	env := db.Environment{ID: envID, ProjectID: 10, Name: "prod"}

	require.NoError(t, svc.FillEnvironmentSecrets(&env, true))

	require.Len(t, env.Secrets, 3)
	assert.Equal(t, "p1", env.Secrets[0].Secret)
	assert.Equal(t, "p2", env.Secrets[1].Secret)
	assert.Equal(t, "p2", env.Secrets[2].Secret, "two secrets may point at one key")
	assert.Equal(t, db.EnvironmentSecretEnv, env.Secrets[2].Type)
	assert.Equal(t, "TOKEN", env.Secrets[2].Name)
}

func TestFillEnvironmentSecrets_KeyNoLongerStringFails(t *testing.T) {
	svc := newFillSecretsService(t, nil)
	projectID := 10
	changed := db.AccessKey{ID: 6, Name: "shared", Type: db.AccessKeyLoginPassword, ProjectID: &projectID}
	svc.environmentRepo = &mockEnvironmentManager{
		GetEnvironmentSecretKeysFn: func(int, int) ([]db.EnvironmentSecretWithKey, error) {
			return []db.EnvironmentSecretWithKey{
				{EnvironmentSecret: db.EnvironmentSecret{ID: 2, Type: db.EnvironmentSecretVar, Name: "token", KeyID: 6}, Key: changed},
			}, nil
		},
	}
	env := db.Environment{ID: 1, ProjectID: 10, Name: "prod"}

	err := svc.FillEnvironmentSecrets(&env, true)

	assert.ErrorContains(t, err, `secret "token" of variable group "prod"`)
	assert.ErrorContains(t, err, `key "shared" is login_password`)
	assert.Empty(t, env.Secrets)

	require.NoError(t, svc.FillEnvironmentSecrets(&env, false), "without values the link is still listed for the form")
	assert.Len(t, env.Secrets, 1)
}
