package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostConfig_Validate_Host(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"plain host", "github.com", false},
		{"subdomain", "git.internal.example.org", false},
		{"single label", "gitea", false},
		{"underscore", "git_host", false},
		{"hyphen inside", "git-01.example.org", false},

		{"empty", "", true},
		{"a URL is not a host", "https://github.com/", true},
		{"leading hyphen", "-oProxyCommand=evil", true},
		{"trailing hyphen", "github.com-", true},
		{"space", "github.com evil", true},
		{"newline injects a config line", "github.com\n  IdentityFile /etc/shadow", true},
		{"semicolon", "github.com;id", true},
		{"asterisk would match every host", "*", true},
		{"question mark", "github.co?", true},
		{"equals breaks the insteadOf key", "github.com=x", true},
		{"port is not part of a host block", "github.com:22", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HostConfig{Type: HostConfigHost, Name: tt.host, SSHKeyID: 1}

			if tt.wantErr {
				assert.Error(t, h.Validate())
			} else {
				assert.NoError(t, h.Validate())
			}
		})
	}
}

func TestHostConfig_Validate_URL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"repository url", "https://github.com/acme/private-repository/", false},
		{"prefix", "https://github.com/acme/", false},
		{"host root", "https://github.com/", false},
		{"http", "http://git.internal.example.org/acme/", false},
		{"port", "https://git.example.org:8443/acme/", false},

		{"empty", "", true},
		{"no scheme", "github.com/acme/", true},
		{"ssh scheme", "ssh://git@github.com/acme/", true},
		{"git scheme", "git://github.com/acme/", true},
		{"scp syntax", "git@github.com:acme/roles.git", true},
		{"credentials in the url", "https://user:pass@github.com/acme/", true},
		{"space", "https://github.com/acme/ evil", true},
		{"newline", "https://github.com/acme/\nHost x", true},
		{"backtick", "https://github.com/acme/`id`", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HostConfig{Type: HostConfigURL, Name: tt.url, SSHKeyID: 1}

			if tt.wantErr {
				assert.Error(t, h.Validate())
			} else {
				assert.NoError(t, h.Validate())
			}
		})
	}
}

func TestHostConfig_Validate_Common(t *testing.T) {
	t.Run("a credential is required", func(t *testing.T) {
		h := HostConfig{Type: HostConfigHost, Name: "github.com"}

		err := h.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "credential")
	})

	t.Run("unsupported type", func(t *testing.T) {
		h := HostConfig{Type: "ftp", Name: "github.com", SSHKeyID: 1}

		assert.Error(t, h.Validate())
	})

	t.Run("surrounding whitespace is trimmed", func(t *testing.T) {
		h := HostConfig{Type: HostConfigHost, Name: "  github.com  ", SSHKeyID: 1}

		require.NoError(t, h.Validate())
		assert.Equal(t, "github.com", h.Name)
	})
}

// The alias is what a URL mapping rewrites to, so two mappings on one host must
// never collide.
func TestHostConfig_SSHAlias(t *testing.T) {
	a := HostConfig{ID: 7, Type: HostConfigURL, Name: "https://github.com/acme/"}
	b := HostConfig{ID: 8, Type: HostConfigURL, Name: "https://github.com/other/"}

	assert.Equal(t, "semaphore-mapping-7", a.SSHAlias())
	assert.NotEqual(t, a.SSHAlias(), b.SSHAlias())
}

func TestHostConfig_ValidateCredential(t *testing.T) {
	tests := []struct {
		name    string
		typ     HostConfigType
		host    string
		keyType AccessKeyType
		err     string
	}{
		{"host with an ssh key", HostConfigHost, "github.com", AccessKeySSH, ""},
		{"host with a login/password", HostConfigHost, "github.com", AccessKeyLoginPassword,
			"a host mapping needs an SSH key"},
		{"host with a secret", HostConfigHost, "github.com", AccessKeyString,
			"a host mapping needs an SSH key"},
		{"url with an ssh key", HostConfigURL, "https://github.com/acme/", AccessKeySSH, ""},
		{"url with a login/password over https", HostConfigURL, "https://github.com/acme/",
			AccessKeyLoginPassword, ""},
		{"url with a login/password over http", HostConfigURL, "http://github.com/acme/",
			AccessKeyLoginPassword, "https URL"},
		{"url with a secret", HostConfigURL, "https://github.com/acme/", AccessKeyString,
			"a URL mapping needs an SSH key or a login/password credential"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HostConfig{Type: tt.typ, Name: tt.host}

			err := h.ValidateCredential(tt.keyType)
			if tt.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.err)
			}
		})
	}
}
