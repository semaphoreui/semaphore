package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const internalCallerFingerprint = "0123456789abcdef"

func internalCallerFor(user db.User) helpers.InternalCaller {
	return helpers.InternalCaller{UserID: user.ID, Auth: audit.AuthMCPToken, TokenFingerprint: internalCallerFingerprint}
}

// internalRequest is an in-process call as a server component builds it: no header, no cookie,
// the caller only in the context.
func internalRequest(store db.Store, method, path string, caller helpers.InternalCaller) *http.Request {
	r := helpers.SetContextValue(httptest.NewRequest(method, path, nil), "store", store)
	return r.WithContext(helpers.WithInternalCaller(r.Context(), caller))
}

func TestInternalCaller_OnlyInContext(t *testing.T) {
	_, ok := helpers.InternalCallerFrom(httptest.NewRequest(http.MethodGet, "/api/user", nil).Context())
	assert.False(t, ok, "a request from the network carries no in-process caller")

	caller := helpers.InternalCaller{UserID: 7, Auth: audit.AuthMCPToken, TokenFingerprint: "f"}
	got, ok := helpers.InternalCallerFrom(helpers.WithInternalCaller(t.Context(), caller))
	require.True(t, ok)
	assert.Equal(t, caller, got)
}

func TestAuthentication_InternalCallerActsAsUser(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, rec := withAuditRecorder(internalRequest(store, http.MethodGet, "/api/user", internalCallerFor(user)))
	ok, req := authenticationHandler(httptest.NewRecorder(), r)

	require.True(t, ok)
	assert.Equal(t, user.ID, helpers.UserFromContext(req).ID)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthMCPToken, internalCallerFingerprint), audit.ActorFrom(req.Context()))
	assert.Empty(t, rec.All(), "a successful sign-in records nothing, as with an API token")
}

func TestAuthentication_InternalCallerReadsUserEveryTime(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	renamed := user
	renamed.Name = "Alice Renamed"
	require.NoError(t, store.UpdateUser(db.UserWithPwd{User: renamed}))

	ok, req := authenticationHandler(httptest.NewRecorder(), internalRequest(store, http.MethodGet, "/api/user", internalCallerFor(user)))
	require.True(t, ok)
	assert.Equal(t, "Alice Renamed", helpers.UserFromContext(req).Name, "the user comes from the store, not from the caller")

	require.NoError(t, store.DeleteUser(user.ID))
	w := httptest.NewRecorder()
	ok, _ = authenticationHandler(w, internalRequest(store, http.MethodGet, "/api/user", internalCallerFor(user)))
	assert.False(t, ok, "a deleted user loses access at once")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthentication_InternalCallerRefused(t *testing.T) {
	store := setupSessionTest(t)
	alice := createUserOptionsTestUser(t, store, "alice")
	bob := createUserOptionsTestUser(t, store, "bob")
	_, err := store.CreateAPIToken(db.APITokenTypeAPI, db.APIToken{ID: "bobtoken1234567890", UserID: bob.ID, Name: "ci"})
	require.NoError(t, err)

	tests := []struct {
		name    string
		caller  helpers.InternalCaller
		prepare func(r *http.Request) *http.Request
	}{
		{name: "no user", caller: helpers.InternalCaller{Auth: audit.AuthMCPToken}},
		{name: "unknown user", caller: helpers.InternalCaller{UserID: 9999, Auth: audit.AuthMCPToken}},
		{name: "no auth method", caller: helpers.InternalCaller{UserID: alice.ID}},
		{name: "claims a session", caller: helpers.InternalCaller{UserID: alice.ID, Auth: audit.AuthSession}},
		{name: "claims an API token", caller: helpers.InternalCaller{UserID: alice.ID, Auth: audit.AuthAPIToken}},
		{name: "unlisted auth method", caller: helpers.InternalCaller{UserID: alice.ID, Auth: audit.AuthMethod("webhook")}},
		{
			name:   "carries another user's API token",
			caller: internalCallerFor(alice),
			prepare: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Bearer bobtoken1234567890")
				return r
			},
		},
		{
			name:   "carries another user's session",
			caller: internalCallerFor(alice),
			prepare: func(r *http.Request) *http.Request {
				return addSessionCookie(t, store, r, bob, db.SessionVerificationNone)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := internalRequest(store, http.MethodGet, "/api/user", tt.caller)
			if tt.prepare != nil {
				r = tt.prepare(r)
			}
			w := httptest.NewRecorder()

			ok, req := authenticationHandler(w, r)

			assert.False(t, ok)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			_, hasUser := helpers.GetOkFromContext(req, "user")
			assert.False(t, hasUser)
		})
	}
}

func TestDenyInternalCallerMiddleware(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	t.Run("refuses an in-process call", func(t *testing.T) {
		r, rec := withAuditRecorder(internalRequest(store, http.MethodGet, "/api/user/tokens", internalCallerFor(user)))
		called := false
		w := httptest.NewRecorder()

		denyInternalCallerMiddleware(newRecordingHandler(&called)).ServeHTTP(w, r)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
		got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
		assert.Equal(t, audit.DenyMetadata{Method: http.MethodGet, Permission: "direct_sign_in"}, got.Event.Metadata)
	})

	t.Run("lets a direct request through", func(t *testing.T) {
		called := false
		w := httptest.NewRecorder()

		denyInternalCallerMiddleware(newRecordingHandler(&called)).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/user/tokens", nil))

		assert.True(t, called)
	})
}

// TestRouter_InternalCaller drives the real router, as an MCP tool does in-process.
func TestRouter_InternalCaller(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	send := func(method, path string, user db.User) (*httptest.ResponseRecorder, []audit.Event) {
		r, rec := withAuditRecorder(internalRequest(f.store, method, path, internalCallerFor(user)))
		r = helpers.SetContextValue(r, "task_pool", f.pool)
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, r)
		var events []audit.Event
		for _, recorded := range rec.All() {
			events = append(events, recorded.Event)
		}
		return w, events
	}

	t.Run("acts as the user", func(t *testing.T) {
		_, err := f.store.CreateProject(db.Project{Name: "only-admins-see-it"})
		require.NoError(t, err)

		w, _ := send(http.MethodGet, "/api/projects", f.root)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "only-admins-see-it", "an administrator lists every project")

		w, _ = send(http.MethodGet, "/api/projects", f.plain)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "only-admins-see-it", "a user who is not a member does not")
	})

	t.Run("keeps the user's own permissions", func(t *testing.T) {
		w, events := send(http.MethodGet, "/api/options", f.plain)
		assert.Equal(t, http.StatusForbidden, w.Code, "a plain user may not read server options through REST")
		require.Len(t, events, 1)
		assert.Equal(t, audit.AuthAuthorizationDeny, events[0].Kind)

		w, _ = send(http.MethodGet, "/api/options", f.root)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("never reaches the token routes", func(t *testing.T) {
		for _, route := range []struct{ method, path string }{
			{http.MethodGet, "/api/user/tokens"},
			{http.MethodPost, "/api/user/tokens"},
			{http.MethodDelete, "/api/user/tokens/abcdefgh"},
			{http.MethodGet, "/api/user/mcp-tokens"},
			{http.MethodPost, "/api/user/mcp-tokens"},
			{http.MethodDelete, "/api/user/mcp-tokens/abcdefgh"},
		} {
			w, events := send(route.method, route.path, f.root)
			assert.Equal(t, http.StatusForbidden, w.Code, route.method+" "+route.path)
			require.Len(t, events, 1, route.method+" "+route.path)
			assert.Equal(t, audit.AuthAuthorizationDeny, events[0].Kind)
		}
	})

	t.Run("still reaches the user settings", func(t *testing.T) {
		w, _ := send(http.MethodGet, "/api/user/options", f.plain)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// The split of the /user routes must not close the settings to a direct request.
func TestRouter_UserRoutesForDirectRequests(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	for _, path := range []string{"/api/user/tokens", "/api/user/options"} {
		r := addSessionCookie(t, f.store, httptest.NewRequest(http.MethodGet, path, nil), f.plain, db.SessionVerificationNone)
		r = helpers.SetContextValue(r, "store", f.store)
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code, path)
	}
}

// sendRouter sends one request to the real router, in-process for caller or, without one, as an anonymous
// client. A handler that panics on the nil services Route is built with here was reached all the same.
func (f globalRoutesFixture) sendRouter(method, path string, caller *helpers.InternalCaller) (code int, reached bool, denied bool) {
	r := helpers.SetContextValue(httptest.NewRequest(method, path, nil), "store", f.store)
	r = helpers.SetContextValue(r, "task_pool", f.pool)
	if caller != nil {
		r = r.WithContext(helpers.WithInternalCaller(r.Context(), *caller))
	}
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()
	func() {
		defer func() {
			if recover() != nil {
				reached = true
			}
		}()
		f.router.ServeHTTP(w, r)
	}()
	for _, recorded := range rec.All() {
		metadata, _ := recorded.Event.Metadata.(audit.DenyMetadata)
		if recorded.Event.Kind == audit.AuthAuthorizationDeny && metadata.Permission == "direct_sign_in" {
			denied = true
		}
	}
	return w.Code, reached, denied
}

// An in-process call is never refused as anonymous: it is either served as the user or denied outright.
// Every route an anonymous client gets past — login, webhooks, runners, metrics, static files — acts without
// a signed-in user, and there the call must be denied before any handler runs: a login, for one, would mint
// a session for the caller. The denial is told from a plan or permission refusal by its audit event.
func TestRouter_InternalCallerReachesOnlyAuthenticatedRoutes(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	caller := internalCallerFor(f.root)

	open := 0
	err := f.router.(*mux.Router).Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		template, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		path := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(template, "1")
		for _, method := range methods {
			name := method + " " + path
			code, reached, denied := f.sendRouter(method, path, &caller)
			assert.NotEqual(t, http.StatusUnauthorized, code, "%s refused an in-process call as anonymous", name)

			anonymousCode, anonymousReached, _ := f.sendRouter(method, path, nil)
			if anonymousCode == http.StatusUnauthorized && !anonymousReached {
				continue // authentication guards it, the call was served as the user
			}
			open++
			assert.False(t, reached, "%s ran its handler for an in-process call", name)
			assert.Equal(t, http.StatusForbidden, code, name)
			assert.True(t, denied, "%s: no direct_sign_in denial", name)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Greater(t, open, 10, "the walk found too few open routes: it or the anonymous probe is broken")

	// Routes that answer 401 by themselves, so the walk takes them for authenticated, and a page of the UI.
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/auth/login"},
		{http.MethodPost, "/api/auth/verify"},
		{http.MethodPost, "/api/auth/recovery"},
		{http.MethodGet, "/api/ws"},
		{http.MethodGet, "/api/metrics"},
		{http.MethodGet, "/api/ping"},
		{http.MethodPost, "/api/internal/runners"},
		{http.MethodGet, "/api/internal/runners"},
		{http.MethodDelete, "/api/internal/runners"},
		{http.MethodPost, "/api/integrations/some-alias"},
		{http.MethodGet, "/api/terraform/some-alias"},
		{http.MethodGet, "/.well-known/jwks.json"},
		{http.MethodGet, "/project/1/templates"}, // served by the not-found handler
	} {
		code, reached, denied := f.sendRouter(route.method, route.path, &caller)
		assert.False(t, reached, route.method+" "+route.path)
		assert.Equal(t, http.StatusForbidden, code, route.method+" "+route.path)
		assert.True(t, denied, route.method+" "+route.path)
	}
}

// The deny middleware on the open groups must not change them for direct requests.
func TestRouter_OpenRoutesForDirectRequests(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	code, _, _ := f.sendRouter(http.MethodGet, "/api/ping", nil)
	assert.Equal(t, http.StatusOK, code)
	code, _, _ = f.sendRouter(http.MethodGet, "/api/projects", nil)
	assert.Equal(t, http.StatusUnauthorized, code)
}
