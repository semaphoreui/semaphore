package helpers

import (
	"context"
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

// contextKey namespaces the request-scoped values set by these helpers. Using a
// package-private type instead of a bare string keeps the keys from colliding
// with values stored by other packages on the same request context.
type contextKey string

func GetFromContext(r *http.Request, key string) any {
	return r.Context().Value(contextKey(key))
}

func GetOkFromContext(r *http.Request, key string) (res any, ok bool) {
	res = r.Context().Value(contextKey(key))
	return res, res != nil
}

func SetContextValue(r *http.Request, key string, value any) *http.Request {
	ctx := context.WithValue(r.Context(), contextKey(key), value)
	return r.WithContext(ctx)
}

func UserFromContext(r *http.Request) *db.User {
	return GetFromContext(r, "user").(*db.User)
}

func GetGlobalRole(r *http.Request) db.Role {
	return GetFromContext(r, "role").(db.Role)
}
