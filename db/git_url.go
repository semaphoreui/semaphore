package db

import (
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// ValidateGitURL rejects repository URLs that git would interpret as a
// command-line option instead of a repository location. The CmdGitClient
// passes the URL to the git binary as a positional argument, so a value
// beginning with "-" (e.g. "--upload-pack=/path/to/script") would be parsed
// by git as an option and could lead to arbitrary command execution
// (git option injection). Legitimate git URLs (https://, ssh://, git://,
// file://, scp-like user@host:path, or local filesystem paths) never begin
// with "-", so rejecting them here is safe.
func ValidateGitURL(url string, objectName string) error {
	if strings.HasPrefix(strings.TrimSpace(url), "-") {
		return common_errors.NewValidationError(objectName + " url is invalid")
	}

	return nil
}
