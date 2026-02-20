package util

import log "github.com/sirupsen/logrus"

type OidcProvider struct {
	ClientID         string       `json:"client_id"`
	ClientIDFile     string       `json:"client_id_file"`
	ClientSecret     string       `json:"client_secret"`
	ClientSecretFile string       `json:"client_secret_file"`
	RedirectURL      string       `json:"redirect_url"`
	Scopes           []string     `json:"scopes"`
	DisplayName      string       `json:"display_name"`
	Color            string       `json:"color"`
	Icon             string       `json:"icon"`
	AutoDiscovery    string       `json:"provider_url"`
	Endpoint         oidcEndpoint `json:"endpoint"`
	UsernameClaim    string       `json:"username_claim" default:"preferred_username"`
	NameClaim        string       `json:"name_claim" default:"preferred_username"`
	EmailClaim       string       `json:"email_claim" default:"email"`
	GroupsClaim      string       `json:"groups_claim" default:"groups"`
	AdminGroup       string       `json:"admin_group" default:""` // Group from the groups in the claims that members will be mapped as admin on semaphore
	Order            int          `json:"order"`
	// ReturnViaState when true, passes the return path via the OAuth state parameter instead of the redirect URL path. This is useful for OAuth providers that have strict redirect URL validation.
	ReturnViaState bool `json:"return_via_state"`
}

type ClaimsProvider interface {
	GetUsernameClaim() string
	GetEmailClaim() string
	GetNameClaim() string
	IsAdminMappingEnable() bool
	IsAdminUserClaims(claims map[string]any) bool
}

func (p *OidcProvider) IsAdminMappingEnable() bool {
	return p.AdminGroup != ""
}

func (p *OidcProvider) IsAdminUserClaims(claims map[string]any) bool {
	// SECURITY NOTE: Group membership check is CASE-SENSITIVE.
	//                We match exact string values from the groups claim.
	//                This follows JWT/OIDC conventions and prevents accidental privilege escalation
	//                due to inconsistent casing from IdPs (e.g. "Admin" ≠ "admin").
	rawList, ok := claims[p.GroupsClaim].([]any)

	if !ok {
		log.Info("Warning: groups claim is not a list/array of strings → cannot reliably detect admin group membership. Assuming non-admin for security.")
		return false
	}

	for _, g := range rawList {
		group, ok := g.(string)

		if ok {
			if p.AdminGroup == group {
				return true
			}
		} else {
			log.Info("Warning: groups claim contains non-string value(s) — ignoring invalid entries")
		}
	}
	return false
}

func (p *OidcProvider) GetUsernameClaim() string {
	return p.UsernameClaim
}

func (p *OidcProvider) GetEmailClaim() string {
	return p.EmailClaim
}

func (p *OidcProvider) GetNameClaim() string {
	return p.NameClaim
}
