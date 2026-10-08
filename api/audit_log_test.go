package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	proApi "github.com/semaphoreui/semaphore/pro/api"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogMiddleware_DeniesNonAdminAndRecordsIt(t *testing.T) {
	r, rec := withAuditRecorder(asActor(httptest.NewRequest(http.MethodGet, "/api/audit/events", nil),
		db.User{ID: 5, Username: "mallory"}))
	w := httptest.NewRecorder()
	called := false

	auditLogMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, called)
	got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodGet, Permission: "audit_log"}, got.Event.Metadata)
}

func TestAuditLogMiddleware_LetsAdminThrough(t *testing.T) {
	r, rec := withAuditRecorder(asActor(httptest.NewRequest(http.MethodGet, "/api/audit/events", nil),
		db.User{ID: 1, Admin: true}))
	called := false

	auditLogMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).
		ServeHTTP(httptest.NewRecorder(), r)

	assert.True(t, called)
	assert.Empty(t, rec.All())
}

func TestAuditLogMiddleware_RefusesCrossSiteRequests(t *testing.T) {
	for site, allowed := range map[string]bool{"cross-site": false, "same-origin": true, "none": true, "": true} {
		r, rec := withAuditRecorder(asActor(httptest.NewRequest(http.MethodGet, "/api/audit/events/export", nil),
			db.User{ID: 1, Admin: true}))
		if site != "" {
			r.Header.Set("Sec-Fetch-Site", site)
		}
		w := httptest.NewRecorder()
		called := false

		auditLogMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)

		assert.Equal(t, allowed, called, site)
		if !allowed {
			assert.Equal(t, http.StatusForbidden, w.Code, site)
		}
		assert.Empty(t, rec.All(), site)
	}
}

func auditLogRequest(t *testing.T, target string, events int) *http.Request {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	for i := 0; i < events; i++ {
		_, err := store.CreateAuditEvent(context.Background(), db.AuditEvent{
			EventID: "e", SchemaVersion: "1", Category: "auth", EventCode: "auth.login", Type: "start",
			Action: "authenticate", Outcome: "success", ActorType: "user", ActorID: "1", ActorName: "admin",
			InstanceID: "test", Metadata: "{}",
		})
		require.NoError(t, err)
	}
	return helpers.SetContextValue(httptest.NewRequest(http.MethodGet, target, nil), "store", store)
}

func notFiltered(t *testing.T) http.HandlerFunc {
	return func(http.ResponseWriter, *http.Request) { t.Fatal("the request went to the filtered handler") }
}

func TestGetAuditEvents_PagesTheWholeLog(t *testing.T) {
	w := httptest.NewRecorder()

	getAuditEvents(notFiltered(t))(w, auditLogRequest(t, "/api/audit/events?before=3", 3))

	require.Equal(t, http.StatusOK, w.Code)
	var page auditPage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
	require.Len(t, page.Events, 2)
	assert.Equal(t, int64(2), page.Events[0].Seq)
	assert.Equal(t, "auth.login", page.Events[0].EventCode)
	assert.Nil(t, page.Older)
	require.NotNil(t, page.Newer)
	assert.Equal(t, int64(2), *page.Newer)
}

func TestGetAuditEvents_SendsAnyOtherParameterToTheFilteredHandler(t *testing.T) {
	for _, target := range []string{"/api/audit/events?user=1", "/api/audit/events?before=3&ip=10.0.0.1",
		"/api/audit/events?anything=x"} {
		called := false
		getAuditEvents(func(http.ResponseWriter, *http.Request) { called = true })(httptest.NewRecorder(),
			auditLogRequest(t, target, 0))
		assert.True(t, called, target)
	}
}

func TestGetAuditEvents_RejectsBadCursor(t *testing.T) {
	for _, target := range []string{"/api/audit/events?before=x", "/api/audit/events?after=0",
		"/api/audit/events?before=5&after=2", "/api/audit/events?after=9223372036854775807"} {
		w := httptest.NewRecorder()
		getAuditEvents(notFiltered(t))(w, auditLogRequest(t, target, 0))
		assert.Equal(t, http.StatusBadRequest, w.Code, target)
	}
}

func TestAuditStub_RefusesFiltersAndExport(t *testing.T) {
	controller := proApi.NewAuditController(nil, nil)
	w := httptest.NewRecorder()
	controller.GetEvents(w, httptest.NewRequest(http.MethodGet, "/api/audit/events?user=1", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	w = httptest.NewRecorder()
	controller.ExportEvents(w, httptest.NewRequest(http.MethodGet, "/api/audit/events/export?format=csv", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuditRoutes_AreGetOnly(t *testing.T) {
	prev := util.Config
	t.Cleanup(func() { util.Config = prev })
	util.Config = &util.ConfigType{Debugging: &util.DebuggingConfig{}}
	router := Route(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	for _, target := range []string{"/api/audit/events", "/api/audit/events/export?format=csv"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodHead, target, nil))
		assert.Equal(t, http.StatusNotFound, w.Code, target)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		assert.Equal(t, http.StatusUnauthorized, w.Code, target)
	}
}
