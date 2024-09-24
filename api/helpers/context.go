package helpers

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
	"github.com/gorilla/context"
)

func UserFromContext(r *http.Request) *db.User {
	return context.Get(r, "user").(*db.User)
}
