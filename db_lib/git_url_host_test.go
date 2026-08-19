package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestGitURLHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantHost string
		wantOk   bool
	}{
		{"https url", "https://gitserver/group/submodule1", "gitserver", true},
		{"https url with port", "https://gitserver:8443/group/submodule1.git", "gitserver:8443", true},
		{"http url", "http://gitserver/x", "gitserver", true},
		{"ssh scheme url with port", "ssh://git@gitserver:2222/group/repo.git", "gitserver:2222", true},
		{"scp-like with user", "git@gitserver:group/repo.git", "gitserver", true},
		{"scp-like without user", "gitserver:repo.git", "gitserver", true},
		{"uppercase host normalized", "HTTPS://GitServer/group/repo.git", "gitserver", true},
		{"local absolute path", "/local/path/to/repo", "", false},
		{"windows drive letter", `C:\path\to\repo`, "", false},
		{"empty string", "", "", false},
		{"file url without host", "file:///local/path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, ok := gitURLHost(tt.url)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantHost, host)
		})
	}
}

func TestResolveSubmoduleAccessKey(t *testing.T) {
	mainKey := db.AccessKey{ID: 1, Name: "main"}
	matchedKey := db.AccessKey{ID: 2, Name: "matched"}

	creds := []db.RepositorySubmoduleCredential{
		{Host: "gitserver", AccessKey: matchedKey},
	}

	t.Run("matches configured host", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, creds, "https://gitserver/group/submodule1")
		assert.Equal(t, matchedKey.ID, got.ID)
	})

	t.Run("matches case-insensitively", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, creds, "https://GitServer/group/submodule1")
		assert.Equal(t, matchedKey.ID, got.ID)
	})

	t.Run("falls back to main key when host unmatched", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, creds, "https://other-host/group/submodule2")
		assert.Equal(t, mainKey.ID, got.ID)
	})

	t.Run("falls back to main key when url has no host", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, creds, "/local/submodule")
		assert.Equal(t, mainKey.ID, got.ID)
	})

	t.Run("falls back to main key when no credentials configured", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, nil, "https://gitserver/group/submodule1")
		assert.Equal(t, mainKey.ID, got.ID)
	})

	t.Run("does not match when the url carries a port and the credential does not", func(t *testing.T) {
		got := resolveSubmoduleAccessKey(mainKey, creds, "https://gitserver:8443/group/submodule1")
		assert.Equal(t, mainKey.ID, got.ID)
	})
}
