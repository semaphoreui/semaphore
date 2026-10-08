package svn

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

const svnSSHScheme = "svn+ssh://"

// IsSvnSSHURL reports whether the URL tunnels Subversion through ssh. The
// scheme is case-insensitive (RFC 3986). It is matched apart from the other
// schemes because "+" is not a word character.
func IsSvnSSHURL(rawURL string) bool {
	return len(rawURL) >= len(svnSSHScheme) && strings.EqualFold(rawURL[:len(svnSSHScheme)], svnSSHScheme)
}

// revisionPattern matches a revision number, which a task of a Subversion
// repository records in place of a commit hash.
var revisionPattern = regexp.MustCompile(`^[0-9]{1,10}$`)

// ValidateRevision rejects a revision that is not a revision number, which
// keeps it from carrying svn options, revision keywords or ranges.
func ValidateRevision(revision string, objectName string) error {
	if revision == "" {
		return nil
	}

	if !revisionPattern.MatchString(revision) {
		return common_errors.NewValidationError(objectName + " commit hash is invalid")
	}

	return nil
}

// defaultSvnPort is the svnserve port, which svn drops from the URLs it prints.
const defaultSvnPort = "3690"

// SameURL reports whether two URLs name the same location once canonicalized
// the way svn canonicalizes the URL of a working copy: scheme and host in
// lower case, no svnserve default port, path percent-encoded and without a
// trailing slash.
func SameURL(a, b string) bool {
	return canonicalURL(a) == canonicalURL(b)
}

func canonicalURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	scheme := strings.ToLower(u.Scheme)

	host := strings.ToLower(u.Hostname())
	if port := u.Port(); port != "" && !(scheme == "svn" && port == defaultSvnPort) {
		host += ":" + port
	}

	user := ""
	if u.User != nil {
		user = u.User.String() + "@"
	}

	return scheme + "://" + user + host + strings.TrimRight(u.Path, "/")
}
