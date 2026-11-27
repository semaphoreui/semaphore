package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// globalRoleStoreMiddleware puts the edition's global role store on the request, where the permission checks
// below find it. The community store answers with no permissions.
func globalRoleStoreMiddleware(globalRoles db.GlobalRoleStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, helpers.SetContextValue(r, "global_roles", globalRoles))
		})
	}
}

// effectiveGlobalPermissions is what the user may do instance-wide. An admin holds everything without a
// store read, so the break-glass account works even when the role tables are unreadable. Without a global
// role store on the request only the admin flag counts.
func effectiveGlobalPermissions(r *http.Request, user *db.User) (db.GlobalPermission, error) {
	if user == nil {
		return 0, nil
	}
	if user.Admin {
		return db.AllGlobalPermissions, nil
	}
	globalRoles, ok := helpers.GetFromContext(r, "global_roles").(db.GlobalRoleStore)
	if !ok || globalRoles == nil {
		return 0, nil
	}
	return globalRoles.GetEffectiveGlobalPermissions(user.ID)
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
