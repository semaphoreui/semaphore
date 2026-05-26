// Package jwt provides JWT issuance for Semaphore task executions so that
// playbooks can authenticate to external systems (e.g. HashiCorp Vault) using
// the JWT auth method. Tokens are signed with RS256 and the corresponding
// public key is published at the /.well-known/jwks.json endpoint.
package jwt

import "time"

// TaskClaims is the JWT payload issued by Semaphore for a single task run.
type TaskClaims struct {
	// Registered claims
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	JWTID     string `json:"jti,omitempty"`

	// Semaphore-specific claims
	TaskID       int    `json:"task_id"`
	ProjectID    int    `json:"project_id"`
	ProjectName  string `json:"project_name,omitempty"`
	TemplateID   int    `json:"template_id"`
	TemplateName string `json:"template_name,omitempty"`
	UserID       *int   `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
}

// TaskInfo bundles the data needed to mint a TaskClaims set.
type TaskInfo struct {
	TaskID       int
	ProjectID    int
	ProjectName  string
	TemplateID   int
	TemplateName string
	UserID       *int
	Username     string
}

// SignerOptions controls token issuance parameters.
type SignerOptions struct {
	Issuer   string
	Audience string
	TTL      time.Duration
}
