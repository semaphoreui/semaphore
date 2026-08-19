package util

import "sort"

// LdapProvider is one LDAP directory configured under `ldap_providers`
// (analog of OidcProvider for `oidc_providers`). Field JSON names mirror
// the legacy flat ldap_* config options.
type LdapProvider struct {
	DisplayName string `json:"display_name"`
	Server      string `json:"server"`
	NeedTLS     bool   `json:"need_tls"`
	// TLSSkipVerify disables verification of the LDAP server's TLS certificate.
	// It defaults to false so certificates are verified: a network attacker
	// cannot impersonate the LDAP server to capture bind or user credentials.
	// Only enable it for trusted networks with self-signed certificates.
	TLSSkipVerify bool          `json:"tls_skip_verify"`
	BindDN        string        `json:"bind_dn"`
	BindPassword  string        `json:"bind_password"`
	SearchDN      string        `json:"search_dn"`
	SearchFilter  string        `json:"search_filter"`
	Mappings      *LdapMappings `json:"mappings"`
	Color         string        `json:"color"`
	Icon          string        `json:"icon"`
	Order         int           `json:"order"`
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
				DisplayName:   "LDAP",
				Server:        conf.LdapServer,
				NeedTLS:       conf.LdapNeedTLS,
				TLSSkipVerify: conf.LdapTLSSkipVerify,
				BindDN:        conf.LdapBindDN,
				BindPassword:  conf.LdapBindPassword,
				SearchDN:      conf.LdapSearchDN,
				SearchFilter:  conf.LdapSearchFilter,
				Mappings:      conf.LdapMappings,
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
