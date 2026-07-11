package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
