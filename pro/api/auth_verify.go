package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

func VerifySessionByEmail(session *db.Session, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	return
}
