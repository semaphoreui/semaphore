package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeGitOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no sensitive info",
			input:    "fatal: repository 'https://gitlab.com/group/repo.git' not found",
			expected: "fatal: repository 'https://gitlab.com/group/repo.git' not found",
		},
		{
			name:     "ssh url unchanged",
			input:    "fatal: Could not read from remote repository git@gitlab.com:group/repo.git",
			expected: "fatal: Could not read from remote repository git@gitlab.com:group/repo.git",
		},
		{
			name:     "user and password in url",
			input:    "fatal: Authentication failed for 'https://myuser:secretpassword123@gitlab.com/group/repo.git/'",
			expected: "fatal: Authentication failed for 'https://myuser:***@gitlab.com/group/repo.git/'",
		},
		{
			name:     "token in url",
			input:    "fatal: unable to access 'https://glpat-xxxxxxxxxxxxxxxxxxxx@gitlab.example.com/project.git/': The requested URL returned error: 403",
			expected: "fatal: unable to access 'https://***@gitlab.example.com/project.git/': The requested URL returned error: 403",
		},
		{
			name:     "http protocol with port and user pass",
			input:    "remote: HTTP Basic: Access denied\nfatal: Authentication failed for 'http://admin:p@ssword@git.corp.local:8080/dev/infra.git'",
			expected: "remote: HTTP Basic: Access denied\nfatal: Authentication failed for 'http://admin:***@git.corp.local:8080/dev/infra.git'",
		},
		{
			name:     "multiple urls in output",
			input:    "Cloning from https://oauth2:token123@gitlab.com/a.git and https://bot:pass456@github.com/b.git",
			expected: "Cloning from https://oauth2:***@gitlab.com/a.git and https://bot:***@github.com/b.git",
		},
		{
			name:     "query parameter access token in url",
			input:    "fatal: unable to access 'https://git.example/repo.git?access_token=secret_token_123&env=prod': 401 Unauthorized",
			expected: "fatal: unable to access 'https://git.example/repo.git?access_token=***&env=prod': 401 Unauthorized",
		},
		{
			name:     "query parameter private token in url",
			input:    "fatal: remote error from https://gitlab.corp/project.git?private_token=super_secret_pat",
			expected: "fatal: remote error from https://gitlab.corp/project.git?private_token=***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SanitizeGitOutput(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFormatGitErrorSummary(t *testing.T) {
	tests := []struct {
		name     string
		subCmd   string
		stderr   string
		expected string
	}{
		{
			name:     "empty stderr",
			subCmd:   "ls-remote",
			stderr:   "",
			expected: "git ls-remote failed",
		},
		{
			name:   "gitlab verbose auth denied",
			subCmd: "ls-remote",
			stderr: "remote: HTTP Basic: Access denied. If a password was provided for Git authentication, the password was in-correct or you're required to use a token in-stead of a password. If a token was provided, it was either incorrect, expired, or improperly scoped. See https://gitlab.com/help/topics/git/troubleshooting_git.md#error-on-git-fetch-http-basic-access-denied\nfatal: Authentication failed for 'https://gitlab.com/semaphoreui/non-existent-private-repo.git/'",
			expected: "Authentication failed: Access denied (check token or password)",
		},
		{
			name:     "github permission denied publickey",
			subCmd:   "clone",
			stderr:   "git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.",
			expected: "Permission denied: Invalid or missing SSH key",
		},
		{
			name:     "repository not found",
			subCmd:   "ls-remote",
			stderr:   "remote: Repository not found.\nfatal: repository 'https://github.com/user/private.git/' not found",
			expected: "Repository not found: Check URL or repository permissions",
		},
		{
			name:     "could not resolve host",
			subCmd:   "clone",
			stderr:   "fatal: unable to access 'https://invalid-domain-xyz.local/repo.git/': Could not resolve host: invalid-domain-xyz.local",
			expected: "Could not resolve host: Unable to connect to Git server",
		},
		{
			name:     "unmatched fatal error sanitized",
			subCmd:   "checkout",
			stderr:   "fatal: unable to access 'https://user:secretpass@git.host/repo.git': arbitrary failure",
			expected: "fatal: unable to access 'https://user:***@git.host/repo.git': arbitrary failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FormatGitErrorSummary(tt.subCmd, tt.stderr)
			assert.Equal(t, tt.expected, actual)
		})
	}
}