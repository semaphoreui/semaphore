package git

import (
	"regexp"
	"strings"
)

var (
	// urlUserInfoRegex matches scheme://userinfo@ in URLs (including passwords with '@' characters)
	urlUserInfoRegex = regexp.MustCompile(`(https?://)([^/\s]+)@`)
	// urlQueryParamRegex matches sensitive credential query parameters in URLs
	urlQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:access_token|token|private_token|password|secret|api_key|apikey)=)([^&\s]+)`)
)

// SanitizeGitOutput redacts sensitive credentials (such as passwords, tokens, and basic auth)
// from git command output, stderr, and error strings so they can be safely logged and shown to users.
func SanitizeGitOutput(output string) string {
	if output == "" {
		return ""
	}

	sanitized := urlUserInfoRegex.ReplaceAllStringFunc(output, func(match string) string {
		sub := urlUserInfoRegex.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		scheme := sub[1]
		userInfo := sub[2]

		colonIdx := strings.Index(userInfo, ":")
		if colonIdx == -1 {
			// Token only: scheme://token@ -> scheme://***@
			return scheme + "***@"
		}

		user := userInfo[:colonIdx]
		// scheme://user:password@ -> scheme://user:***@
		return scheme + user + ":***@"
	})

	return urlQueryParamRegex.ReplaceAllString(sanitized, "${1}***")
}