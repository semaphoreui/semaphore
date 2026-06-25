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

	// AllowIdPInitiated enables the Third-Party Initiated Login endpoint
	// (/api/auth/oidc/<id>/initiate) for this provider. When the identity
	// provider redirects the browser to that endpoint (OpenID Connect Core 1.0
	// §4), Semaphore starts a normal SP-initiated Authorization Code flow.
	// Off by default for security.
	AllowIdPInitiated bool `json:"allow_idp_initiated" default:"false"`
}

// ExpectedIssuer returns the issuer identifier that an IdP-initiated request
// must carry in its "iss" parameter. When auto-discovery is used the discovery
// URL is the issuer (go-oidc validates that the discovered "issuer" claim
// equals the URL passed to oidc.NewProvider), otherwise the explicitly
// configured issuer endpoint is used.
func (p *OidcProvider) ExpectedIssuer() string {
	if p.Endpoint.IssuerURL != "" {
		return p.Endpoint.IssuerURL
	}
	return p.AutoDiscovery
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
