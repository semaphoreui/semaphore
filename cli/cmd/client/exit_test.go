package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/stretchr/testify/assert"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, ExitOK},
		{"plain error", errors.New("boom"), ExitError},
		{"usage", usage("semaphore client x", "bad"), ExitUsage},
		{"wrapped usage", fmt.Errorf("ctx: %w", usage("", "bad")), ExitUsage},
		{"task error", &TaskFinishedError{TaskID: 1, Status: "error"}, ExitTaskFailed},
		{"task stopped", &TaskFinishedError{TaskID: 1, Status: "stopped"}, ExitTaskStopped},
		{"unauthorized", &apiclient.APIError{Status: 401}, ExitAuth},
		{"forbidden", &apiclient.APIError{Status: 403}, ExitAuth},
		{"not found", &apiclient.APIError{Status: 404}, ExitNotFound},
		{"ambiguous", &AmbiguousError{Kind: "template", Name: "x", Candidates: []string{"1", "2"}}, ExitNotFound},
		{"server error", &apiclient.APIError{Status: 500}, ExitError},
		{"transport", &apiclient.TransportError{Op: "GET", Err: errors.New("refused")}, ExitError},
		{"timeout", &WaitTimeoutError{TaskID: 3}, ExitTimeout},
		{"interrupt", context.Canceled, ExitInterrupted},
		{"wrapped interrupt", fmt.Errorf("wait: %w", context.Canceled), ExitInterrupted},
		{"no url", config.ErrNoURL, ExitUsage},
		{"unknown context", fmt.Errorf("%w %q", config.ErrUnknownContext, "x"), ExitUsage},
		{"token without url", config.ErrTokenWithoutURL, ExitUsage},
		{"bad output", config.ErrBadOutput, ExitUsage},
		{"aborted", ErrAborted, ExitError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExitCode(tt.err))
		})
	}
}

func TestUsageError_Message(t *testing.T) {
	assert.Equal(t, "bad flag\nRun 'semaphore client info --help' for usage.", usage("semaphore client info", "bad %s", "flag").Error())
	assert.Equal(t, "bad", usage("", "bad").Error())
	assert.Equal(t, `template name "web" matches several objects (3, 7); use the id`, (&AmbiguousError{Kind: "template", Name: "web", Candidates: []string{"3", "7"}}).Error())
}
