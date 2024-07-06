package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
)

func getOptions(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)

	if !currentUser.Admin {
		helpers.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error": "User must be admin",
		})
		return
	}
}
