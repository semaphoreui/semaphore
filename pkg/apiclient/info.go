package apiclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// SystemInfo is the part of GET /api/info a client cares about.
type SystemInfo struct {
	Version           string `json:"version"`
	WebHost           string `json:"web_host"`
	UseRemoteRunner   bool   `json:"use_remote_runner"`
	LoginWithPassword bool   `json:"login_with_password"`
	SubscriptionState string `json:"subscription_state"`
	GitClient         string `json:"git_client"`
	ScheduleTimezone  string `json:"schedule_timezone"`
	BoltdbUsed        bool   `json:"boltdb_used"`
	AuthMethods       struct {
		Totp *struct {
			AllowRecovery bool `json:"allow_recovery"`
		} `json:"totp,omitempty"`
		Email *struct{} `json:"email,omitempty"`
	} `json:"auth_methods"`
}

// LoginMethods is GET /api/auth/login: how this server lets a user in. It is
// public, so it works before authentication.
type LoginMethods struct {
	LoginWithPassword bool `json:"login_with_password"`
	LoginWithLdap     bool `json:"login_with_ldap"`
	OidcProviders     []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"oidc_providers"`
	LdapProviders []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"ldap_providers"`
	AuthMethods struct {
		Totp  *struct{} `json:"totp,omitempty"`
		Email *struct{} `json:"email,omitempty"`
	} `json:"auth_methods"`
}

// Me is GET /api/user: the authenticated user and the global flags the UI
// derives its menus from.
type Me struct {
	db.User
	CanCreateProject      bool  `json:"can_create_project"`
	CanReadAuditLog       bool  `json:"can_read_audit_log"`
	GlobalPermissions     int64 `json:"global_permissions"`
	HasActiveSubscription bool  `json:"has_active_subscription"`
}

// Ping checks that the URL points at a Semaphore server. It needs no token.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.DoRaw(ctx, http.MethodGet, "/ping", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
	if strings.TrimSpace(string(body)) != "pong" {
		return fmt.Errorf("%s does not look like a Semaphore server (GET /api/ping answered %q)", c.BaseURL(), strings.TrimSpace(string(body)))
	}
	return nil
}

// Info returns the server description. It needs a valid token.
func (c *Client) Info(ctx context.Context) (SystemInfo, error) {
	var info SystemInfo
	err := c.Do(ctx, http.MethodGet, "/info", nil, &info)
	return info, err
}

// LoginMethods returns the public login configuration of the server.
func (c *Client) LoginMethods(ctx context.Context) (LoginMethods, error) {
	var m LoginMethods
	err := c.Do(ctx, http.MethodGet, "/auth/login", nil, &m)
	return m, err
}

// Me returns the authenticated user. It is the cheapest way to validate a token.
func (c *Client) Me(ctx context.Context) (Me, error) {
	var me Me
	err := c.Do(ctx, http.MethodGet, "/user", nil, &me)
	return me, err
}
