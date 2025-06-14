package helpers

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/semaphoreui/semaphore/db"
)

func UserFromContext(r *http.Request) *db.User {
	return context.Get(r, "user").(*db.User)
}
