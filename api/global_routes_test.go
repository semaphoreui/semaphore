package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	taskServices "github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// globalRoute is one instance-wide route and the global permission that opens it. The body is what an allowed
// caller sends; it is chosen so that the handler answers with the nil services Route is built with here.
type globalRoute struct {
	method     string
	path       string
	body       string
	permission db.GlobalPermission
	// denyName is the permission name a denial is recorded with; empty means the catalog id of permission.
	denyName string
	// forbidsWithoutPlan marks a route whose handler answers 403 by itself when the plan lacks the feature
	// (the Pro controllers of roles and audit, and the community stubs), so for an allowed caller only the
	// audit trail tells a permission denial from that answer.
	forbidsWithoutPlan bool
}

// globalRoutes mirrors the "Route → permission" table of
// AGENTS/work/auth-and-identity/global-roles-and-permissions.md. A route added to systemAPI or rolesAdminAPI
// belongs here too.
func globalRoutes() []globalRoute {
	system := func(method, path, body string) globalRoute {
		return globalRoute{method: method, path: path, body: body, permission: db.CanManageGlobalSystem}
	}
	roles := func(method, path string) globalRoute {
		return globalRoute{method: method, path: path, permission: db.CanManageGlobalRoles, forbidsWithoutPlan: true}
	}
	audit := func(path string) globalRoute {
		return globalRoute{method: http.MethodGet, path: path, permission: db.CanReadGlobalAudit, denyName: "audit_log", forbidsWithoutPlan: true}
	}
	return []globalRoute{
		// manage system
		system(http.MethodGet, "/api/options", ""),
		system(http.MethodPost, "/api/options", `{"key":"test.option","value":"1"}`),
		system(http.MethodGet, "/api/admin/info", ""),
		system(http.MethodGet, "/api/cluster", ""),
		system(http.MethodGet, "/api/cluster/tasks", ""),
		system(http.MethodDelete, "/api/cluster/tasks", ""),
		system(http.MethodGet, "/api/runners", ""),
		system(http.MethodPost, "/api/runners", ""),
		system(http.MethodGet, "/api/runner_tags", ""),
		system(http.MethodDelete, "/api/cache", ""),
		system(http.MethodGet, "/api/runners/1", ""),
		system(http.MethodPut, "/api/runners/1", ""),
		system(http.MethodPost, "/api/runners/1", ""),
		system(http.MethodPost, "/api/runners/1/active", ""),
		system(http.MethodPost, "/api/runners/1/registration-token", ""),
		system(http.MethodDelete, "/api/runners/1", ""),
		system(http.MethodDelete, "/api/runners/1/cache", ""),
		system(http.MethodGet, "/api/role_mappings", ""),
		system(http.MethodPost, "/api/role_mappings", ""),
		system(http.MethodDelete, "/api/role_mappings/1", ""),
		system(http.MethodGet, "/api/apps/bash", ""),
		system(http.MethodPut, "/api/apps/bash", ""),
		system(http.MethodPost, "/api/apps/bash", ""),
		system(http.MethodPost, "/api/apps/bash/active", ""),
		system(http.MethodDelete, "/api/apps/bash", ""),
		system(http.MethodGet, "/api/tasks", ""),
		system(http.MethodGet, "/api/tasks/999", ""),
		system(http.MethodDelete, "/api/tasks/999", ""),

		// manage roles
		roles(http.MethodGet, "/api/roles"),
		roles(http.MethodPost, "/api/roles"),
		roles(http.MethodGet, "/api/roles/custom"),
		roles(http.MethodPut, "/api/roles/custom"),
		roles(http.MethodPost, "/api/roles/custom"),
		roles(http.MethodDelete, "/api/roles/custom"),
		{method: http.MethodGet, path: "/api/global-roles/permissions", permission: db.CanManageGlobalRoles}, // the catalog needs no plan
		roles(http.MethodGet, "/api/global-roles"),
		roles(http.MethodPost, "/api/global-roles"),
		roles(http.MethodGet, "/api/global-roles/1"),
		roles(http.MethodPut, "/api/global-roles/1"),
		roles(http.MethodPost, "/api/global-roles/1"),
		roles(http.MethodDelete, "/api/global-roles/1"),
		roles(http.MethodGet, "/api/users/1/global-roles"),
		roles(http.MethodPost, "/api/users/1/global-roles"),
		roles(http.MethodDelete, "/api/users/1/global-roles/1"),
		roles(http.MethodGet, "/api/users/1/global-permissions"),

		// manage users
		{method: http.MethodPost, path: "/api/users", permission: db.CanManageGlobalUsers},

		// read audit
		audit("/api/audit/events"),
		audit("/api/audit/events/export?format=csv"),
	}
}

// fakeRoleMappingStore keeps the Pro role-mappings controller from dereferencing a nil store: no mappings,
// nothing to delete.
type fakeRoleMappingStore struct{}

func (fakeRoleMappingStore) GetRoleMappings() ([]db.RoleMapping, error) {
	return []db.RoleMapping{}, nil
}
func (fakeRoleMappingStore) CreateRoleMapping(db.RoleMapping) (db.RoleMapping, error) {
	return db.RoleMapping{}, db.ErrNotFound
}
func (fakeRoleMappingStore) DeleteRoleMapping(int) (db.RoleMapping, []db.MappedRole, error) {
	return db.RoleMapping{}, nil, db.ErrNotFound
}
func (fakeRoleMappingStore) SyncMappedRoles(int, string, []string) ([]db.MappedRole, []db.MappedRole, error) {
	return nil, nil, nil
}

// globalRoutesFixture drives the real router: session cookie, CSRF, store, global role store and the permission
// middleware chain, exactly as a request from the browser goes.
type globalRoutesFixture struct {
	store   db.Store
	roles   fakeGlobalRoleStore
	router  http.Handler
	pool    *taskServices.TaskPool
	root    db.User
	plain   db.User
	holders map[db.GlobalPermission]db.User // exactly one bit
	others  map[db.GlobalPermission]db.User // every bit but that one
}

func newGlobalRoutesFixture(t *testing.T) globalRoutesFixture {
	t.Helper()
	store := setupSessionTest(t)
	util.Config.Debugging = &util.DebuggingConfig{}
	util.Config.TmpPath = t.TempDir()

	f := globalRoutesFixture{
		store:   store,
		roles:   fakeGlobalRoleStore{permissions: map[int]db.GlobalPermission{}},
		pool:    poolWithTasks(),
		holders: map[db.GlobalPermission]db.User{},
		others:  map[db.GlobalPermission]db.User{},
	}
	f.root = makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	f.plain = createUserOptionsTestUser(t, store, "plain")
	for _, definition := range db.GlobalPermissionCatalog() {
		holder := createUserOptionsTestUser(t, store, "holder-"+definition.ID)
		f.roles.permissions[holder.ID] = definition.Permission
		f.holders[definition.Permission] = holder

		other := createUserOptionsTestUser(t, store, "others-"+definition.ID)
		f.roles.permissions[other.ID] = db.AllGlobalPermissions &^ definition.Permission
		f.others[definition.Permission] = other
	}

	// noLicense stands in for the subscription service: the Pro controllers then refuse their routes by plan,
	// which the forbidsWithoutPlan flag of the table accounts for.
	f.router = Route(store, f.roles, nil, nil, fakeRoleMappingStore{}, nil, f.pool, nil, nil, nil, nil, nil, nil, nil, noLicense{}, nil, nil, nil, nil)
	return f
}

func (f globalRoutesFixture) send(t *testing.T, route globalRoute, actor db.User) (*httptest.ResponseRecorder, *audittest.Recorder) {
	t.Helper()
	r := httptest.NewRequest(route.method, route.path, bytesReader(route.body))
	r = addSessionCookie(t, f.store, r, actor, db.SessionVerificationNone)
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "task_pool", f.pool)
	r, rec := withAuditRecorder(r)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	return w, rec
}

func denials(rec *audittest.Recorder) []audittest.Recorded {
	var denied []audittest.Recorded
	for _, recorded := range rec.All() {
		if recorded.Event.Kind == audit.AuthAuthorizationDeny {
			denied = append(denied, recorded)
		}
	}
	return denied
}

func (route globalRoute) name() string { return route.method + " " + route.path }

func (route globalRoute) deniedAs() string {
	if route.denyName != "" {
		return route.denyName
	}
	return globalPermissionName(route.permission)
}

func (f globalRoutesFixture) assertAllowed(t *testing.T, route globalRoute, actor db.User) {
	t.Helper()
	w, rec := f.send(t, route, actor)
	assert.Empty(t, denials(rec), "%s: %s must not be denied", route.name(), actor.Username)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "%s: %s must be authenticated", route.name(), actor.Username)
	if !route.forbidsWithoutPlan {
		assert.NotEqual(t, http.StatusForbidden, w.Code, "%s: %s must pass", route.name(), actor.Username)
	}
}

func (f globalRoutesFixture) assertDenied(t *testing.T, route globalRoute, actor db.User) {
	t.Helper()
	w, rec := f.send(t, route, actor)
	assert.Equal(t, http.StatusForbidden, w.Code, "%s: %s must be refused", route.name(), actor.Username)
	denied := denials(rec)
	require.Len(t, denied, 1, "%s: %s must be refused with one audit event", route.name(), actor.Username)
	assert.Equal(t, audit.OutcomeFailure, denied[0].Event.Outcome)
	assert.Equal(t, audit.ReasonForbidden, denied[0].Event.Reason)
	assert.Equal(t, audit.DenyMetadata{Method: route.method, Permission: route.deniedAs()}, denied[0].Event.Metadata)
}

// Every instance-wide route is opened by its own global permission and by nothing else: not by the other bits
// one at a time, not by all the other bits together, and not by an account without roles. An admin passes
// everywhere. Driving the real router catches a handler that re-checks the admin flag behind the middleware.
func TestGlobalRoutes_PermissionTable(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	for _, route := range globalRoutes() {
		t.Run(route.name(), func(t *testing.T) {
			f.assertAllowed(t, route, f.root)
			f.assertAllowed(t, route, f.holders[route.permission])

			f.assertDenied(t, route, f.plain)
			f.assertDenied(t, route, f.others[route.permission])
			for permission, holder := range f.holders {
				if permission != route.permission {
					f.assertDenied(t, route, holder)
				}
			}
		})
	}
}

// globalRouteParams substitutes the path parameters of a route template with values the probe below can send.
var globalRouteParams = map[string]string{
	"{role_slug}": "custom",
	"{task_id}":   "999", // no such task in the pool, so DELETE stops nothing
	"{app_id}":    "bash",
	"{type}":      "ldap",
	"{provider}":  "default",
}

// globalRouteProbeSkips names route templates the probe skips: they are not authenticated API routes, or they need
// services Route is built without here.
func globalRouteProbeSkips(template string) bool {
	for _, prefix := range []string{"/api/ping", "/api/metrics", "/api/ws", "/api/auth/", "/api/internal/",
		"/api/integrations/", "/api/terraform/", "/.well-known/"} {
		if strings.HasPrefix(template, prefix) {
			return true
		}
	}
	return strings.Contains(template, "{project_id}")
}

// TestGlobalRoutes_TableCoversTheRouter walks the real router and sends every instance-level route as a user
// without roles. A route that refuses that user with the name of a global permission is behind the global
// permission middleware and must have a row in the table above; otherwise a route added to systemAPI or
// rolesAdminAPI would go untested.
func TestGlobalRoutes_TableCoversTheRouter(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	covered := map[string]bool{}
	for _, route := range globalRoutes() {
		covered[route.method+" "+strings.SplitN(route.path, "?", 2)[0]] = true
	}
	globalPermissionNames := map[string]bool{"audit_log": true}
	for _, definition := range db.GlobalPermissionCatalog() {
		globalPermissionNames[definition.ID] = true
	}

	gated := map[string]bool{}
	err := f.router.(*mux.Router).Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		template, err := route.GetPathTemplate()
		if err != nil || globalRouteProbeSkips(template) {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		path := template
		for param, value := range globalRouteParams {
			path = strings.ReplaceAll(path, param, value)
		}
		path = regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "1")
		for _, method := range methods {
			if method == http.MethodHead || method == http.MethodOptions {
				continue
			}
			rec := f.probe(t, method, path)
			for _, denied := range denials(rec) {
				metadata, _ := denied.Event.Metadata.(audit.DenyMetadata)
				if globalPermissionNames[metadata.Permission] {
					gated[method+" "+path] = true
				}
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, gated, "the probe found no gated route: the walk or the actor is broken")

	for name := range gated {
		assert.True(t, covered[name], "%s is behind a global permission but missing from globalRoutes()", name)
	}
}

// probe sends one request as the user without roles. The middleware answers before the handler on a gated
// route; on an open route a handler may hit one of the nil services and panic, which only says the route is
// not gated, so the panic is swallowed.
func (f globalRoutesFixture) probe(t *testing.T, method string, path string) *audittest.Recorder {
	t.Helper()
	r := httptest.NewRequest(method, path, nil)
	r = addSessionCookie(t, f.store, r, f.plain, db.SessionVerificationNone)
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "task_pool", f.pool)
	r, rec := withAuditRecorder(r)
	func() {
		defer func() { _ = recover() }()
		f.router.ServeHTTP(httptest.NewRecorder(), r)
	}()
	return rec
}
