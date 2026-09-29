package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// "{passcode}" and "{recovery}" in body are filled in.
func totpSession(t *testing.T, path string, body string) (*http.Request, db.User, string) {
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
	return addSessionCookie(t, store, r, user, db.SessionVerificationTotp), user, code
}

func TestVerifySession_TotpSuccessIsRecorded(t *testing.T) {
	r, user, _ := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	r, rec := withAuditRecorder(r)

	verifySession(httptest.NewRecorder(), r)

	got := onlyEvent(t, rec, audit.AuthMFAVerifyTOTP)
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), got.Actor)
}

func TestVerifySession_WrongPasscodeIsRecorded(t *testing.T) {
	r, _, _ := totpSession(t, "/api/auth/verify", `{"passcode":"abcdef"}`)
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()

	verifySession(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.ReasonInvalidPasscode, onlyEvent(t, rec, audit.AuthMFAVerifyTOTP).Event.Reason)
}

// Only a checked code is an MFA failure, a disabled method is not.
func TestVerifySession_DisabledMethodIsNotRecorded(t *testing.T) {
	r, _, _ := totpSession(t, "/api/auth/verify", `{"passcode":"{passcode}"}`)
	util.Config.Mfa.Totp.Enabled = false
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()

	verifySession(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, rec.All())
}

func TestRecoverySession_IsRecorded(t *testing.T) {
	r, _, _ := totpSession(t, "/api/auth/recovery", `{"recovery_code":"{recovery}"}`)
	r, rec := withAuditRecorder(r)

	recoverySession(httptest.NewRecorder(), r)

	assert.Equal(t, audit.OutcomeSuccess, onlyEvent(t, rec, audit.AuthMFARecover).Event.Outcome)
}

func TestRecoverySession_WrongCodeIsRecorded(t *testing.T) {
	r, _, _ := totpSession(t, "/api/auth/recovery", `{"recovery_code":"wrong"}`)
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()

	recoverySession(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.ReasonInvalidRecoveryCode, onlyEvent(t, rec, audit.AuthMFARecover).Event.Reason)
}
