package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationAuthMethod_AcceptsKeyType(t *testing.T) {
	single := []IntegrationAuthMethod{
		IntegrationAuthToken, IntegrationAuthHmac, IntegrationAuthHmacSha512,
		IntegrationAuthGitHub, IntegrationAuthBitbucket,
	}
	for _, m := range single {
		t.Run(string(m)+" takes a string or a login/password key", func(t *testing.T) {
			assert.True(t, m.AcceptsKeyType(AccessKeyString))
			assert.True(t, m.AcceptsKeyType(AccessKeyLoginPassword))
			assert.False(t, m.AcceptsKeyType(AccessKeySSH))
			assert.False(t, m.AcceptsKeyType(AccessKeyNone))
		})
	}

	t.Run("basic needs a login, so only login/password", func(t *testing.T) {
		assert.True(t, IntegrationAuthMethod(IntegrationAuthBasic).AcceptsKeyType(AccessKeyLoginPassword))
		assert.False(t, IntegrationAuthMethod(IntegrationAuthBasic).AcceptsKeyType(AccessKeyString))
		assert.False(t, IntegrationAuthMethod(IntegrationAuthBasic).AcceptsKeyType(AccessKeySSH))
	})

	t.Run("a method that reads no key accepts any", func(t *testing.T) {
		// "" is the stored value for no auth; clients and the API docs also send "none".
		for _, m := range []IntegrationAuthMethod{IntegrationAuthNone, "none", "jwt"} {
			for _, kt := range []AccessKeyType{AccessKeyString, AccessKeyLoginPassword, AccessKeySSH, AccessKeyNone, ""} {
				assert.True(t, m.AcceptsKeyType(kt), "%q / %q", m, kt)
			}
		}
	})
}

func TestIntegrationAuth_VerificationSecret(t *testing.T) {
	stringKey := AccessKey{Type: AccessKeyString, String: "tok"}
	loginKey := AccessKey{Type: AccessKeyLoginPassword, LoginPassword: LoginPassword{Login: "bob", Password: "pw"}}
	sshKey := AccessKey{Type: AccessKeySSH, SshKey: SshKey{PrivateKey: "k"}}

	tests := []struct {
		name   string
		method IntegrationAuthMethod
		key    AccessKey
		want   string
		wantOk bool
	}{
		{"token with a string key gives its value", IntegrationAuthToken, stringKey, "tok", true},
		{"hmac with a login/password key gives the password", IntegrationAuthHmac, loginKey, "pw", true},
		{"github with a string key", IntegrationAuthGitHub, stringKey, "tok", true},
		{"bitbucket with a string key", IntegrationAuthBitbucket, stringKey, "tok", true},
		{"hmac-sha512 with a string key", IntegrationAuthHmacSha512, stringKey, "tok", true},
		{"token with an empty string key is refused", IntegrationAuthToken, AccessKey{Type: AccessKeyString}, "", false},
		{"token with a login/password key without a password is refused", IntegrationAuthToken, AccessKey{Type: AccessKeyLoginPassword, LoginPassword: LoginPassword{Login: "bob"}}, "", false},
		{"token with an ssh key is refused", IntegrationAuthToken, sshKey, "", false},
		{"token with a none key is refused", IntegrationAuthToken, AccessKey{Type: AccessKeyNone}, "", false},
		{"token with an unloaded key is refused", IntegrationAuthToken, AccessKey{}, "", false},
		{"basic does not use a shared secret", IntegrationAuthBasic, stringKey, "", true},
		{"none does not use a shared secret", IntegrationAuthNone, sshKey, "", true},
		{"unknown method does not use a shared secret", "jwt", AccessKey{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := IntegrationAuth{AuthMethod: tt.method, AuthSecret: tt.key}

			got, ok := auth.VerificationSecret()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestEndpoint_Validate(t *testing.T) {
	assert.NoError(t, (&Endpoint{Name: "hooks"}).Validate())
	assert.ErrorContains(t, (&Endpoint{}).Validate(), "No Name set for endpoint")
}

func TestIntegrationGroup_Validate(t *testing.T) {
	tests := []struct {
		name    string
		group   IntegrationGroup
		wantErr string
	}{
		{"valid", IntegrationGroup{Name: "ci", EndpointID: 3}, ""},
		{"no name", IntegrationGroup{EndpointID: 3}, "No Name set for integration group"},
		{"no endpoint", IntegrationGroup{Name: "ci"}, "Integration group must use an endpoint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIntegrationGroup_EffectiveAuth(t *testing.T) {
	keyID := 5
	endpoint := Endpoint{IntegrationAuth: IntegrationAuth{AuthMethod: IntegrationAuthToken, AuthSecretID: &keyID, AuthHeader: "X-Token"}}
	own := IntegrationAuth{AuthMethod: IntegrationAuthBasic}

	inheriting := IntegrationGroup{IntegrationAuth: own}
	assert.Equal(t, endpoint.IntegrationAuth, inheriting.EffectiveAuth(endpoint), "without override the endpoint decides")

	overriding := IntegrationGroup{AuthOverride: true, IntegrationAuth: own}
	assert.Equal(t, own, overriding.EffectiveAuth(endpoint))
}

func TestIntegrationGroup_NormalizeAuth(t *testing.T) {
	keyID := 5
	own := IntegrationAuth{AuthMethod: IntegrationAuthToken, AuthSecretID: &keyID, AuthHeader: "X-Token"}

	inheriting := IntegrationGroup{IntegrationAuth: own}
	inheriting.NormalizeAuth()
	assert.Equal(t, IntegrationAuth{}, inheriting.IntegrationAuth, "a group that inherits keeps no auth of its own")

	overriding := IntegrationGroup{AuthOverride: true, IntegrationAuth: own}
	overriding.NormalizeAuth()
	assert.Equal(t, own, overriding.IntegrationAuth)
}
