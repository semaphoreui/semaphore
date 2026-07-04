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

func TestRepositoryValidate_AcceptsNormalURL(t *testing.T) {
	repo := Repository{
		Name:      "ok",
		GitURL:    "https://github.com/semaphoreui/semaphore.git",
		GitBranch: "main",
		SSHKeyID:  1,
	}

	assert.NoError(t, repo.Validate())
}
