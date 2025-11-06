package helpers

import (
	"context"
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

func GetFromContext(r *http.Request, key string) any {
	return r.Context().Value(key)
}

func GetOkFromContext(r *http.Request, key string) (res any, ok bool) {
	res = r.Context().Value(key)
	return res, res != nil
}

func SetContextValue(r *http.Request, key string, value any) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, key, value)
	return r.WithContext(ctx)
}

func UserFromContext(r *http.Request) *db.User {
	return GetFromContext(r, "user").(*db.User)
}

func GetGlobalRole(r *http.Request) db.Role {
	return GetFromContext(r, "role").(db.Role)
}
