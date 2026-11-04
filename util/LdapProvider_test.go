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

func TestActiveLdapProviders_GroupSettingsFromFlatConfig(t *testing.T) {
	conf := &ConfigType{LdapEnable: true, LdapGroupSearchDN: "ou=groups,dc=e2e,dc=test",
		LdapGroupSearchFilter: "(memberUid={{ .username }})", LdapGroupNameAttribute: "description"}

	p := conf.ActiveLdapProviders()[0].Provider

	assert.Equal(t, "ou=groups,dc=e2e,dc=test", p.GroupSearchDN)
	assert.Equal(t, "(memberUid={{ .username }})", p.GroupSearchFilter)
	assert.Equal(t, "description", p.GroupNameAttribute)
}

func TestLdapProvider_GroupDefaults(t *testing.T) {
	assert.Equal(t, "(|(member={{ .dn }})(uniqueMember={{ .dn }})(memberUid={{ .username }}))", LdapProvider{}.GetGroupSearchFilter())
	assert.Equal(t, "cn", LdapProvider{}.GetGroupNameAttribute())
	custom := LdapProvider{GroupSearchFilter: "(member={{ .dn }})", GroupNameAttribute: "memberOf"}
	assert.Equal(t, "(member={{ .dn }})", custom.GetGroupSearchFilter())
	assert.Equal(t, "memberOf", custom.GetGroupNameAttribute())
}
