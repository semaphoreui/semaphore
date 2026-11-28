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

func TestSameURL(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"default svnserve port", "svn://127.0.0.1:3690/trunk", "svn://127.0.0.1/trunk", true},
		{"case and trailing slash", "SVN://Host.EXAMPLE:3690/a%20b/", "svn://host.example/a%20b", true},
		{"encoded path", "svn+ssh://host/a b", "svn+ssh://host/a%20b", true},
		{"ssh port is kept", "svn+ssh://user@host:22/trunk", "svn+ssh://user@host:22/trunk", true},
		{"other port", "svn://host:3691/x", "svn://host/x", false},
		{"other branch", "svn://host/trunk", "svn://host/branches/release", false},
		{"other host", "svn://host/trunk", "svn://other/trunk", false},
		{"other user", "svn+ssh://a@host/trunk", "svn+ssh://b@host/trunk", false},
		{"file url", "file:///srv/repo/trunk", "file:///srv/repo/trunk", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SameURL(tt.a, tt.b))
		})
	}
}
