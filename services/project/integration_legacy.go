package project

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

// LegacyIntegration is what a backup written before 2.21.11 said about how
// an integration was reached: through the project aliases (searchable) with
// its own auth, or through its own aliases.
type LegacyIntegration struct {
	Name       string
	Searchable bool
	AuthMethod db.IntegrationAuthMethod
	AuthHeader string
	// AuthSecret is the key by name, as the backup stores it.
	AuthSecret *string
	Aliases    []string
}

// LegacyConversion is the endpoint/group layout equivalent to a legacy
// configuration. GroupOf names the group of each input integration, in input
// order.
type LegacyConversion struct {
	Endpoints []BackupEndpoint
	Groups    []BackupIntegrationGroup
	GroupOf   []string
}

// legacyDefaultEndpoint is the endpoint the project aliases become.
const legacyDefaultEndpoint = "Default"

// isNoAuth treats "none" (the old column default) like "" (what the
// application wrote): both mean no auth.
func isNoAuth(method db.IntegrationAuthMethod) bool {
	return method == db.IntegrationAuthNone || method == "none"
}

// ConvertLegacyIntegrations applies the rules the 2.21.11 database migration
// applies, so an old backup restores to the same layout an upgraded
// installation has:
//
//   - the project aliases, or any searchable integration, make a "Default"
//     endpoint with no auth that holds the project aliases;
//   - every integration gets a group named after it;
//   - a searchable integration joins Default and keeps its own auth as an
//     override (none stays inherited);
//   - any other integration gets an endpoint of its own, named after it,
//     carrying its auth and its aliases, which its group inherits.
//
// The alias of a searchable integration never matched before and is dropped.
// Names are made unique the way the backup loader does, with a " (n)" suffix.
func ConvertLegacyIntegrations(projectAliases []string, integrations []LegacyIntegration) LegacyConversion {
	conv := LegacyConversion{
		Endpoints: make([]BackupEndpoint, 0),
		Groups:    make([]BackupIntegrationGroup, 0),
		GroupOf:   make([]string, 0, len(integrations)),
	}
	endpointNames := uniqueNames{}
	groupNames := uniqueNames{}

	needsDefault := len(projectAliases) > 0
	for _, i := range integrations {
		needsDefault = needsDefault || i.Searchable
	}

	defaultName := ""
	if needsDefault {
		defaultName = endpointNames.take(legacyDefaultEndpoint)
		aliases := make([]string, 0, len(projectAliases))
		aliases = append(aliases, projectAliases...)
		conv.Endpoints = append(conv.Endpoints, BackupEndpoint{
			Endpoint: db.Endpoint{Name: defaultName, Description: "Created on upgrade from the project aliases"},
			Aliases:  aliases,
		})
	}

	for _, i := range integrations {
		group := BackupIntegrationGroup{IntegrationGroup: db.IntegrationGroup{Name: groupNames.take(i.Name)}}

		if i.Searchable {
			group.Endpoint = defaultName
			if !isNoAuth(i.AuthMethod) {
				group.AuthOverride = true
				group.AuthMethod = i.AuthMethod
				group.AuthHeader = i.AuthHeader
				group.AuthSecretName = i.AuthSecret
			}
		} else {
			method := i.AuthMethod
			if isNoAuth(method) {
				method = db.IntegrationAuthNone
			}
			aliases := make([]string, 0, len(i.Aliases))
			aliases = append(aliases, i.Aliases...)
			endpoint := BackupEndpoint{
				Endpoint: db.Endpoint{
					Name:            endpointNames.take(i.Name),
					IntegrationAuth: db.IntegrationAuth{AuthMethod: method, AuthHeader: i.AuthHeader},
				},
				Aliases:        aliases,
				AuthSecretName: i.AuthSecret,
			}
			conv.Endpoints = append(conv.Endpoints, endpoint)
			group.Endpoint = endpoint.Name
		}

		conv.Groups = append(conv.Groups, group)
		conv.GroupOf = append(conv.GroupOf, group.Name)
	}

	return conv
}

// uniqueNames hands out names once; a repeat gets the first free " (n)" suffix.
type uniqueNames map[string]bool

func (u uniqueNames) take(name string) string {
	candidate := name
	for n := 2; u[candidate]; n++ {
		candidate = fmt.Sprintf("%s (%d)", name, n)
	}
	u[candidate] = true
	return candidate
}

// upgradeLegacyIntegrations rewrites a backup written before 2.21.11 — no
// endpoints and no groups, integrations without a group — into the current
// layout. A current backup is left alone, and so is a legacy one already
// converted, so the call is safe to repeat.
func (backup *BackupFormat) upgradeLegacyIntegrations() {
	if len(backup.Endpoints) > 0 || len(backup.IntegrationGroups) > 0 {
		return
	}

	legacy := len(backup.IntegrationAliases) > 0
	for _, i := range backup.Integration {
		legacy = legacy || i.Group == nil
	}
	if !legacy {
		return
	}

	integrations := make([]LegacyIntegration, 0, len(backup.Integration))
	for _, i := range backup.Integration {
		integrations = append(integrations, LegacyIntegration{
			Name:       i.Name,
			Searchable: i.LegacySearchable,
			AuthMethod: i.LegacyAuthMethod,
			AuthHeader: i.LegacyAuthHeader,
			AuthSecret: i.LegacyAuthSecret,
			Aliases:    i.LegacyAliases,
		})
	}

	conv := ConvertLegacyIntegrations(backup.IntegrationAliases, integrations)
	backup.Endpoints = conv.Endpoints
	backup.IntegrationGroups = conv.Groups
	for n := range backup.Integration {
		groupName := conv.GroupOf[n]
		backup.Integration[n].Group = &groupName
	}
	backup.IntegrationAliases = nil
}
