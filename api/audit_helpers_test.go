package api

import (
	"net/http"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/require"
)

func withAuditRecorder(r *http.Request) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	return helpers.SetContextValue(r, "audit", rec), rec
}

func onlyEvent(t *testing.T, rec *audittest.Recorder, kind audit.Kind) audittest.Recorded {
	t.Helper()
	got, err := rec.Only(kind)
	require.NoError(t, err)
	return got
}

func asActor(r *http.Request, user db.User) *http.Request {
	r = helpers.SetContextValue(r, "user", &user)
	return r.WithContext(audit.WithActor(r.Context(), audit.UserActor(user.ID, user.Username, audit.AuthSession, "")))
}

func setupSessionTest(t *testing.T) *sql.SqlDb {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	util.Config.Mfa = &util.MultifactorAuthConfig{Totp: &util.TotpConfig{Enabled: true, AllowRecovery: true}}
	util.Cookie = securecookie.New(securecookie.GenerateRandomKey(32), nil)
	return store
}

func addSessionCookie(t *testing.T, store db.Store, r *http.Request, user db.User, method db.SessionVerificationMethod) *http.Request {
	t.Helper()
	session, err := store.CreateSession(db.Session{
		UserID:             user.ID,
		Created:            tz.Now(),
		LastActive:         tz.Now(),
		VerificationMethod: method,
		Verified:           method == db.SessionVerificationNone,
	})
	require.NoError(t, err)
	encoded, err := util.Cookie.Encode("semaphore", map[string]any{"user": user.ID, "session": session.ID})
	require.NoError(t, err)
	r.AddCookie(&http.Cookie{Name: "semaphore", Value: encoded})
	return r
}
