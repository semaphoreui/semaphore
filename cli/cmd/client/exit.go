package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
)

// Exit codes of `semaphore client`. They let a script tell a CLI failure from
// a task that ran and failed, the way gh's --exit-status and coreutils'
// timeout do.
const (
	ExitOK          = 0
	ExitError       = 1   // the client or the API failed
	ExitUsage       = 2   // wrong flags or arguments, or a prompt with no TTY
	ExitTaskFailed  = 3   // the awaited task finished with status "error"
	ExitAuth        = 4   // 401 or 403
	ExitNotFound    = 5   // 404 or an ambiguous name
	ExitTaskStopped = 6   // the awaited task finished with status "stopped"
	ExitTimeout     = 124 // --timeout elapsed while waiting
	ExitInterrupted = 130 // Ctrl+C
)

// UsageError is a mistake in how the command was called. Cobra reports every
// error with code 1, so the client wraps argument and flag problems in this
// type to exit with 2 and append a help hint.
type UsageError struct {
	Msg  string
	Hint string
}

func (e *UsageError) Error() string {
	if e.Hint == "" {
		return e.Msg
	}
	return e.Msg + "\n" + e.Hint
}

// usage builds a UsageError with a "--help" hint for the given command path.
func usage(cmdPath, format string, args ...any) *UsageError {
	hint := ""
	if cmdPath != "" {
		hint = fmt.Sprintf("Run '%s --help' for usage.", cmdPath)
	}
	return &UsageError{Msg: fmt.Sprintf(format, args...), Hint: hint}
}

// TaskFinishedError says the task the command waited for ended without
// success. Status is the final task status (error, stopped).
type TaskFinishedError struct {
	TaskID int
	Status string
}

func (e *TaskFinishedError) Error() string {
	return fmt.Sprintf("task %d finished with status %s", e.TaskID, e.Status)
}

// WaitTimeoutError says --timeout elapsed; the task keeps running on the server.
type WaitTimeoutError struct {
	TaskID int
}

func (e *WaitTimeoutError) Error() string {
	return fmt.Sprintf("timed out waiting for task %d; it is still running on the server", e.TaskID)
}

// AmbiguousError says a name matched several objects, so an id is needed.
type AmbiguousError struct {
	Kind       string
	Name       string
	Candidates []string
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("%s name %q matches several objects (%s); use the id", e.Kind, e.Name, strings.Join(e.Candidates, ", "))
}

// ErrAborted is returned when the user answers "no" to a confirmation.
var ErrAborted = errors.New("aborted")

// ExitCode maps an error to the process exit code. nil is ExitOK.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}

	var usageErr *UsageError
	var finished *TaskFinishedError
	var timeout *WaitTimeoutError
	var ambiguous *AmbiguousError

	switch {
	case errors.As(err, &usageErr):
		return ExitUsage
	case errors.As(err, &timeout):
		return ExitTimeout
	case errors.As(err, &finished):
		if finished.Status == "error" {
			return ExitTaskFailed
		}
		return ExitTaskStopped
	case errors.Is(err, apiclient.ErrUnauthorized):
		return ExitAuth
	case errors.Is(err, apiclient.ErrNotFound), errors.As(err, &ambiguous):
		return ExitNotFound
	case errors.Is(err, context.Canceled):
		return ExitInterrupted
	case errors.Is(err, config.ErrNoURL), errors.Is(err, config.ErrUnknownContext),
		errors.Is(err, config.ErrTokenWithoutURL), errors.Is(err, config.ErrTokenAndTokenFile),
		errors.Is(err, config.ErrBadOutput):
		return ExitUsage
	}
	return ExitError
}
