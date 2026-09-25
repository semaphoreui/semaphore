package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"
)

// fakeEncryptionService implements only DeserializeSecret; calling any other
// method of the embedded nil interface panics, which the tests must not do.
type fakeEncryptionService struct {
	server.AccessKeyEncryptionService
	err error
}

func (f *fakeEncryptionService) DeserializeSecret(key *db.AccessKey) error {
	if f.err != nil {
		return f.err
	}
	key.LoginPassword.Password = "decrypted-token"
	return nil
}

func TestRepositoryController_DecryptRepositoryKey(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedOK     bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "key is decrypted in place",
			expectedOK: true,
		},
		{
			name:           "expired key is reported to the user",
			err:            server.ErrAccessKeyExpired,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "The repository's access key has expired",
		},
		{
			name:           "other failures are not shown to the user",
			err:            errors.New("vault at https://vault.internal:8200 is sealed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to read the repository's access key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewRepositoryController(nil, &fakeEncryptionService{err: tt.err})
			repo := db.Repository{
				ID:       1,
				SSHKeyID: 2,
				SSHKey:   db.AccessKey{Type: db.AccessKeyLoginPassword},
			}
			w := httptest.NewRecorder()

			ok := c.decryptRepositoryKey(w, &repo)

			assert.Equal(t, tt.expectedOK, ok)
			if tt.expectedOK {
				assert.Equal(t, "decrypted-token", repo.SSHKey.LoginPassword.Password)
				assert.Equal(t, 0, w.Body.Len(), "nothing may be written on success")
				return
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			var body map[string]string
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, tt.expectedError, body["error"])
			assert.NotContains(t, w.Body.String(), "vault.internal")
		})
	}
}

func TestWriteRepositoryError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "git command error is shown to the user",
			err: &db_lib.GitCommandError{
				Command: "ls-remote",
				Detail:  "fatal: Authentication failed for 'https://gitlab.com/group/repo.git/'",
				Err:     errors.New("exit status 128"),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "git ls-remote: fatal: Authentication failed for 'https://gitlab.com/group/repo.git/'",
		},
		{
			name: "wrapped git command error is shown to the user",
			err: fmt.Errorf("pull: %w", &db_lib.GitCommandError{
				Command: "pull",
				Detail:  "fatal: couldn't find remote ref main",
				Err:     errors.New("exit status 1"),
			}),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "git pull: fatal: couldn't find remote ref main",
		},
		{
			name:           "other errors are not shown to the user",
			err:            errors.New("mkdir /var/lib/semaphore/project_1/repository_2: permission denied"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to load repository files",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			writeRepositoryError(w, tt.err, "Failed to load repository files", log.Fields{"repository_id": 2})

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body map[string]string
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, tt.expectedError, body["error"])
			assert.NotContains(t, w.Body.String(), "/var/lib/semaphore")
		})
	}
}
