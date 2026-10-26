package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

// validInternalCaller checks an in-process call (helpers.InternalCaller) before
// authentication acts on it. Such a request carries no credential of its own:
// one that does was built wrong, for example with the client's headers copied.
func validInternalCaller(r *http.Request, caller helpers.InternalCaller) bool {
	_, cookieErr := r.Cookie("semaphore")
	hasCredential := r.Header.Get("Authorization") != "" || cookieErr == nil

	// Sign-in methods of components allowed to call in-process.
	if !hasCredential && caller.UserID > 0 && caller.Auth == audit.AuthMCPToken {
		return true
	}

	log.WithFields(log.Fields{
		"context":        "auth",
		"user_id":        caller.UserID,
		"auth":           caller.Auth,
		"has_credential": hasCredential,
	}).Error("Refused an in-process call with an invalid caller")
	return false
}

// denyInternalCallerMiddleware closes a route to in-process calls: routes outside
// the authenticated API and routes that need the user's own sign-in (tokens).
func denyInternalCallerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, internal := helpers.InternalCallerFrom(r.Context()); internal {
			helpers.RecordDenied(r, "direct_sign_in", 0)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
