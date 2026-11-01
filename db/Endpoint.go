package db

import (
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type IntegrationAuthMethod string

const (
	IntegrationAuthNone       = ""
	IntegrationAuthGitHub     = "github"
	IntegrationAuthToken      = "token"
	IntegrationAuthHmac       = "hmac"
	IntegrationAuthHmacSha512 = "hmac-sha512"
	IntegrationAuthBitbucket  = "bitbucket"
	IntegrationAuthBasic      = "basic"
)

// usesVerificationSecret reports whether the auth method checks the request
// against one shared secret string: token, HMAC-SHA256, HMAC-SHA512, GitHub
// and Bitbucket do. BasicAuth compares a login and a password instead, and
// none checks nothing.
func (m IntegrationAuthMethod) usesVerificationSecret() bool {
	switch m {
	case IntegrationAuthToken, IntegrationAuthHmac, IntegrationAuthHmacSha512,
		IntegrationAuthGitHub, IntegrationAuthBitbucket:
		return true
	default:
		return false
	}
}

// AcceptsKeyType reports whether a credential of type t can back the auth
// method. A method that verifies one shared secret takes a string key or
// the password of a login/password key; BasicAuth needs a login as well, so
// only a login/password key fits. A method that reads no key — none, or one
// the webhook receiver does not know and refuses outright (clients send
// "none" as well as "") — accepts any key.
func (m IntegrationAuthMethod) AcceptsKeyType(t AccessKeyType) bool {
	switch {
	case m == IntegrationAuthBasic:
		return t == AccessKeyLoginPassword
	case m.usesVerificationSecret():
		return t == AccessKeyLoginPassword || t == AccessKeyString
	default:
		return true
	}
}

// IntegrationAuth is how an incoming webhook is verified. An Endpoint carries
// one, and an IntegrationGroup carries its own when it overrides the
// endpoint's. AuthSecret is the key loaded by AuthSecretID; the store never
// fills it.
type IntegrationAuth struct {
	AuthMethod   IntegrationAuthMethod `db:"auth_method" json:"auth_method"`
	AuthSecretID *int                  `db:"auth_secret_id" json:"auth_secret_id" backup:"-"`
	AuthHeader   string                `db:"auth_header" json:"auth_header"`
	AuthSecret   AccessKey             `db:"-" json:"-" backup:"-"`
}

// VerificationSecret is the shared secret an incoming webhook is verified
// against by the token, HMAC, GitHub and Bitbucket methods: the value of a
// string key, or the password of a login/password key.
//
// ok is false when the auth method needs that secret and the key cannot
// provide it — wrong key type, unloaded key or an empty value — so the
// request must be refused. It is never reported as an empty secret: with
// the token method that would match a request carrying no header at all.
// For none and BasicAuth, which do not use a shared secret, the secret is
// empty and ok is true.
func (a *IntegrationAuth) VerificationSecret() (secret string, ok bool) {
	if !a.AuthMethod.usesVerificationSecret() {
		return "", true
	}

	switch a.AuthSecret.Type {
	case AccessKeyString:
		secret = a.AuthSecret.String
	case AccessKeyLoginPassword:
		secret = a.AuthSecret.LoginPassword.Password
	}

	return secret, secret != ""
}

// Endpoint is a project's entry point for webhooks: the aliases (public URLs)
// requests arrive at and the authorization they are checked with. Integration
// groups choose an endpoint; the integrations of those groups run.
type Endpoint struct {
	ID          int    `db:"id" json:"id" backup:"-"`
	ProjectID   int    `db:"project_id" json:"project_id" backup:"-"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`

	IntegrationAuth
}

func (e *Endpoint) Validate() error {
	if e.Name == "" {
		return common_errors.NewValidationError("No Name set for endpoint")
	}

	return nil
}

// EndpointAlias is one public URL of an endpoint. The alias value is a random
// string that works as a bearer secret: it is unique across the instance and
// is never written to the audit log.
type EndpointAlias struct {
	ID         int    `db:"id" json:"id" backup:"-"`
	Alias      string `db:"alias" json:"alias"`
	ProjectID  int    `db:"project_id" json:"project_id" backup:"-"`
	EndpointID int    `db:"endpoint_id" json:"endpoint_id" backup:"-"`
}

func (alias EndpointAlias) ToAlias() Alias {
	return Alias{
		ID:        alias.ID,
		Alias:     alias.Alias,
		ProjectID: alias.ProjectID,
	}
}
