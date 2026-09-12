package server

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// taskSecretRepoMock overrides only the methods used by the task-secret
// service; any other call panics via the embedded nil interface.
type taskSecretRepoMock struct {
	db.AccessKeyManager

	created *db.AccessKey
	stored  *db.AccessKey
	deleted bool
}

func (m *taskSecretRepoMock) CreateAccessKey(key db.AccessKey) (db.AccessKey, error) {
	m.created = &key
	return key, nil
}

func (m *taskSecretRepoMock) GetTaskAccessKey(int, int) (db.AccessKey, error) {
	if m.stored == nil {
		return db.AccessKey{}, db.ErrNotFound
	}
	return *m.stored, nil
}

func (m *taskSecretRepoMock) DeleteTaskAccessKeys(int, int) error {
	m.deleted = true
	return nil
}

func TestCreateTaskSurveySecrets(t *testing.T) {
	util.Config = &util.ConfigType{}
	repo := &taskSecretRepoMock{}
	svc := NewAccessKeyEncryptionService(repo, nil, nil, nil)

	expireAt := tz.Now().Add(time.Hour)
	err := svc.CreateTaskSurveySecrets(1, 42, `{"passwd":"123456"}`, expireAt)

	require.NoError(t, err)
	require.NotNil(t, repo.created)
	assert.Equal(t, db.AccessKeyTaskSecret, repo.created.Owner)
	assert.Equal(t, db.AccessKeyString, repo.created.Type)
	assert.Equal(t, "task-42-survey-secrets", repo.created.Name)
	require.NotNil(t, repo.created.TaskID)
	assert.Equal(t, 42, *repo.created.TaskID)
	require.NotNil(t, repo.created.ExpireAt)
	assert.Equal(t, expireAt, *repo.created.ExpireAt)

	// Without an encryption key configured the secret is base64-encoded.
	require.NotNil(t, repo.created.Secret)
	secret, err := base64.StdEncoding.DecodeString(*repo.created.Secret)
	require.NoError(t, err)
	assert.Equal(t, `{"passwd":"123456"}`, string(secret))
}

func TestGetTaskSurveySecrets(t *testing.T) {
	util.Config = &util.ConfigType{}
	expireAt := tz.Now().Add(time.Hour)
	repo := &taskSecretRepoMock{
		stored: &db.AccessKey{
			Type:     db.AccessKeyString,
			Owner:    db.AccessKeyTaskSecret,
			ExpireAt: &expireAt,
			Secret:   new(base64.StdEncoding.EncodeToString([]byte(`{"passwd":"123456"}`))),
		},
	}
	svc := NewAccessKeyEncryptionService(repo, nil, nil, nil)

	secrets, err := svc.GetTaskSurveySecrets(1, 42)

	require.NoError(t, err)
	assert.Equal(t, `{"passwd":"123456"}`, secrets)
}

func TestGetTaskSurveySecrets_NotFound(t *testing.T) {
	util.Config = &util.ConfigType{}
	svc := NewAccessKeyEncryptionService(&taskSecretRepoMock{}, nil, nil, nil)

	secrets, err := svc.GetTaskSurveySecrets(1, 42)

	require.NoError(t, err)
	assert.Empty(t, secrets)
}

func TestGetTaskSurveySecrets_Expired(t *testing.T) {
	util.Config = &util.ConfigType{}
	expireAt := tz.Now().Add(-time.Minute)
	repo := &taskSecretRepoMock{
		stored: &db.AccessKey{
			Type:     db.AccessKeyString,
			Owner:    db.AccessKeyTaskSecret,
			ExpireAt: &expireAt,
			Secret:   new(base64.StdEncoding.EncodeToString([]byte(`{"passwd":"123456"}`))),
		},
	}
	svc := NewAccessKeyEncryptionService(repo, nil, nil, nil)

	_, err := svc.GetTaskSurveySecrets(1, 42)

	assert.ErrorIs(t, err, ErrAccessKeyExpired)
}

func TestDeleteTaskSurveySecrets(t *testing.T) {
	repo := &taskSecretRepoMock{}
	svc := NewAccessKeyEncryptionService(repo, nil, nil, nil)

	err := svc.DeleteTaskSurveySecrets(1, 42)

	require.NoError(t, err)
	assert.True(t, repo.deleted)
}

func TestDeserializeSecret_ExpireAt(t *testing.T) {
	util.Config = &util.ConfigType{}
	svc := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	past := tz.Now().Add(-time.Minute)
	future := tz.Now().Add(time.Minute)
	secret := base64.StdEncoding.EncodeToString([]byte("value"))

	tests := []struct {
		name     string
		expireAt *time.Time
		expired  bool
	}{
		{"no expiry", nil, false},
		{"future expiry", &future, false},
		{"past expiry", &past, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := db.AccessKey{
				Type:     db.AccessKeyString,
				ExpireAt: tt.expireAt,
				Secret:   &secret,
			}

			err := svc.DeserializeSecret(&key)

			if tt.expired {
				assert.ErrorIs(t, err, ErrAccessKeyExpired)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "value", key.String)
			}
		})
	}
}
