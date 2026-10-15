package pro_interfaces

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
)

// ErrGroupsClaim means the groups claim in the sign-in claims has a wrong type.
var ErrGroupsClaim = errors.New("groups claim malformed")

// RoleSyncService applies the IdP group role mappings at sign-in.
type RoleSyncService interface {
	// SyncLoginRoles makes the user's mapped roles follow the groups in the sign-in claims.
	SyncLoginRoles(userID int, provider string, claims map[string]any) (added []db.MappedRole, removed []db.MappedRole, err error)
}
