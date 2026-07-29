package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveLdapProviders_Order(t *testing.T) {
	conf := &ConfigType{
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]LdapProvider{
			"corp":   {DisplayName: "Corp AD", Server: "corp.example.com:389", Order: 2},
			"berlin": {DisplayName: "Berlin", Server: "berlin.example.com:389", Order: 1},
			// Reserved ID must be skipped: it would collide with the legacy provider.
			"ldap": {DisplayName: "Impostor", Server: "evil.example.com:389"},
		},
	}

	providers := conf.ActiveLdapProviders()

	require.Len(t, providers, 3)
	// Legacy flat config always first, right after the internal login tab.
	assert.Equal(t, "ldap", providers[0].ID)
	assert.Equal(t, "legacy.example.com:389", providers[0].Provider.Server)
	assert.Equal(t, "berlin", providers[1].ID)
	assert.Equal(t, "corp", providers[2].ID)
}

func TestActiveLdapProviders_NewConfigOnly(t *testing.T) {
	conf := &ConfigType{
		LdapProviders: map[string]LdapProvider{
			"corp": {DisplayName: "Corp AD", Server: "corp.example.com:389"},
		},
	}

	providers := conf.ActiveLdapProviders()

	require.Len(t, providers, 1)
	assert.Equal(t, "corp", providers[0].ID)
}

func TestActiveLdapProviders_Empty(t *testing.T) {
	conf := &ConfigType{}
	assert.Empty(t, conf.ActiveLdapProviders())
}

func TestGetLdapProvider(t *testing.T) {
	conf := &ConfigType{
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]LdapProvider{
			"corp": {Server: "corp.example.com:389"},
		},
	}

	p, ok := conf.GetLdapProvider("corp")
	require.True(t, ok)
	assert.Equal(t, "corp.example.com:389", p.Server)

	p, ok = conf.GetLdapProvider("ldap")
	require.True(t, ok)
	assert.Equal(t, "legacy.example.com:389", p.Server)

	_, ok = conf.GetLdapProvider("nope")
	assert.False(t, ok)
}

func TestLdapProvider_GetMappings_DefaultsWhenNil(t *testing.T) {
	p := LdapProvider{}
	m := p.GetMappings()
	require.NotNil(t, m)
	assert.Equal(t, "dn", m.DN)
	assert.Equal(t, "mail", m.Mail)
	assert.Equal(t, "uid", m.UID)
	assert.Equal(t, "cn", m.CN)
}

func TestLdapProvider_ShouldVerifyTLS(t *testing.T) {
	tests := []struct {
		name     string
		provider LdapProvider
		expected bool
	}{
		{"unconfigured provider verifies by default", LdapProvider{}, true},
		{"explicit opt-out", LdapProvider{TLSSkipVerify: true}, false},
		{"CA bundle implies verification", LdapProvider{TLSCACertFile: "/etc/ssl/ca.pem"}, true},
		{"CA bundle overrides opt-out", LdapProvider{TLSSkipVerify: true, TLSCACertFile: "/etc/ssl/ca.pem"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.provider.ShouldVerifyTLS())
		})
	}
}

// Providers under `ldap_providers` are new config, so they verify unless the
// operator opts out — unlike the synthesized legacy entry.
func TestActiveLdapProviders_NewProviderVerifiesByDefault(t *testing.T) {
	conf := &ConfigType{
		LdapProviders: map[string]LdapProvider{
			"corp":     {Server: "corp.example.com:636", NeedTLS: true},
			"insecure": {Server: "old.example.com:636", NeedTLS: true, TLSSkipVerify: true},
		},
	}

	corp, ok := conf.GetLdapProvider("corp")
	require.True(t, ok)
	assert.False(t, corp.TLSSkipVerify)
	assert.True(t, corp.ShouldVerifyTLS())

	insecure, ok := conf.GetLdapProvider("insecure")
	require.True(t, ok)
	assert.False(t, insecure.ShouldVerifyTLS())
}

// The legacy flat config keeps the released behaviour: no verification unless
// the operator opts in, so upgrading installs on self-signed certs still work.
func TestActiveLdapProviders_LegacyTLSDefaults(t *testing.T) {
	tests := []struct {
		name           string
		conf           ConfigType
		wantSkipVerify bool
		wantVerify     bool
	}{
		{
			name:           "unset keeps released insecure behaviour",
			conf:           ConfigType{LdapEnable: true, LdapServer: "legacy.example.com:636", LdapNeedTLS: true},
			wantSkipVerify: true,
			wantVerify:     false,
		},
		{
			name:           "opt in via ldap_tls_verify",
			conf:           ConfigType{LdapEnable: true, LdapServer: "legacy.example.com:636", LdapNeedTLS: true, LdapTLSVerify: true},
			wantSkipVerify: false,
			wantVerify:     true,
		},
		{
			name:           "CA bundle alone forces verification",
			conf:           ConfigType{LdapEnable: true, LdapServer: "legacy.example.com:636", LdapNeedTLS: true, LdapTLSCACertFile: "/etc/ssl/corp-ca.pem"},
			wantSkipVerify: true,
			wantVerify:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := tt.conf
			p, ok := conf.GetLdapProvider("ldap")
			require.True(t, ok)
			assert.Equal(t, tt.wantSkipVerify, p.TLSSkipVerify)
			assert.Equal(t, tt.wantVerify, p.ShouldVerifyTLS())
		})
	}
}

func TestActiveLdapProviders_MapsLegacyTLSFields(t *testing.T) {
	conf := &ConfigType{
		LdapEnable:        true,
		LdapServer:        "legacy.example.com:636",
		LdapNeedTLS:       true,
		LdapTLSVerify:     true,
		LdapTLSCACertFile: "/etc/ssl/corp-ca.pem",
	}

	p, ok := conf.GetLdapProvider("ldap")
	require.True(t, ok)
	assert.False(t, p.TLSSkipVerify)
	assert.Equal(t, "/etc/ssl/corp-ca.pem", p.TLSCACertFile)
	assert.True(t, p.ShouldVerifyTLS())
}
