package db

import (
	"net/url"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func isHTTPURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// hasURLScheme reports whether the address is a URL with a scheme such as
// "https://", "ssh://" or "git://". scp-style SSH addresses like
// "git@github.com:user/repo.git" have no scheme, carry no URL userinfo and
// cannot be handled by net/url.
func hasURLScheme(rawURL string) bool {
	return strings.Contains(rawURL, "://")
}

// ValidateGitURL rejects repository URLs that git would interpret as a
// command-line option instead of a repository location. The CmdGitClient
// passes the URL to the git binary as a positional argument, so a value
// beginning with "-" (e.g. "--upload-pack=/path/to/script") would be parsed
// by git as an option and could lead to arbitrary command execution
// (git option injection). Legitimate git URLs (https://, ssh://, git://,
// file://, scp-like user@host:path, or local filesystem paths) never begin
// with "-", so rejecting them here is safe.
//
// HTTP(S) URLs must additionally be parseable by net/url. go-git parses the
// URL with net/url and returns the parse error verbatim, and that error quotes
// the whole URL, so a malformed URL such as
// "https://user:secret%zz@example.com/repo.git" would end up in task logs
// together with the credentials typed into it.
func ValidateGitURL(rawURL string, objectName string) error {
	if strings.HasPrefix(strings.TrimSpace(rawURL), "-") {
		return common_errors.NewValidationError(objectName + " url is invalid")
	}

	if isHTTPURL(rawURL) {
		if _, err := url.Parse(rawURL); err != nil {
			return common_errors.NewValidationError(objectName + " url is invalid")
		}
	}

	return nil
}
