package api

import (
	"errors"
	"github.com/gorilla/context"
	"github.com/pquerna/otp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

func getSession(r *http.Request) (*db.Session, bool) {
	// fetch session from cookie
	cookie, err := r.Cookie("semaphore")
	if err != nil {
		return nil, false
	}

	value := make(map[string]any)
	if err = util.Cookie.Decode("semaphore", cookie.Value, &value); err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	user, ok := value["user"]
	sessionVal, okSession := value["session"]
	if !ok || !okSession {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	userID := user.(int)
	sessionID := sessionVal.(int)

	// fetch session
	session, err := helpers.Store(r).GetSession(userID, sessionID)

	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	if time.Since(session.LastActive).Hours() > 7*24 {
		// more than week old unused session
		// destroy.
		if err = helpers.Store(r).ExpireSession(userID, sessionID); err != nil {
			// it is internal error, it doesn't concern the user
			log.Error(err)
		}

		return nil, false
	}

	return &session, true

}

type totpRequestBody struct {
	Passcode string `json:"passcode"`
}

type totpRecoveryRequestBody struct {
	RecoveryCode string `json:"recovery_code"`
}

// recoverySession handles the recovery of a user session using a recovery code.
// It validates the recovery code provided by the user and, if valid, verifies the session.
// If the recovery code is invalid or recovery is not allowed, it returns an appropriate HTTP status code.
//
// HTTP Request:
// - Method: POST
// - Body: JSON object containing the recovery code (e.g., {"recovery_code": "code"}).
//
// Responses:
// - 204 No Content: Recovery successful, session verified.
// - 400 Bad Request: Invalid request body or user does not have TOTP enabled.
// - 401 Unauthorized: Invalid recovery code or session not found.
// - 403 Forbidden: TOTP recovery is disabled.
// - 500 Internal Server Error: An unexpected error occurred.
//
// Preconditions:
// - The session must exist and be valid.
// - TOTP recovery must be enabled in the configuration.
//
// Parameters:
// - w: The HTTP response writer.
// - r: The HTTP request.
func recoverySession(w http.ResponseWriter, r *http.Request) {
	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch session.VerificationMethod {
	case db.SessionVerificationTotp:
		if !util.Config.Auth.Totp.Enabled || !util.Config.Auth.Totp.AllowRecovery {
			helpers.WriteErrorStatus(w, "TOTP_DISABLED", http.StatusForbidden)
			return
		}

		var body totpRecoveryRequestBody
		if !helpers.Bind(w, r, &body) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		store := helpers.Store(r)

		user, err := store.GetUser(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Totp == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !util.VerifyRecoveryCode(body.RecoveryCode, user.Totp.RecoveryHash) {
			helpers.WriteErrorStatus(w, "INVALID_RECOVERY_CODE", http.StatusUnauthorized)
			return
		}

		err = store.DeleteTotpVerification(user.ID, user.Totp.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = store.VerifySession(session.UserID, session.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	case db.SessionVerificationNone:
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func verifySession(w http.ResponseWriter, r *http.Request) {
	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch session.VerificationMethod {
	case db.SessionVerificationEmail:
		verifySessionByEmail(w, r)
		return

	case db.SessionVerificationTotp:
		if !util.Config.Auth.Totp.Enabled {
			helpers.WriteErrorStatus(w, "TOTP_DISABLED", http.StatusForbidden)
			return
		}

		var body totpRequestBody
		if !helpers.Bind(w, r, &body) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := helpers.Store(r).GetUser(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		key, err := otp.NewKeyFromURL(user.Totp.URL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !totp.Validate(body.Passcode, key.Secret()) {
			helpers.WriteErrorStatus(w, "INVALID_PASSCODE", http.StatusUnauthorized)
			return
		}

		err = helpers.Store(r).VerifySession(session.UserID, session.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

	case db.SessionVerificationNone:
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) bool {
	var userID int

	authHeader := strings.ToLower(r.Header.Get("authorization"))

	if len(authHeader) > 0 && strings.Contains(authHeader, "bearer") {
		token, err := helpers.Store(r).GetAPIToken(strings.Replace(authHeader, "bearer ", "", 1))

		if err != nil {
			if !errors.Is(err, db.ErrNotFound) {
				log.Error(err)
			}

			w.WriteHeader(http.StatusUnauthorized)
			return false
		}

		userID = token.UserID
	} else {
		session, ok := getSession(r)

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}

		if !session.IsVerified() {
			helpers.WriteErrorStatus(w, "TOTP_REQUIRED", http.StatusUnauthorized)
			return false
		}

		userID = session.UserID

		if err := helpers.Store(r).TouchSession(userID, session.ID); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
	}

	user, err := helpers.Store(r).GetUser(userID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			// internal error
			log.Error(err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	context.Set(r, "user", &user)
	return true
}

// nolint: gocyclo
func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := authenticationHandler(w, r)
		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

// nolint: gocyclo
func authenticationWithStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := helpers.Store(r)

		var ok bool

		db.StoreSession(store, r.URL.String(), func() {
			ok = authenticationHandler(w, r)
		})

		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.Get(r, "user").(*db.User)

		if !user.Admin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
