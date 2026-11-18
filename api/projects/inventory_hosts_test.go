package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newManagedInventory(t *testing.T, f resourceFixture, name string) db.Inventory {
	t.Helper()
	inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: name, Type: db.InventoryManaged})
	require.NoError(t, err)
	return inventory
}

// hostCall runs a host handler and returns its response and what it recorded.
func hostCall(f resourceFixture, inventory db.Inventory, method, body string, hostID int, h http.HandlerFunc) (*httptest.ResponseRecorder, *audittest.Recorder) {
	vars := map[string]string{"inventory_id": strconv.Itoa(inventory.ID)}
	if hostID != 0 {
		vars["inventory_host_id"] = strconv.Itoa(hostID)
	}
	r, rec := f.request(method, body, map[string]any{"inventory": inventory}, vars)
	w := httptest.NewRecorder()
	h(w, r)
	return w, rec
}

func TestInventoryHosts_CreateListUpdateDelete(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")

	w, _ := hostCall(f, inventory, http.MethodPost, `{"name":"web1","enabled":true,"variables":"a: 1"}`, 0, AddInventoryHost)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventoryHost
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, inventory.ID, created.InventoryID)

	w, _ = hostCall(f, inventory, http.MethodGet, "", 0, GetInventoryHosts)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-Total-Count"))
	var hosts []db.InventoryHost
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &hosts))
	require.Len(t, hosts, 1)
	assert.Equal(t, "web1", hosts[0].Name)

	w, _ = hostCall(f, inventory, http.MethodPut, `{"name":"web2","enabled":false}`, created.ID, UpdateInventoryHost)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

	w, _ = hostCall(f, inventory, http.MethodGet, "", created.ID, GetInventoryHost)
	require.Equal(t, http.StatusOK, w.Code)
	var got db.InventoryHost
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "web2", got.Name)
	assert.True(t, got.Enabled)

	w, _ = hostCall(f, inventory, http.MethodPut, `{"enabled":false}`, created.ID, SetInventoryHostEnabled)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got, err := f.store.GetInventoryHost(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.False(t, got.Enabled)
	assert.Equal(t, "web2", got.Name)

	w, _ = hostCall(f, inventory, http.MethodDelete, "", created.ID, RemoveInventoryHost)
	require.Equal(t, http.StatusNoContent, w.Code)
	_, err = f.store.GetInventoryHost(f.project.ID, inventory.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestInventoryHosts_CreateDefaultsToEnabled(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		enabled bool
	}{
		{"enabled omitted", `{"name":"web1"}`, true},
		{"enabled false", `{"name":"web2","enabled":false}`, false},
	}
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := hostCall(f, inventory, http.MethodPost, tt.body, 0, AddInventoryHost)

			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			var created db.InventoryHost
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
			assert.Equal(t, tt.enabled, created.Enabled)
		})
	}
}

func TestInventoryHosts_RejectDuplicateAndInvalid(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	web1, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true})
	require.NoError(t, err)
	web2, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "web2", Enabled: true})
	require.NoError(t, err)

	tests := []struct {
		name   string
		method string
		body   string
		hostID int
		call   http.HandlerFunc
	}{
		{"duplicate", http.MethodPost, `{"name":"web1"}`, 0, AddInventoryHost},
		{"rename to taken", http.MethodPut, `{"name":"web1"}`, web2.ID, UpdateInventoryHost},
		{"range name", http.MethodPost, `{"name":"web[1:3]"}`, 0, AddInventoryHost},
		{"bad variables", http.MethodPost, `{"name":"web3","variables":"- a"}`, 0, AddInventoryHost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := hostCall(f, inventory, tt.method, tt.body, tt.hostID, tt.call)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}

	hosts, total, err := f.store.GetInventoryHosts(f.project.ID, inventory.ID, db.InventoryHostListParams{})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, hosts, 2)
	assert.Equal(t, []string{web1.Name, web2.Name}, []string{hosts[0].Name, hosts[1].Name})
}

func TestInventoryHosts_OtherInventoryIs404(t *testing.T) {
	f := newResourceFixture(t)
	a := newManagedInventory(t, f, "a")
	b := newManagedInventory(t, f, "b")
	host, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: a.ID, Name: "web1", Enabled: true})
	require.NoError(t, err)

	tests := []struct {
		name   string
		method string
		call   http.HandlerFunc
	}{
		{"get", http.MethodGet, GetInventoryHost},
		{"put", http.MethodPut, UpdateInventoryHost},
		{"delete", http.MethodDelete, RemoveInventoryHost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := hostCall(f, b, tt.method, `{"name":"moved"}`, host.ID, tt.call)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}

	got, err := f.store.GetInventoryHost(f.project.ID, a.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, "web1", got.Name)
}

func TestInventoryHosts_NotManagedIs400(t *testing.T) {
	f := newResourceFixture(t)
	inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "s", Type: db.InventoryStatic})
	require.NoError(t, err)

	tests := []struct {
		name   string
		method string
		hostID int
		call   http.HandlerFunc
	}{
		{"list", http.MethodGet, 0, GetInventoryHosts},
		{"create", http.MethodPost, 0, AddInventoryHost},
		{"get", http.MethodGet, 1, GetInventoryHost},
		{"update", http.MethodPut, 1, UpdateInventoryHost},
		{"delete", http.MethodDelete, 1, RemoveInventoryHost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := hostCall(f, inventory, tt.method, `{"name":"web1"}`, tt.hostID, tt.call)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "inventory is not managed")
		})
	}
}

func TestInventory_ManagedVariablesAreValidated(t *testing.T) {
	tests := []struct {
		name      string
		variables string
		code      int
	}{
		{"not a mapping", "- a", http.StatusBadRequest},
		{"mapping", "a: 1", http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			body, err := json.Marshal(map[string]any{"name": "m", "project_id": f.project.ID, "type": "managed", "variables": tt.variables, "inventory": "all"})
			require.NoError(t, err)
			r, _ := f.request(http.MethodPost, string(body), nil, nil)
			w := httptest.NewRecorder()

			AddInventory(w, r)

			assert.Equal(t, tt.code, w.Code, w.Body.String())
			inventories, err := f.store.GetInventories(f.project.ID, db.RetrieveQueryParams{}, nil)
			require.NoError(t, err)
			for _, inventory := range inventories {
				assert.Empty(t, inventory.Inventory)
			}
		})
	}
}

func TestInventory_ManagedTypeCanNotChange(t *testing.T) {
	tests := []struct {
		name     string
		from, to db.InventoryType
	}{
		{"managed to static", db.InventoryManaged, db.InventoryStatic},
		{"static to managed", db.InventoryStatic, db.InventoryManaged},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "i", Type: tt.from})
			require.NoError(t, err)
			body, err := json.Marshal(map[string]any{"id": inventory.ID, "name": "i", "project_id": f.project.ID, "type": tt.to})
			require.NoError(t, err)
			r, _ := f.request(http.MethodPut, string(body), map[string]any{"inventory": inventory}, nil)
			w := httptest.NewRecorder()

			UpdateInventory(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "the type of a managed inventory can not be changed")
			got, err := f.store.GetInventory(f.project.ID, inventory.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.from, got.Type)
		})
	}
}

func TestInventoryHosts_ListPageSize(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	for i := 0; i < 60; i++ {
		_, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "web" + strconv.Itoa(i), Enabled: true})
		require.NoError(t, err)
	}

	tests := []struct {
		name  string
		query string
		rows  int
	}{
		{"no count", "", 50},
		{"zero count", "?count=0", 50},
		{"small count", "?count=10", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := f.request(http.MethodGet, "", map[string]any{"inventory": inventory}, map[string]string{"inventory_id": strconv.Itoa(inventory.ID)})
			r.URL.RawQuery = strings.TrimPrefix(tt.query, "?")
			w := httptest.NewRecorder()

			GetInventoryHosts(w, r)

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "60", w.Header().Get("X-Total-Count"))
			var hosts []db.InventoryHost
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &hosts))
			assert.Len(t, hosts, tt.rows)
		})
	}
}

func TestDiffVariables(t *testing.T) {
	tests := []struct {
		name    string
		oldText string
		newText string
		changed []string
	}{
		{"same", "a: 1\nb: [x, y]", "b:\n  - x\n  - y\na: 1", nil},
		{"value", "a: 1", "a: 2", []string{"a"}},
		{"quoted", "flag: no", `flag: "no"`, []string{"flag"}},
		{"nested quoted", "a:\n  flag: yes", "a:\n  flag: 'yes'", []string{"a"}},
		{"tag", "a: !!int 1", "a: !!str 1", []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := diffVariables(tt.oldText, tt.newText)

			assert.Equal(t, tt.changed, diff.ChangedVariables)
			assert.Empty(t, diff.AddedVariables)
			assert.Empty(t, diff.RemovedVariables)
		})
	}
}

func TestSetInventoryVariables(t *testing.T) {
	tests := []struct {
		name string
		kind db.InventoryType
		body string
		code int
		want string
	}{
		{"mapping", db.InventoryManaged, `{"variables":"a: 2"}`, http.StatusNoContent, "a: 2"},
		{"not a mapping", db.InventoryManaged, `{"variables":"- a"}`, http.StatusBadRequest, "a: 1"},
		{"not managed", db.InventoryStatic, `{"variables":"a: 2"}`, http.StatusBadRequest, "a: 1"},
		{"cleared", db.InventoryManaged, `{"variables":""}`, http.StatusNoContent, ""},
		{"omitted", db.InventoryManaged, `{}`, http.StatusBadRequest, "a: 1"},
		{"null", db.InventoryManaged, `{"variables":null}`, http.StatusBadRequest, "a: 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "i", Type: tt.kind, Variables: new("a: 1")})
			require.NoError(t, err)

			w, _ := hostCall(f, inventory, http.MethodPut, tt.body, 0, SetInventoryVariables)

			assert.Equal(t, tt.code, w.Code, w.Body.String())
			got, err := f.store.GetInventory(f.project.ID, inventory.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.VariablesText())
		})
	}
}

func TestUpdateInventory_KeepsVariables(t *testing.T) {
	f := newResourceFixture(t)
	inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "i", Type: db.InventoryManaged, Variables: new("a: 1")})
	require.NoError(t, err)
	body, err := json.Marshal(map[string]any{"id": inventory.ID, "name": "renamed", "project_id": f.project.ID, "type": "managed", "variables": "a: 9"})
	require.NoError(t, err)
	r, _ := f.request(http.MethodPut, string(body), map[string]any{"inventory": inventory}, nil)
	w := httptest.NewRecorder()

	UpdateInventory(w, r)

	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got, err := f.store.GetInventory(f.project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)
	assert.Equal(t, "a: 1", got.VariablesText())
}

func TestSetInventoryHostEnabled_NeedsTheValue(t *testing.T) {
	for _, body := range []string{`{}`, `{"enabled":null}`} {
		t.Run(body, func(t *testing.T) {
			f := newResourceFixture(t)
			inventory := newManagedInventory(t, f, "managed")
			host, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "web1", Enabled: true})
			require.NoError(t, err)

			w, _ := hostCall(f, inventory, http.MethodPut, body, host.ID, SetInventoryHostEnabled)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			got, err := f.store.GetInventoryHost(f.project.ID, inventory.ID, host.ID)
			require.NoError(t, err)
			assert.True(t, got.Enabled)
		})
	}
}
