package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoginConfig() {
	util.Config = &util.ConfigType{
		// Config loading initializes nested pointer structs via reflection;
		// tests setting util.Config directly must do it themselves.
		Mfa:        &util.MultifactorAuthConfig{Totp: &util.TotpConfig{}},
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]util.LdapProvider{
			"corp": {DisplayName: "Corp AD", Server: "corp.example.com:389", Order: 1},
		},
	}
}

func TestLoginMetadata_LdapProviders(t *testing.T) {
	setupLoginConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	w := httptest.NewRecorder()

	login(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"ldap_providers":[{"id":"ldap","name":"LDAP"},{"id":"corp","name":"Corp AD"}]`)
	assert.Contains(t, body, `"login_with_ldap":true`)
}

func TestLogin_UnknownLdapProvider(t *testing.T) {
	setupLoginConfig()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"auth":"jdoe","password":"x","method":"ldap","provider":"nope"}`))
	w := httptest.NewRecorder()

	login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLdapGroupName(t *testing.T) {
	for in, want := range map[string]string{
		"CN=DevOps,OU=Groups,DC=corp,DC=com": "DevOps",
		"cn=devops,ou=groups":                "devops",
		"CN=Dev\\, Ops,OU=Groups":            "Dev, Ops",
		"devops":                             "devops",
		"team=ops":                           "team=ops",
		"team=ops,region=west":               "team=ops,region=west",
		"ou=groups,dc=corp":                  "ou=groups,dc=corp",
		"not a dn, really":                   "not a dn, really",
	} {
		assert.Equal(t, want, ldapGroupName(in), in)
	}
}

func TestLdapGroupFilter(t *testing.T) {
	got, err := ldapGroupFilter(util.DefaultLdapGroupSearchFilter, "uid=al(ice),ou=people", "al*ice")
	require.NoError(t, err)
	assert.Equal(t, `(|(member=uid=al\28ice\29,ou=people)(uniqueMember=uid=al\28ice\29,ou=people)(memberUid=al\2aice))`, got)

	got, err = ldapGroupFilter("(sAMAccountName={{ .username }})", "cn=x", "alice")
	require.NoError(t, err)
	assert.Equal(t, "(sAMAccountName=alice)", got)

	_, err = ldapGroupFilter("(memberUid={{ .usernme }})", "cn=x", "alice")
	assert.Error(t, err)
}

func TestLdapGroupNames(t *testing.T) {
	entries := []*ldap.Entry{
		ldap.NewEntry("cn=devops,ou=groups", map[string][]string{"cn": {"devops"}}),
		ldap.NewEntry("cn=devops,ou=teams", map[string][]string{"cn": {"devops"}}),
		ldap.NewEntry("uid=alice,ou=people", map[string][]string{"memberOf": {"CN=Release,OU=Groups", "", "CN=Release,OU=Groups"}}),
	}
	assert.Equal(t, []string{"devops"}, ldapGroupNames(entries[:2], "cn"))
	assert.Equal(t, []string{"Release"}, ldapGroupNames(entries[2:], "memberOf"))
	assert.Equal(t, []string{"Release"}, ldapGroupNames(entries[2:], "memberof"))
	assert.Equal(t, []string{}, ldapGroupNames(nil, "cn"))
}
