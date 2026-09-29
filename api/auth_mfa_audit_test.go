package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// "{passcode}" and "{recovery}" in body are filled in. The caller adds the session cookie.
func totpSession(t *testing.T, path string, body string) (*http.Request, db.User, *sql.SqlDb) {
	t.Helper()
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Semaphore", AccountName: user.Email})
	require.NoError(t, err)
	code, hash, err := util.GenerateRecoveryCode()
	require.NoError(t, err)
	_, err = store.AddTotpVerification(user.ID, key.URL(), hash)
	require.NoError(t, err)
	passcode, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	body = strings.NewReplacer("{passcode}", passcode, "{recovery}", code).Replace(body)
	r := helpers.SetContextValue(httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)), "store", store)
	return r, user, store
}

func addCookiesFrom(r *http.Request, w *httptest.ResponseRecorder) *http.Request {
	for _, cookie := range w.Result().Cookies() {
		r.AddCookie(cookie)
	}
	return r
}

func TestLogin_WithTotpRecordsLoginAfterPasscode(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	loginReq, loginRec := withAuditRecorder(loginRequest(store, `{"auth":"alice","password":"verystrongpassword1","method":"password"}`))
	loginW := httptest.NewRecorder()

	login(loginW, loginReq)

	require.Equal(t, http.StatusNoContent, loginW.Code)
	assert.Empty(t, loginRec.All(), "no login before the second factor")

	r, rec := withAuditRecorder(addCookiesFrom(r, loginW))
	verifySession(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	require.Equal(t, []audit.Kind{audit.AuthMFAVerifyTOTP, audit.AuthLogin}, kinds)
	got := rec.All()[1]
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), got.Actor)
	assert.Equal(t, audit.UserTarget(user.ID, "alice"), got.Event.Target)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodPassword}, got.Event.Metadata)
}

// A verified session can post a valid code again, the login is not recorded twice.
func TestLogin_WithTotpSecondVerifyRecordsNoLogin(t *testing.T) {
	r, _, store := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	loginW := httptest.NewRecorder()
	login(loginW, loginRequest(store, `{"auth":"alice","password":"verystrongpassword1","method":"password"}`))

	verifySession(httptest.NewRecorder(), addCookiesFrom(helpers.SetContextValue(httptest.NewRequest(http.MethodPost, "/api/auth/verify", bytes.NewReader(body)), "store", store), loginW))
	r, rec := withAuditRecorder(addCookiesFrom(helpers.SetContextValue(httptest.NewRequest(http.MethodPost, "/api/auth/verify", bytes.NewReader(body)), "store", store), loginW))
	verifySession(httptest.NewRecorder(), r)

	onlyEvent(t, rec, audit.AuthMFAVerifyTOTP)
}

func TestLogin_WithTotpWrongPasscodeRecordsNoLogin(t *testing.T) {
	r, _, store := totpSession(t, "/api/auth/verify", `{"passcode":"abcdef"}`)
	loginW := httptest.NewRecorder()
	login(loginW, loginRequest(store, `{"auth":"alice","password":"verystrongpassword1","method":"password"}`))

	r, rec := withAuditRecorder(addCookiesFrom(r, loginW))
	verifySession(httptest.NewRecorder(), r)

	assert.Equal(t, audit.ReasonInvalidPasscode, onlyEvent(t, rec, audit.AuthMFAVerifyTOTP).Event.Reason)
}

func TestLogin_WithRecoveryCodeRecordsLogin(t *testing.T) {
	r, _, store := totpSession(t, "/api/auth/recovery", `{"recovery_code":"{recovery}"}`)
	loginW := httptest.NewRecorder()
	login(loginW, loginRequest(store, `{"auth":"alice","password":"verystrongpassword1","method":"password"}`))

	r, rec := withAuditRecorder(addCookiesFrom(r, loginW))
	recoverySession(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	require.Equal(t, []audit.Kind{audit.AuthMFARecover, audit.AuthLogin}, kinds)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodPassword}, rec.All()[1].Event.Metadata)
}

// The OIDC redirect needs a live provider, so its session is created directly.
func TestCreateSession_OidcWithTotpRecordsLoginAfterPasscode(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	meta := audit.AuthMethodMetadata{Method: audit.LoginMethodOIDC, Provider: "corp"}
	user, err := store.GetUser(user.ID)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	verified, err := createSession(w, helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/", nil), "store", store), user, meta)
	require.NoError(t, err)
	assert.False(t, verified)

	r, rec := withAuditRecorder(addCookiesFrom(r, w))
	verifySession(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	require.Equal(t, []audit.Kind{audit.AuthMFAVerifyTOTP, audit.AuthLogin}, kinds)
	assert.Equal(t, meta, rec.All()[1].Event.Metadata)
}

// A cookie set before the upgrade has no login method, the login is still recorded.
func TestVerifySession_TotpSuccessIsRecorded(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	r, rec := withAuditRecorder(addSessionCookie(t, store, r, user, db.SessionVerificationTotp))

	verifySession(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	require.Equal(t, []audit.Kind{audit.AuthMFAVerifyTOTP, audit.AuthLogin}, kinds)
	got := rec.All()[0]
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), got.Actor)
	assert.Equal(t, audit.AuthMethodMetadata{}, rec.All()[1].Event.Metadata)
}

func TestVerifySession_WrongPasscodeIsRecorded(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/verify", `{"passcode":"abcdef"}`)
	r, rec := withAuditRecorder(addSessionCookie(t, store, r, user, db.SessionVerificationTotp))
	w := httptest.NewRecorder()

	verifySession(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.ReasonInvalidPasscode, onlyEvent(t, rec, audit.AuthMFAVerifyTOTP).Event.Reason)
}

// Only a checked code is an MFA failure, a disabled method is not.
func TestVerifySession_DisabledMethodIsNotRecorded(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	util.Config.Mfa.Totp.Enabled = false
	r, rec := withAuditRecorder(addSessionCookie(t, store, r, user, db.SessionVerificationTotp))
	w := httptest.NewRecorder()

	verifySession(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, rec.All())
}

func TestRecoverySession_IsRecorded(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/recovery", `{"recovery_code":"{recovery}"}`)
	r, rec := withAuditRecorder(addSessionCookie(t, store, r, user, db.SessionVerificationTotp))

	recoverySession(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	require.Equal(t, []audit.Kind{audit.AuthMFARecover, audit.AuthLogin}, kinds)
	assert.Equal(t, audit.OutcomeSuccess, rec.All()[0].Event.Outcome)
}

func TestRecoverySession_WrongCodeIsRecorded(t *testing.T) {
	r, user, store := totpSession(t, "/api/auth/recovery", `{"recovery_code":"wrong"}`)
	r, rec := withAuditRecorder(addSessionCookie(t, store, r, user, db.SessionVerificationTotp))
	w := httptest.NewRecorder()

	recoverySession(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.ReasonInvalidRecoveryCode, onlyEvent(t, rec, audit.AuthMFARecover).Event.Reason)
}
