package svn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSvnSSHURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"svn+ssh://user@svn.example.com/repo", true},
		{"SVN+SSH://svn.example.com/repo", true},
		{"svn://svn.example.com/repo", false},
		{"git+ssh://git@example.com/repo.git", false},
		{"svn+ssh:/", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSvnSSHURL(tt.url))
		})
	}
}

func TestValidateRevision(t *testing.T) {
	tests := []struct {
		name     string
		revision string
		wantErr  bool
	}{
		{"empty is allowed", "", false},
		{"revision number", "42", false},
		{"HEAD keyword", "HEAD", true},
		{"git hash", "a1b2c3d", true},
		{"option injection", "--config-dir=/tmp", true},
		{"negative", "-1", true},
		{"range", "1:2", true},
		{"too long", "12345678901", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRevision(tt.revision, "task")
			if tt.wantErr {
				assert.EqualError(t, err, "task commit hash is invalid")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
