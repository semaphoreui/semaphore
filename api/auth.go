package api

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	proApi "github.com/semaphoreui/semaphore/pro/api"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"

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
		if !util.Config.Mfa.Totp.Enabled || !util.Config.Mfa.Totp.AllowRecovery {
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
		proApi.VerifySessionByEmail(session, w, r)
		return

	case db.SessionVerificationTotp:
		if !util.Config.Mfa.Totp.Enabled {
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

		if user.Totp == nil {
			w.WriteHeader(http.StatusBadRequest)
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

func authenticationHandler(w http.ResponseWriter, r *http.Request) (ok bool, req *http.Request) {
	var userID int

	req = r

	authHeader := strings.ToLower(r.Header.Get("authorization"))

	if len(authHeader) > 0 && strings.Contains(authHeader, "bearer") {
		token, err := helpers.Store(r).GetAPIToken(strings.Replace(authHeader, "bearer ", "", 1))

		if err != nil {
			if !errors.Is(err, db.ErrNotFound) {
				log.Error(err)
			}

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if token.IsExpiredAt(tz.Now()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID = token.UserID
	} else {
		session, found := getSession(r)

		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !session.IsVerified() {
			switch session.VerificationMethod {
			case db.SessionVerificationEmail:
				helpers.WriteErrorStatus(w, "EMAIL_OTP_REQUIRED", http.StatusUnauthorized)
			case db.SessionVerificationTotp:
				helpers.WriteErrorStatus(w, "TOTP_REQUIRED", http.StatusUnauthorized)
			default:
				helpers.WriteErrorStatus(w, "SESSION_NOT_VERIFIED", http.StatusUnauthorized)
			}
			return
		}

		userID = session.UserID

		if err := helpers.Store(r).TouchSession(userID, session.ID); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	user, err := helpers.Store(r).GetUser(userID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			// internal error
			log.Error(err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ok = true
	req = helpers.SetContextValue(r, "user", &user)
	return
}

// nolint: gocyclo
func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, r := authenticationHandler(w, r)
		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

// nolint: gocyclo
func authenticationWithStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool

		ok, r = authenticationHandler(w, r)

		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetFromContext(r, "user").(*db.User)

		if !user.Admin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func metricsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password := "", ""
		if util.Config.Metrics != nil {
			username = util.Config.Metrics.Username
			password = util.Config.Metrics.Password
		}

		reqUser, reqPass, ok := r.BasicAuth()
		userMatch := subtle.ConstantTimeCompare([]byte(reqUser), []byte(username)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(reqPass), []byte(password)) == 1

		if username == "" || password == "" || !ok || !userMatch || !passMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isStateChangingMethod reports whether an HTTP method can modify server state
// and therefore requires CSRF protection. Safe methods (GET, HEAD, OPTIONS,
// TRACE) are excluded.
func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// requestOriginHost extracts the origin host (host[:port]) of the request from
// the Origin header, falling back to the Referer header. The boolean is false
// when neither header is present or parseable.
func requestOriginHost(r *http.Request) (string, bool) {
	for _, header := range []string{"Origin", "Referer"} {
		value := r.Header.Get(header)
		if value == "" {
			continue
		}

		u, err := url.Parse(value)
		if err != nil || u.Host == "" {
			continue
		}

		return u.Host, true
	}

	return "", false
}

// isSameOriginHost reports whether host belongs to Semaphore itself. Both the
// configured public web host and the host the request was addressed to are
// accepted, so reverse-proxy deployments keep working.
func isSameOriginHost(host string, r *http.Request) bool {
	if host == r.Host {
		return true
	}

	if util.WebHostURL != nil && host == util.WebHostURL.Host {
		return true
	}

	return false
}

// csrfProtectionMiddleware blocks cross-site state-changing requests that rely
// on the session cookie, providing defense-in-depth against CSRF on top of the
// SameSite=Lax session cookie.
//
// Requests authenticated with an API token (Authorization: bearer) are exempt:
// browsers never attach such tokens automatically, so token-based clients are
// not vulnerable to CSRF and must keep working without an Origin header.
//
// When neither Origin nor Referer is present (e.g. non-browser clients using a
// cookie), the request is allowed — the SameSite=Lax cookie already prevents a
// browser from sending the session cookie cross-site in that case.
func csrfProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isStateChangingMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := strings.ToLower(r.Header.Get("authorization"))
		if strings.Contains(authHeader, "bearer") {
			next.ServeHTTP(w, r)
			return
		}

		if origin, ok := requestOriginHost(r); ok && !isSameOriginHost(origin, r) {
			log.WithFields(log.Fields{
				"origin": origin,
				"host":   r.Host,
				"path":   r.URL.Path,
				"method": r.Method,
			}).Warn("Blocked cross-origin request (possible CSRF)")
			helpers.WriteErrorStatus(w, "CROSS_ORIGIN_REQUEST_BLOCKED", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
