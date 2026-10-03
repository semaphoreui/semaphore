package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bearerRequest(store db.Store, token string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	return helpers.SetContextValue(r, "store", store)
}

func TestAuthentication_UnknownAPITokenIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	r, rec := withAuditRecorder(bearerRequest(store, "unknowntoken12345"))
	w := httptest.NewRecorder()

	ok, _ := authenticationHandler(w, r)

	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	got := onlyEvent(t, rec, audit.AuthAPITokenReject)
	assert.Equal(t, audit.ReasonTokenUnknown, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetAPIToken, ID: audit.TokenFingerprint("unknowntoken12345")}, got.Event.Target)
	assert.Equal(t, audit.AnonymousActor(), got.Actor)
}

func TestAuthentication_ExpiredAPITokenIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	past := time.Now().Add(-time.Hour)
	_, err := store.CreateAPIToken(db.APIToken{ID: "expiredtoken1234567890", UserID: user.ID, ExpiresAt: &past, Name: "ci"})
	require.NoError(t, err)

	r, rec := withAuditRecorder(bearerRequest(store, "expiredtoken1234567890"))
	ok, _ := authenticationHandler(httptest.NewRecorder(), r)

	assert.False(t, ok)
	got := onlyEvent(t, rec, audit.AuthAPITokenReject)
	assert.Equal(t, audit.ReasonTokenExpired, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetAPIToken, ID: audit.TokenFingerprint("expiredtoken1234567890")}, got.Event.Target,
		"only the fingerprint, the name is found through iam.api_token/create")
}

func TestAuthentication_APITokenActor(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateAPIToken(db.APIToken{ID: "validtoken1234567890", UserID: user.ID, Name: "ci"})
	require.NoError(t, err)

	r, rec := withAuditRecorder(bearerRequest(store, "validtoken1234567890"))
	ok, req := authenticationHandler(httptest.NewRecorder(), r)

	require.True(t, ok)
	assert.Empty(t, rec.All())
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthAPIToken, audit.TokenFingerprint("validtoken1234567890")), audit.ActorFrom(req.Context()))
}

func TestAuthentication_ActorReachesHandler(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/user", nil), "store", store)
	r = addSessionCookie(t, store, r, user, db.SessionVerificationNone)
	r, rec := withAuditRecorder(r)

	authentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.Audit(r).Record(r.Context(), audit.Event{Kind: audit.IAMUserDelete})
	})).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), onlyEvent(t, rec, audit.IAMUserDelete).Actor)
}

func TestAdminMiddleware_RecordsDenial(t *testing.T) {
	router := mux.NewRouter()
	router.Handle("/api/options", adminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))).Methods(http.MethodPost)

	r, rec := withAuditRecorder(asActor(httptest.NewRequest(http.MethodPost, "/api/options", nil), db.User{ID: 7, Username: "bob"}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
	assert.Equal(t, audit.ReasonForbidden, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetRoute, ID: "/api/options"}, got.Event.Target)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: "admin"}, got.Event.Metadata)
	assert.Equal(t, "7", got.Actor.ID)
}

func TestGetUserMiddleware_RecordsDenialOnChanges(t *testing.T) {
	store := setupSessionTest(t)
	target := createUserOptionsTestUser(t, store, "alice")
	intruder := createUserOptionsTestUser(t, store, "mallory")

	router := mux.NewRouter()
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.Handle("/api/users/{user_id}", NewUsersController(nil).GetUserMiddleware(ok)).Methods(http.MethodPut, http.MethodGet)

	tests := []struct {
		method     string
		wantEvents int
	}{
		{http.MethodPut, 1},
		{http.MethodGet, 0},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := helpers.SetContextValue(httptest.NewRequest(tt.method, fmt.Sprintf("/api/users/%d", target.ID), nil), "store", store)
			r, rec := withAuditRecorder(asActor(r, intruder))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "the response does not change")
			require.Len(t, rec.All(), tt.wantEvents)
			if tt.wantEvents == 1 {
				got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
				assert.Equal(t, &audit.Target{Type: audit.TargetRoute, ID: "/api/users/{user_id}"}, got.Event.Target)
				assert.Equal(t, audit.DenyMetadata{Method: http.MethodPut, Permission: "admin"}, got.Event.Metadata)
				assert.Equal(t, "mallory", got.Actor.Name)
			}
		})
	}
}

func TestAdminMiddleware_AdminIsNotRecorded(t *testing.T) {
	r, rec := withAuditRecorder(asActor(httptest.NewRequest(http.MethodPost, "/api/options", nil), db.User{ID: 1, Admin: true}))
	adminMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(httptest.NewRecorder(), r)
	assert.Empty(t, rec.All())
}

func TestCsrfProtectionMiddleware_RecordsBlock(t *testing.T) {
	router := mux.NewRouter()
	router.Handle("/api/project/{project_id}/keys", csrfProtectionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))).Methods(http.MethodPost)

	r := httptest.NewRequest(http.MethodPost, "/api/project/1/keys", nil)
	r.Host = "semaphore.example.com"
	r.Header.Set("Origin", "https://attacker.example")
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	got := onlyEvent(t, rec, audit.AuthCSRFBlock)
	assert.Equal(t, audit.ReasonCrossOrigin, got.Event.Reason)
	assert.Equal(t, &audit.Target{Type: audit.TargetRoute, ID: "/api/project/{project_id}/keys"}, got.Event.Target)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost}, got.Event.Metadata)
}
