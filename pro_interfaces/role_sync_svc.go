package pro_interfaces

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
)

// ErrGroupsClaim means the groups claim in the sign-in claims has a wrong type.
var ErrGroupsClaim = errors.New("groups claim malformed")

// RoleSyncService applies the IdP group role mappings at sign-in.
type RoleSyncService interface {
	// SyncLoginRoles makes the user's mapped roles follow the groups in the OIDC sign-in claims.
	SyncLoginRoles(userID int, provider string, claims map[string]any) (added []db.MappedRole, removed []db.MappedRole, err error)
	// SyncGroupRoles makes the user's mapped roles follow the given groups. Nil groups mean the provider does not read
	// groups, so nothing changes. An empty list takes every mapped role away, a password sign-in sends one.
	SyncGroupRoles(userID int, provider string, groups []string) (added []db.MappedRole, removed []db.MappedRole, err error)
}
