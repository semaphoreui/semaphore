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

type auditRequestContextKey struct{}

// AuditRequestContext contains normalized request metadata for audit mapping.
type AuditRequestContext struct {
	RequestID string
	SourceIP  string
	UserAgent string
}

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

func AuditRequestContextFrom(r *http.Request) (AuditRequestContext, bool) {
	value, ok := r.Context().Value(auditRequestContextKey{}).(AuditRequestContext)
	return value, ok
}

func SetAuditRequestContext(r *http.Request, value AuditRequestContext) *http.Request {
	ctx := context.WithValue(r.Context(), auditRequestContextKey{}, value)
	return r.WithContext(ctx)
}

func UserFromContext(r *http.Request) *db.User {
	return GetFromContext(r, "user").(*db.User)
}

func GetGlobalRole(r *http.Request) db.Role {
	return GetFromContext(r, "role").(db.Role)
}
