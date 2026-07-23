package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/semaphoreui/semaphore/db"
)

// Task survey secrets are stored in the access_key table as keys with owner
// AccessKeyTaskSecret: bound to the task via task_id, hidden from the project
// secrets list (non-empty owner), encrypted with the shared keyring like any
// other secret, and unusable after ExpireAt. This keeps them readable by any
// HA node serving the runner dispatch while never persisting plaintext.

// ErrTaskSurveySecretsNotFound is returned when a task has no stored survey-
// secret key. Callers that require secrets must treat this as a dispatch
// failure; callers dispatching tasks that never had survey secrets may ignore it.
var ErrTaskSurveySecretsNotFound = errors.New("task survey secrets not found")

// CreateTaskSurveySecrets stores the survey-secrets JSON of a task as an
// encrypted task-bound access key. secrets must be a JSON object string.
func (s *accessKeyEncryptionServiceImpl) CreateTaskSurveySecrets(
	projectID int,
	taskID int,
	secrets string,
	expireAt time.Time,
) error {
	key := db.AccessKey{
		Name:           fmt.Sprintf("task-%d-survey-secrets", taskID),
		Type:           db.AccessKeyString,
		ProjectID:      &projectID,
		Owner:          db.AccessKeyTaskSecret,
		TaskID:         &taskID,
		ExpireAt:       &expireAt,
		String:         secrets,
		OverrideSecret: true,
	}

	if err := s.SerializeSecret(&key); err != nil {
		return err
	}

	_, err := s.accessKeyRepo.CreateAccessKey(key)
	return err
}

// GetTaskSurveySecrets returns the decrypted survey-secrets JSON of a task.
// A task without survey secrets yields ErrTaskSurveySecretsNotFound; an
// expired secret yields ErrAccessKeyExpired.
func (s *accessKeyEncryptionServiceImpl) GetTaskSurveySecrets(projectID int, taskID int) (string, error) {
	key, err := s.accessKeyRepo.GetTaskAccessKey(projectID, taskID)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return "", ErrTaskSurveySecretsNotFound
		}
		return "", err
	}

	if err = s.DeserializeSecret(&key); err != nil {
		return "", err
	}

	return key.String, nil
}

// DeleteTaskSurveySecrets removes the task-bound secret once the task has
// reached a terminal state. Missing keys are not an error.
func (s *accessKeyEncryptionServiceImpl) DeleteTaskSurveySecrets(projectID int, taskID int) error {
	return s.accessKeyRepo.DeleteTaskAccessKeys(projectID, taskID)
}
