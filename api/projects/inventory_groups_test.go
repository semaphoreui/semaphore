package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// groupCall runs a group handler and returns its response.
func groupCall(f resourceFixture, inventory db.Inventory, method, body string, groupID int, h http.HandlerFunc) *httptest.ResponseRecorder {
	vars := map[string]string{"inventory_id": strconv.Itoa(inventory.ID)}
	if groupID != 0 {
		vars["inventory_group_id"] = strconv.Itoa(groupID)
	}
	r, _ := f.request(method, body, map[string]any{"inventory": inventory}, vars)
	w := httptest.NewRecorder()
	h(w, r)
	return w
}

func newInventoryGroup(t *testing.T, f resourceFixture, inventory db.Inventory, name string, children ...int) db.InventoryGroup {
	t.Helper()
	group, err := f.store.CreateInventoryGroup(db.InventoryGroup{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: name}, children)
	require.NoError(t, err)
	return group
}

func newInventoryHost(t *testing.T, f resourceFixture, inventory db.Inventory, name string, groups ...int) db.InventoryHost {
	t.Helper()
	host, err := f.store.CreateInventoryHost(db.InventoryHost{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: name, Enabled: true}, groups)
	require.NoError(t, err)
	return host
}

func TestInventoryGroups_CreateListGet(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")

	w := groupCall(f, inventory, http.MethodGet, "", 0, GetInventoryGroups)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String())

	child := newInventoryGroup(t, f, inventory, "eu")
	w = groupCall(f, inventory, http.MethodPost, `{"name":"web","variables":"a: 1","add_children":[`+strconv.Itoa(child.ID)+`]}`, 0, AddInventoryGroup)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventoryGroup
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, "web", created.Name)
	assert.Equal(t, []int{child.ID}, created.Children)

	w = groupCall(f, inventory, http.MethodGet, "", 0, GetInventoryGroups)
	require.Equal(t, http.StatusOK, w.Code)
	var groups []db.InventoryGroup
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &groups))
	require.Len(t, groups, 2)
	assert.Equal(t, []string{"eu", "web"}, []string{groups[0].Name, groups[1].Name})

	w = groupCall(f, inventory, http.MethodGet, "", created.ID, GetInventoryGroup)
	require.Equal(t, http.StatusOK, w.Code)
	var got db.InventoryGroup
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "a: 1", got.Variables)
	assert.Equal(t, []int{child.ID}, got.Children)

	w = groupCall(f, inventory, http.MethodPut, `{"name":"app","remove_children":[`+strconv.Itoa(child.ID)+`]}`, created.ID, UpdateInventoryGroup)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got, err := f.store.GetInventoryGroup(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "app", got.Name)
	assert.Empty(t, got.Children)

	w = groupCall(f, inventory, http.MethodDelete, "", created.ID, RemoveInventoryGroup)
	require.Equal(t, http.StatusNoContent, w.Code)
	_, err = f.store.GetInventoryGroup(f.project.ID, inventory.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestInventoryGroups_CreateAnswersStoredValues(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	child := newInventoryGroup(t, f, inventory, "eu")
	other := newInventoryGroup(t, f, inventory, "us")
	newInventoryHost(t, f, inventory, "web1", child.ID)
	body := `{"id":999,"name":"web","host_count":999,"children":[` + strconv.Itoa(other.ID) + `],"add_children":[` + strconv.Itoa(child.ID) + `]}`

	w := groupCall(f, inventory, http.MethodPost, body, 0, AddInventoryGroup)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventoryGroup
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	stored, err := f.store.GetInventoryGroup(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, stored, created)
	assert.Equal(t, []int{child.ID}, created.Children)
	assert.Equal(t, 1, created.HostCount)
}

func TestGetInventoryGroups_LeavesVariablesOut(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	group, err := f.store.CreateInventoryGroup(db.InventoryGroup{ProjectID: f.project.ID, InventoryID: inventory.ID, Name: "web", Variables: "a: 1"}, nil)
	require.NoError(t, err)

	w := groupCall(f, inventory, http.MethodGet, "", 0, GetInventoryGroups)
	require.Equal(t, http.StatusOK, w.Code)
	var groups []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &groups))
	require.Len(t, groups, 1)
	assert.Equal(t, "web", groups[0]["name"])
	assert.EqualValues(t, 1, groups[0]["variable_count"])
	assert.NotContains(t, groups[0], "variables")
	assert.Contains(t, groups[0], "children")
	assert.Contains(t, groups[0], "host_count")

	w = groupCall(f, inventory, http.MethodGet, "", group.ID, GetInventoryGroup)
	require.Equal(t, http.StatusOK, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "a: 1", got["variables"])
	assert.EqualValues(t, 1, got["variable_count"])
}

func TestInventoryGroups_CycleIsRejected(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	child := newInventoryGroup(t, f, inventory, "child")
	parent := newInventoryGroup(t, f, inventory, "parent", child.ID)

	w := groupCall(f, inventory, http.MethodPut, `{"name":"renamed","add_children":[`+strconv.Itoa(parent.ID)+`]}`, child.ID, UpdateInventoryGroup)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	got, err := f.store.GetInventoryGroup(f.project.ID, inventory.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, "child", got.Name)
	assert.Empty(t, got.Children)
}

func TestInventoryGroups_HostOfOtherInventoryIsRejected(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	other := newManagedInventory(t, f, "other")
	group := newInventoryGroup(t, f, inventory, "web")
	host := newInventoryHost(t, f, inventory, "web1")
	foreign := newInventoryHost(t, f, other, "web2")
	body, err := json.Marshal(map[string]any{"hosts": []int{host.ID, foreign.ID}, "add": []int{group.ID}})
	require.NoError(t, err)

	w := groupCall(f, inventory, http.MethodPut, string(body), 0, SetInventoryHostGroups)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	links, err := f.store.GetInventoryGroupHosts(f.project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestInventoryGroups_SetHostGroups(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	web := newInventoryGroup(t, f, inventory, "web")
	db1 := newInventoryGroup(t, f, inventory, "db")
	host := newInventoryHost(t, f, inventory, "web1", db1.ID)
	body, err := json.Marshal(map[string]any{"hosts": []int{host.ID}, "add": []int{web.ID}, "remove": []int{db1.ID}})
	require.NoError(t, err)

	w := groupCall(f, inventory, http.MethodPut, string(body), 0, SetInventoryHostGroups)

	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got, err := f.store.GetInventoryHost(f.project.ID, inventory.ID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{web.ID}, got.Groups)
}

func TestInventoryGroups_HostCarriesGroups(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	web := newInventoryGroup(t, f, inventory, "web")
	db1 := newInventoryGroup(t, f, inventory, "db")

	w, _ := hostCall(f, inventory, http.MethodPost, `{"name":"web1","add_groups":[`+strconv.Itoa(web.ID)+`]}`, 0, AddInventoryHost)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.InventoryHost
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, []int{web.ID}, created.Groups)

	body := `{"name":"web1","add_groups":[` + strconv.Itoa(db1.ID) + `],"remove_groups":[` + strconv.Itoa(web.ID) + `]}`
	w, _ = hostCall(f, inventory, http.MethodPut, body, created.ID, UpdateInventoryHost)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	got, err := f.store.GetInventoryHost(f.project.ID, inventory.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{db1.ID}, got.Groups)
}

func TestInventoryGroups_AddedAndRemovedIsRejected(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	group := newInventoryGroup(t, f, inventory, "web")
	other := newInventoryGroup(t, f, inventory, "db")
	host := newInventoryHost(t, f, inventory, "web1")
	id := strconv.Itoa(other.ID)

	tests := []struct {
		name string
		body string
		id   int
		call http.HandlerFunc
	}{
		{"host groups", `{"hosts":[` + strconv.Itoa(host.ID) + `],"add":[` + id + `],"remove":[` + id + `]}`, 0, SetInventoryHostGroups},
		{"group", `{"name":"web","add_children":[` + id + `],"remove_children":[` + id + `]}`, group.ID, UpdateInventoryGroup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := groupCall(f, inventory, http.MethodPut, tt.body, tt.id, tt.call)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}

	w, _ := hostCall(f, inventory, http.MethodPut, `{"name":"web1","add_groups":[`+id+`],"remove_groups":[`+id+`]}`, host.ID, UpdateInventoryHost)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

	links, err := f.store.GetInventoryGroupHosts(f.project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestInventoryGroups_HostGroupsNeedsHostsAndChanges(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	group := newInventoryGroup(t, f, inventory, "web")
	host := newInventoryHost(t, f, inventory, "web1")
	tooMany := make([]int, 1001)
	for i := range tooMany {
		tooMany[i] = host.ID
	}

	tests := []struct {
		name string
		body map[string]any
	}{
		{"no hosts", map[string]any{"hosts": []int{}, "add": []int{group.ID}}},
		{"no changes", map[string]any{"hosts": []int{host.ID}, "add": nil}},
		{"too many hosts", map[string]any{"hosts": tooMany, "add": []int{group.ID}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := groupCall(f, inventory, http.MethodPut, string(body), 0, SetInventoryHostGroups)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}

	links, err := f.store.GetInventoryGroupHosts(f.project.ID, inventory.ID)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestInventoryGroups_HostsOfGroup(t *testing.T) {
	f := newResourceFixture(t)
	inventory := newManagedInventory(t, f, "managed")
	child := newInventoryGroup(t, f, inventory, "eu")
	parent := newInventoryGroup(t, f, inventory, "web", child.ID)
	inChild := newInventoryHost(t, f, inventory, "web1", child.ID)
	newInventoryHost(t, f, inventory, "web2")

	tests := []struct {
		name  string
		query string
		code  int
		host  string
	}{
		{"group with descendants", "group=" + strconv.Itoa(parent.ID), http.StatusOK, inChild.Name},
		{"unknown group", "group=" + strconv.Itoa(parent.ID+100), http.StatusNotFound, ""},
		{"group and ungrouped", "group=" + strconv.Itoa(parent.ID) + "&ungrouped=1", http.StatusBadRequest, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := f.request(http.MethodGet, "", map[string]any{"inventory": inventory}, map[string]string{"inventory_id": strconv.Itoa(inventory.ID)})
			r.URL.RawQuery = tt.query
			w := httptest.NewRecorder()

			GetInventoryHosts(w, r)

			require.Equal(t, tt.code, w.Code, w.Body.String())
			if tt.host == "" {
				return
			}
			var hosts []db.InventoryHost
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &hosts))
			require.Len(t, hosts, 1)
			assert.Equal(t, tt.host, hosts[0].Name)
			assert.Equal(t, []int{child.ID}, hosts[0].Groups)
		})
	}
}

func TestInventoryGroups_NotManagedIs400(t *testing.T) {
	f := newResourceFixture(t)
	inventory, err := f.store.CreateInventory(db.Inventory{ProjectID: f.project.ID, Name: "s", Type: db.InventoryStatic})
	require.NoError(t, err)

	tests := []struct {
		name    string
		method  string
		groupID int
		call    http.HandlerFunc
	}{
		{"list", http.MethodGet, 0, GetInventoryGroups},
		{"create", http.MethodPost, 0, AddInventoryGroup},
		{"get", http.MethodGet, 1, GetInventoryGroup},
		{"update", http.MethodPut, 1, UpdateInventoryGroup},
		{"delete", http.MethodDelete, 1, RemoveInventoryGroup},
		{"host groups", http.MethodPut, 0, SetInventoryHostGroups},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := groupCall(f, inventory, tt.method, `{"name":"web","hosts":[1],"add":[1]}`, tt.groupID, tt.call)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "inventory is not managed")
		})
	}
}

func TestInventoryGroups_OtherInventoryIs404(t *testing.T) {
	f := newResourceFixture(t)
	a := newManagedInventory(t, f, "a")
	b := newManagedInventory(t, f, "b")
	group := newInventoryGroup(t, f, a, "web")

	tests := []struct {
		name   string
		method string
		call   http.HandlerFunc
	}{
		{"get", http.MethodGet, GetInventoryGroup},
		{"put", http.MethodPut, UpdateInventoryGroup},
		{"delete", http.MethodDelete, RemoveInventoryGroup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := groupCall(f, b, tt.method, `{"name":"moved"}`, group.ID, tt.call)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}

	got, err := f.store.GetInventoryGroup(f.project.ID, a.ID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "web", got.Name)
}
