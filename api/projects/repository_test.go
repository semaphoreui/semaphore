package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"
)

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
