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