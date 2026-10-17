package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// effectiveGlobalPermissions is what the user may do instance-wide. An admin holds everything without a
// database read, so the break-glass account works even when the role tables are unreadable.
func effectiveGlobalPermissions(r *http.Request, user *db.User) (db.GlobalPermission, error) {
	if user == nil {
		return 0, nil
	}
	if user.Admin {
		return db.AllGlobalPermissions, nil
	}
	return helpers.Store(r).GetEffectiveGlobalPermissions(user.ID)
}

func hasGlobalPermission(r *http.Request, user *db.User, permission db.GlobalPermission) (bool, error) {
	permissions, err := effectiveGlobalPermissions(r, user)
	if err != nil {
		return false, err
	}
	return permissions.Can(permission), nil
}

// globalPermissionName is the catalog id of the bit, recorded with a denial.
func globalPermissionName(permission db.GlobalPermission) string {
	for _, definition := range db.GlobalPermissionCatalog() {
		if definition.Permission == permission {
			return definition.ID
		}
	}
	return "global_permission"
}

// globalPermissionMiddleware lets through admins and users whose global roles grant the permission.
func globalPermissionMiddleware(permission db.GlobalPermission) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := helpers.GetFromContext(r, "user").(*db.User)
			allowed, err := hasGlobalPermission(r, user, permission)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
			if !allowed {
				helpers.RecordDenied(r, globalPermissionName(permission), 0)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// userMayManageUser says whether the editor may change or delete the target. An admin may change anyone.
// A user who manages users may not touch an admin, nor anyone holding a global permission the editor lacks:
// resetting such a password would be a way up.
func userMayManageUser(r *http.Request, editor *db.User, target db.User) (bool, error) {
	if editor.Admin {
		return true, nil
	}
	editorPermissions, err := effectiveGlobalPermissions(r, editor)
	if err != nil || !editorPermissions.Can(db.CanManageGlobalUsers) {
		return false, err
	}
	if target.Admin {
		return false, nil
	}
	targetPermissions, err := effectiveGlobalPermissions(r, &target)
	if err != nil {
		return false, err
	}
	return targetPermissions&^editorPermissions == 0, nil
}
