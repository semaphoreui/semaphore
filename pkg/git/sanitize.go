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

// FormatGitErrorSummary parses raw Git stderr and formats a concise, human-friendly summary
// for UI display, while ensuring all sensitive credentials remain redacted.
func FormatGitErrorSummary(subCmd, stderr string) string {
	clean := strings.TrimSpace(stderr)
	if clean == "" {
		if subCmd != "" {
			return "git " + subCmd + " failed"
		}
		return "git command failed"
	}

	lower := strings.ToLower(clean)

	switch {
	case strings.Contains(lower, "authentication failed") || strings.Contains(lower, "access denied") || strings.Contains(lower, "invalid username or password"):
		return "Authentication failed: Access denied (check token or password)"
	case strings.Contains(lower, "permission denied (publickey)"):
		return "Permission denied: Invalid or missing SSH key"
	case strings.Contains(lower, "could not resolve host"):
		return "Could not resolve host: Unable to connect to Git server"
	case strings.Contains(lower, "connection refused"):
		return "Connection refused: Git server is unreachable"
	case strings.Contains(lower, "repository") && (strings.Contains(lower, "not found") || strings.Contains(lower, "does not exist")):
		return "Repository not found: Check URL or repository permissions"
	case strings.Contains(lower, "ssl certificate problem") || strings.Contains(lower, "certificate verification failed"):
		return "SSL certificate verification failed"
	}

	// Fallback: extract the fatal/error line if present
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "fatal:") || strings.HasPrefix(strings.ToLower(line), "error:") {
			return SanitizeGitOutput(line)
		}
	}

	// Generic first line
	firstLine := strings.Split(clean, "\n")[0]
	return SanitizeGitOutput(strings.TrimSpace(firstLine))
}