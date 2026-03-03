package util

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
	Order            int          `json:"order"`
	// ReturnViaState when true, passes the return path via the OAuth state parameter instead of the redirect URL path. This is useful for OAuth providers that have strict redirect URL validation.
	ReturnViaState bool `json:"return_via_state" default:"true"`
	// AllowIdpInitiated when true, permits IdP-initiated OIDC login (e.g. clicking the app
	// tile in Okta). This skips CSRF state validation since the flow does not originate from
	// Semaphore. Only enable this for trusted identity providers.
	AllowIdpInitiated bool `json:"allow_idp_initiated"`
}

type ClaimsProvider interface {
	GetUsernameClaim() string
	GetEmailClaim() string
	GetNameClaim() string
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
