package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Sentinel errors a caller can test with errors.Is against any error the
// client returns; an *APIError matches them by status code.
var (
	// ErrUnauthorized is a 401 or 403: the token is missing, expired or lacks
	// the permission.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrNotFound is a 404.
	ErrNotFound = errors.New("not found")
)

// APIError is a non-2xx response from the server.
type APIError struct {
	Status  int
	Method  string
	Path    string
	Message string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("%s %s: server returned %d", e.Method, e.Path, e.Status)
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return msg
}

// Is lets errors.Is(err, ErrUnauthorized) and errors.Is(err, ErrNotFound)
// work on an *APIError.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden
	case ErrNotFound:
		return e.Status == http.StatusNotFound
	}
	return false
}

// TransportError is a failure before any HTTP status arrived: DNS, TCP, TLS,
// a timeout or a refused redirect.
type TransportError struct {
	Op  string
	Err error
}

func (e *TransportError) Error() string { return e.Op + ": " + e.Err.Error() }
func (e *TransportError) Unwrap() error { return e.Err }

// newAPIError reads the body of a failed response. Semaphore writes errors as
// {"error": "..."}; anything else is kept as plain text, trimmed, so that a
// proxy's HTML page does not flood the terminal.
func newAPIError(resp *http.Response) *APIError {
	apiErr := &APIError{Status: resp.StatusCode}
	if resp.Request != nil {
		apiErr.Method = resp.Request.Method
		apiErr.Path = resp.Request.URL.Path
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil || len(body) == 0 {
		return apiErr
	}

	var errRes struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &errRes) == nil && errRes.Error != "" {
		apiErr.Message = errRes.Error
		return apiErr
	}

	text := strings.TrimSpace(string(body))
	if strings.HasPrefix(text, "<") {
		return apiErr // HTML from a proxy or the SPA: not a message for the user
	}
	if len(text) > 200 {
		text = text[:200] + "…"
	}
	apiErr.Message = text
	return apiErr
}
