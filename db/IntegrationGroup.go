package db

import (
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// IntegrationGroup is a set of integrations that run from one endpoint. It
// normally inherits the endpoint's authorization; AuthOverride makes the
// group check its own instead, which keeps the pre-redesign behaviour of
// integrations that had their own auth behind a shared project alias.
type IntegrationGroup struct {
	ID         int    `db:"id" json:"id" backup:"-"`
	ProjectID  int    `db:"project_id" json:"project_id" backup:"-"`
	Name       string `db:"name" json:"name"`
	EndpointID int    `db:"endpoint_id" json:"endpoint_id" backup:"-"`

	// AuthOverride is an explicit flag rather than a nullable auth method:
	// "" already means no auth, so "inherit" needs a field of its own.
	AuthOverride bool `db:"auth_override" json:"auth_override"`

	IntegrationAuth
}

func (g *IntegrationGroup) Validate() error {
	if g.Name == "" {
		return common_errors.NewValidationError("No Name set for integration group")
	}

	if g.EndpointID <= 0 {
		return common_errors.NewValidationError("Integration group must use an endpoint")
	}

	return nil
}

// NormalizeAuth clears the group's own auth when it is not overriding, so a
// stale method or key never survives next to AuthOverride = false.
func (g *IntegrationGroup) NormalizeAuth() {
	if g.AuthOverride {
		return
	}

	g.IntegrationAuth = IntegrationAuth{}
}

// EffectiveAuth is the authorization a webhook to this group is checked
// with: the group's own when it overrides, otherwise the endpoint's.
func (g *IntegrationGroup) EffectiveAuth(endpoint Endpoint) IntegrationAuth {
	if g.AuthOverride {
		return g.IntegrationAuth
	}

	return endpoint.IntegrationAuth
}
