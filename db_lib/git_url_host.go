package db_lib

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// scpLikeGitURL matches the scp-like git remote shorthand, e.g.
// "git@gitserver:group/repo.git" or "gitserver:repo.git".
var scpLikeGitURL = regexp.MustCompile(`^(?:[^@/\s]+@)?([^:/\s]+):.+$`)

// gitURLHost extracts the lowercased host[:port] from a git remote URL. It
// supports scheme:// URLs (https, http, ssh, git, ...) and the scp-like
// shorthand (user@host:path). ok is false for local filesystem paths or URLs
// that carry no host (e.g. "file:///path" with no host part).
func gitURLHost(rawURL string) (host string, ok bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}

	if strings.Contains(rawURL, "://") {
		u, err := url.Parse(rawURL)
		if err != nil || u.Host == "" {
			return "", false
		}
		return strings.ToLower(u.Host), true
	}

	m := scpLikeGitURL.FindStringSubmatch(rawURL)
	if m == nil {
		return "", false
	}

	h := m[1]
	// A single ASCII letter before the colon is almost certainly a Windows
	// drive letter (e.g. "C:\path"), not a hostname.
	if len(h) == 1 && ((h[0] >= 'a' && h[0] <= 'z') || (h[0] >= 'A' && h[0] <= 'Z')) {
		return "", false
	}

	return strings.ToLower(h), true
}

// resolveSubmoduleAccessKey returns the AccessKey configured for submoduleURL's
// host among creds (exact host[:port] match, case-insensitive), falling back
// to mainKey when there is no explicit match. This preserves today's behavior
// for submodules that share the main repository's host/credentials, and never
// sends a stored credential to a host the project admin didn't explicitly
// authorize for it.
func resolveSubmoduleAccessKey(mainKey db.AccessKey, creds []db.RepositorySubmoduleCredential, submoduleURL string) db.AccessKey {
	host, ok := gitURLHost(submoduleURL)
	if !ok {
		return mainKey
	}

	for _, cred := range creds {
		if strings.EqualFold(cred.Host, host) {
			return cred.AccessKey
		}
	}

	return mainKey
}
