package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://github.com/semaphoreui/semaphore.git", false},
		{"http", "http://example.com/repo.git", false},
		{"ssh scheme", "ssh://git@example.com/repo.git", false},
		{"scp-like ssh", "git@github.com:semaphoreui/semaphore.git", false},
		{"git scheme", "git://example.com/repo.git", false},
		{"file scheme", "file:///srv/git/repo.git", false},
		{"local absolute path", "/srv/git/repo.git", false},
		{"empty", "", false}, // emptiness is handled separately by Repository.Validate

		{"upload-pack option injection", "--upload-pack=/tmp/evil.sh", true},
		{"single dash option", "-oProxyCommand=evil", true},
		{"leading whitespace then dash", "  --upload-pack=/tmp/evil.sh", true},

		// HTTP(S) URLs net/url can't parse: go-git would quote them, credentials
		// included, in the error it returns.
		{"https invalid percent escape in password", "https://user:secret%zz@example.com/repo.git", true},
		{"http invalid percent escape in path", "http://example.com/re%zzpo.git", true},
		{"https control character", "https://example.com/repo\x00.git", true},
		{"HTTPS uppercase scheme is still validated", "HTTPS://user:secret%zz@example.com/repo.git", true},

		// Non-HTTP URLs are not parsed here; GetGitURL redacts them at use time.
		{"ssh invalid percent escape is not rejected", "ssh://user:secret%zz@example.com/repo.git", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.url, "repository")
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "repository url is invalid")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsHTTPURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/user/repo.git", true},
		{"http://example.com/repo.git", true},
		{"HTTPS://example.com/repo.git", true},
		{"ssh://git@example.com/repo.git", false},
		{"git@github.com:user/repo.git", false},
		{"/srv/git/repo.git", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, isHTTPURL(tt.url))
		})
	}
}

func TestHasURLScheme(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/user/repo.git", true},
		{"ssh://git@example.com/repo.git", true},
		{"git://example.com/repo.git", true},
		{"file:///srv/git/repo.git", true},
		{"git@github.com:user/repo.git", false},
		{"github.com:user/repo.git", false},
		{"/srv/git/repo.git", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, hasURLScheme(tt.url))
		})
	}
}

func TestRepositoryValidate_RejectsOptionInjectionURL(t *testing.T) {
	repo := Repository{
		Name:      "rce",
		GitURL:    "--upload-pack=/tmp/evil.sh",
		GitBranch: "main",
		SSHKeyID:  1,
	}

	err := repo.Validate()
	require.Error(t, err)
	assert.ErrorContains(t, err, "repository url is invalid")
}

func TestRepositoryValidate_RejectsMalformedHTTPURL(t *testing.T) {
	repo := Repository{
		Name:      "malformed",
		GitURL:    "https://user:secret%zz@example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  1,
	}

	err := repo.Validate()
	require.Error(t, err)
	assert.ErrorContains(t, err, "repository url is invalid")
	assert.NotContains(t, err.Error(), "secret")
}

func TestRepositoryValidate_AcceptsNormalURL(t *testing.T) {
	repo := Repository{
		Name:      "ok",
		GitURL:    "https://github.com/semaphoreui/semaphore.git",
		GitBranch: "main",
		SSHKeyID:  1,
	}

	assert.NoError(t, repo.Validate())
}
