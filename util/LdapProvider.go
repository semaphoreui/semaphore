package util

import "sort"

// LdapProvider is one LDAP directory configured under `ldap_providers`
// (analog of OidcProvider for `oidc_providers`). Field JSON names mirror
// the legacy flat ldap_* config options.
type LdapProvider struct {
	DisplayName  string        `json:"display_name"`
	Server       string        `json:"server"`
	NeedTLS      bool          `json:"need_tls"`
	BindDN       string        `json:"bind_dn"`
	BindPassword string        `json:"bind_password"`
	SearchDN     string        `json:"search_dn"`
	SearchFilter string        `json:"search_filter"`
	Mappings     *LdapMappings `json:"mappings"`
	Color        string        `json:"color"`
	Icon         string        `json:"icon"`
	Order        int           `json:"order"`

	// TLSCACertFile is a PEM bundle used to verify the LDAP server's
	// certificate, in addition to the system trust store. Set this when the
	// server uses a self-signed or internal-CA cert. Setting it implies
	// TLSVerify.
	TLSCACertFile string `json:"tls_ca_cert_file"`

	// TLSSkipVerify disables verification of the LDAP server's certificate
	// chain and hostname when NeedTLS is set.
	//
	// It defaults to false, so a provider configured under `ldap_providers`
	// verifies the server certificate. An unverified LDAPS connection is open
	// to MITM and the user's password travels over it, since authentication
	// works by binding as the user — so only enable this on a trusted network
	// with a self-signed certificate, and prefer TLSCACertFile instead.
	//
	// The legacy flat ldap_* config is the exception: released Semaphore has
	// never verified LDAPS certificates, so ActiveLdapProviders synthesizes
	// the "ldap" entry with this set unless the operator opts in via
	// ldap_tls_verify. See issue #749.
	TLSSkipVerify bool `json:"tls_skip_verify"`
}

// ShouldVerifyTLS reports whether the LDAP server's certificate must be
// verified. Supplying a CA bundle counts as opting in: configuring a CA only
// makes sense if it is meant to be checked, so it overrides TLSSkipVerify.
func (p LdapProvider) ShouldVerifyTLS() bool {
	return p.TLSCACertFile != "" || !p.TLSSkipVerify
}

// GetMappings returns the attribute mappings, falling back to the same
// defaults the legacy flat config uses (see LdapMappings `default:` tags).
func (p LdapProvider) GetMappings() *LdapMappings {
	if p.Mappings != nil {
		return p.Mappings
	}
	return &LdapMappings{DN: "dn", Mail: "mail", UID: "uid", CN: "cn"}
}

// LdapProviderEntry couples a provider with its stable ID (the config map
// key). The ID is stored in user__external_identity.provider with type='ldap'.
type LdapProviderEntry struct {
	ID       string
	Provider LdapProvider
}

// ActiveLdapProviders returns configured LDAP providers in login-page order:
// the legacy flat ldap_* config first (reserved ID "ldap"), then entries of
// `ldap_providers` sorted by Order (ties broken by ID). A `ldap_providers`
// entry keyed "ldap" is skipped: it would collide with the legacy provider.
func (conf *ConfigType) ActiveLdapProviders() (res []LdapProviderEntry) {
	if conf.LdapEnable {
		res = append(res, LdapProviderEntry{
			ID: "ldap",
			Provider: LdapProvider{
				DisplayName:  "LDAP",
				Server:       conf.LdapServer,
				NeedTLS:      conf.LdapNeedTLS,
				BindDN:       conf.LdapBindDN,
				BindPassword: conf.LdapBindPassword,
				SearchDN:     conf.LdapSearchDN,
				SearchFilter: conf.LdapSearchFilter,
				Mappings:     conf.LdapMappings,
				// Existing installs upgraded into this release have never had
				// their LDAPS certificate verified, and many use a self-signed
				// cert. Verifying by default here would lock them out, so the
				// legacy provider stays unverified until the operator opts in
				// with ldap_tls_verify or supplies a CA bundle. Providers under
				// `ldap_providers` are new config and verify by default.
				TLSSkipVerify: !conf.LdapTLSVerify,
				TLSCACertFile: conf.LdapTLSCACertFile,
			},
		})
	}

	var rest []LdapProviderEntry
	for id, p := range conf.LdapProviders {
		if id == "ldap" {
			continue
		}
		rest = append(rest, LdapProviderEntry{ID: id, Provider: p})
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Provider.Order != rest[j].Provider.Order {
			return rest[i].Provider.Order < rest[j].Provider.Order
		}
		return rest[i].ID < rest[j].ID
	})

	return append(res, rest...)
}

// GetLdapProvider resolves a provider ID coming from a login or link request.
func (conf *ConfigType) GetLdapProvider(id string) (LdapProvider, bool) {
	for _, entry := range conf.ActiveLdapProviders() {
		if entry.ID == id {
			return entry.Provider, true
		}
	}
	return LdapProvider{}, false
}
