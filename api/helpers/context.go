package helpers

import (
	"github.com/gorilla/context"
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

func GetFromContext(r *http.Request, key string) any {
	return context.Get(r, key)
}

func SetContextValue(r *http.Request, key string, value any) {
	context.Set(r, key, value)
}

func UserFromContext(r *http.Request) *db.User {
	return GetFromContext(r, "user").(*db.User)
}
