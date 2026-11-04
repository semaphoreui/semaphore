package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpSettingsPath = "/api/mcp/settings"

func mcpSettingsRequest(method, body string) globalRoute {
	return globalRoute{method: method, path: mcpSettingsPath, body: body, permission: db.CanManageGlobalSystem}
}

// mcpSettingsBody is a valid PUT body with every setting off, one field replaced by override
// ("" keeps them all off).
func mcpSettingsBody(override string) string {
	fields := []string{`"allow_run":false`, `"allow_destroy":false`, `"admin_allow_write":false`, `"admin_allow_destroy":false`}
	if override != "" {
		name, _, _ := strings.Cut(override, ":")
		for i, field := range fields {
			if strings.HasPrefix(field, name+":") {
				fields[i] = override
			}
		}
	}
	return "{" + strings.Join(fields, ",") + "}"
}

func decodeMCPSettings(t *testing.T, body []byte) db.MCPSettings {
	t.Helper()
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &fields), string(body))
	assert.Len(t, fields, 4, "the answer holds the four settings only")
	var settings db.MCPSettings
	require.NoError(t, json.Unmarshal(body, &settings))
	return settings
}

func TestMCPSettings_DefaultsWithoutStoredRows(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	w, _ := f.send(t, mcpSettingsRequest(http.MethodGet, ""), f.plain)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, db.MCPSettings{}, decodeMCPSettings(t, w.Body.Bytes()), "every group is off by default")
}

func TestMCPSettings_EverySignedInUserReads(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	require.NoError(t, db.WriteMCPSettings(f.store, db.MCPSettings{AllowRun: false, AllowDestroy: true, AdminAllowWrite: true}))

	for _, actor := range []db.User{f.plain, f.root, f.holders[db.CanManageGlobalSystem]} {
		w, rec := f.send(t, mcpSettingsRequest(http.MethodGet, ""), actor)
		require.Equal(t, http.StatusOK, w.Code, "%s: %s", actor.Username, w.Body.String())
		assert.Equal(t, db.MCPSettings{AllowRun: false, AllowDestroy: true, AdminAllowWrite: true}, decodeMCPSettings(t, w.Body.Bytes()), actor.Username)
		assert.Empty(t, rec.All(), "%s: reading is not audited", actor.Username)
	}
}

func TestMCPSettings_SystemManagersWriteAndAreAudited(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	for _, actor := range []db.User{f.root, f.holders[db.CanManageGlobalSystem]} {
		for _, want := range []db.MCPSettings{
			{AllowRun: true, AllowDestroy: false},
			{AllowRun: true, AllowDestroy: true},
			{AdminAllowWrite: true},
			{AdminAllowDestroy: true, AllowDestroy: true},
			{AllowRun: true, AllowDestroy: true, AdminAllowWrite: true, AdminAllowDestroy: true},
			{},
		} {
			body, err := json.Marshal(want)
			require.NoError(t, err)

			w, rec := f.send(t, mcpSettingsRequest(http.MethodPut, string(body)), actor)

			require.Equal(t, http.StatusOK, w.Code, "%s: %s", actor.Username, w.Body.String())
			assert.Equal(t, want, decodeMCPSettings(t, w.Body.Bytes()), "%s: the answer is the stored values", actor.Username)
			stored, err := db.ReadMCPSettings(f.store)
			require.NoError(t, err)
			assert.Equal(t, want, stored, actor.Username)

			require.Len(t, rec.All(), 1, "%s: one audit event per change", actor.Username)
			updated := onlyEvent(t, rec, audit.SystemSettingsUpdate)
			assert.Equal(t, audit.OutcomeSuccess, updated.Event.Outcome)
			assert.Equal(t, audit.SettingsMetadata{Keys: []string{db.MCPAllowRunOption, db.MCPAllowDestroyOption,
				db.MCPAdminAllowWriteOption, db.MCPAdminAllowDestroyOption}}, updated.Event.Metadata)
			assert.Equal(t, actor.Username, updated.Actor.Name, "the change is recorded as the caller's")
		}
	}
}

func TestMCPSettings_OthersMayNotWrite(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	refused := []db.User{f.plain, f.others[db.CanManageGlobalSystem], f.holders[db.CanManageGlobalUsers]}

	for _, actor := range refused {
		w, rec := f.send(t, mcpSettingsRequest(http.MethodPut, mcpSettingsBody(`"allow_destroy":true`)), actor)

		assert.Equal(t, http.StatusForbidden, w.Code, actor.Username)
		denied := denials(rec)
		require.Len(t, denied, 1, actor.Username)
		assert.Equal(t, audit.DenyMetadata{Method: http.MethodPut, Permission: globalPermissionName(db.CanManageGlobalSystem)}, denied[0].Event.Metadata)
		_, err := rec.Only(audit.SystemSettingsUpdate)
		assert.Error(t, err, "%s: no change is recorded", actor.Username)
	}

	stored, err := db.ReadMCPSettings(f.store)
	require.NoError(t, err)
	assert.Equal(t, db.DefaultMCPSettings(), stored, "a refused request changes nothing")
}

func TestMCPSettings_InvalidBodies(t *testing.T) {
	f := newGlobalRoutesFixture(t)
	bodies := map[string]string{
		"empty":                       "",
		"not JSON":                    "allow_run=true",
		"null":                        "null",
		"an array":                    `[true,false,true,false]`,
		"allow_run missing":           `{"allow_destroy":false,"admin_allow_write":false,"admin_allow_destroy":false}`,
		"allow_destroy missing":       `{"allow_run":true,"admin_allow_write":false,"admin_allow_destroy":false}`,
		"admin_allow_write missing":   `{"allow_run":true,"allow_destroy":false,"admin_allow_destroy":false}`,
		"admin_allow_destroy missing": `{"allow_run":true,"allow_destroy":false,"admin_allow_write":false}`,
		"only the project pair":       `{"allow_run":true,"allow_destroy":false}`,
		"a string":                    mcpSettingsBody(`"allow_run":"true"`),
		"a number":                    mcpSettingsBody(`"admin_allow_destroy":0`),
		"null field":                  mcpSettingsBody(`"admin_allow_write":null`),
		"an unknown field":            `{"allow_run":true,"allow_destroy":false,"admin_allow_write":false,"admin_allow_destroy":false,"enabled":true}`,
		"an unknown field for one":    `{"allow_run":true,"allow_destroy":false,"admin_allow_write":false,"enabled":true}`,
		"two objects":                 mcpSettingsBody("") + mcpSettingsBody(""),
		"trailing garbage":            mcpSettingsBody("") + ` x`,
		"larger than the limit":       `{"allow_run":true,"allow_destroy":false,"admin_allow_write":false,"admin_allow_destroy":false,"pad":"` + string(make([]byte, mcpSettingsMaxBodyBytes)) + `"}`,
		"case of the field name":      `{"Allow_Run":true,"allow_destroy":false,"admin_allow_write":false,"admin_allow_destroy":false}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			w, rec := f.send(t, mcpSettingsRequest(http.MethodPut, body), f.root)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.Empty(t, rec.All(), "an invalid body is not audited")
			stored, err := db.ReadMCPSettings(f.store)
			require.NoError(t, err)
			assert.Equal(t, db.DefaultMCPSettings(), stored, "an invalid body changes nothing")
		})
	}
}

// The two MCP endpoints are mounted with exact paths, so the settings route next
// to them is served by its own handler, and nothing else under /api/mcp exists.
func TestMCPSettings_RoutedApartFromTheMCPEndpoints(t *testing.T) {
	f := newGlobalRoutesFixture(t)

	w, _ := f.send(t, mcpSettingsRequest(http.MethodGet, ""), f.plain)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))
	decodeMCPSettings(t, w.Body.Bytes())

	// mcp.enabled is off in the fixture: the MCP handlers answer 501 in both
	// builds, which no /api route of the core does.
	for _, path := range []string{"/api/mcp/project", "/api/mcp/admin"} {
		for _, method := range []string{http.MethodPost, http.MethodGet} {
			w, _ = f.send(t, globalRoute{method: method, path: path, body: `{}`}, f.plain)
			assert.Equal(t, http.StatusNotImplemented, w.Code, "%s %s reaches the MCP handler", method, path)
			assert.Equal(t, "no-store", w.Header().Get("Cache-Control"), "%s %s reaches the MCP handler", method, path)
		}
	}

	// The bare path of the first MCP release and any path below the endpoints
	// or the settings is no route at all.
	for _, path := range []string{"/api/mcp", "/api/mcp/", "/api/mcp/settings/x", "/api/mcp/project/x", "/api/mcp/admin/x", "/api/mcp/other"} {
		w, _ = f.send(t, globalRoute{method: http.MethodPost, path: path, body: `{}`}, f.plain)
		assert.Equal(t, http.StatusNotFound, w.Code, path)
	}
}
