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

func TestIntegration_VerificationSecret(t *testing.T) {
	stringKey := AccessKey{Type: AccessKeyString, String: "tok"}
	loginKey := AccessKey{Type: AccessKeyLoginPassword, LoginPassword: LoginPassword{Login: "bob", Password: "pw"}}
	sshKey := AccessKey{Type: AccessKeySSH, SshKey: SshKey{PrivateKey: "k"}}

	tests := []struct {
		name           string
		key            AccessKey
		expectedSecret string
		expectedOk     bool
	}{
		{"token with a string key gives its value", stringKey, "tok", true},
		{"hmac with a login/password key gives the password", loginKey, "pw", true},
		{"github with a string key", stringKey, "tok", true},
		{"bitbucket with a string key", stringKey, "tok", true},
		{"hmac-sha512 with a string key", stringKey, "tok", true},
		{"token with an empty string key is refused", AccessKey{Type: AccessKeyString}, "", true},
		{"token with a login/password key without a password is refused", AccessKey{Type: AccessKeyLoginPassword, LoginPassword: LoginPassword{Login: "bob"}}, "", true},
		{"token with an ssh key is refused", sshKey, "", false},
		{"token with a none key is refused", AccessKey{Type: AccessKeyNone}, "", false},
		{"token with an unloaded key is refused", AccessKey{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, ok := tt.key.VerificationToken()

			assert.Equal(t, tt.expectedSecret, got)
			assert.Equal(t, tt.expectedOk, ok)
		})
	}
}
