package projects

import (
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"gopkg.in/yaml.v3"
)

// writtenValue keeps the tag and the quoting of a scalar, since Ansible reads no and "no" differently
func writtenValue(node *yaml.Node) any {
	switch node.Kind {
	case yaml.MappingNode:
		values := map[string]any{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			values[node.Content[i].Value] = writtenValue(node.Content[i+1])
		}
		return values
	case yaml.SequenceNode:
		values := []any{}
		for _, item := range node.Content {
			values = append(values, writtenValue(item))
		}
		return values
	}
	quoted := node.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle|yaml.LiteralStyle|yaml.FoldedStyle) != 0
	return [3]any{node.Value, node.ShortTag(), quoted}
}

func writtenVariables(text string) map[string]any {
	var doc yaml.Node
	if _, err := db.ParseInventoryVariables(text); err != nil || yaml.Unmarshal([]byte(text), &doc) != nil || len(doc.Content) == 0 {
		return nil
	}
	values, _ := writtenValue(doc.Content[0]).(map[string]any)
	return values
}

// diffVariables names the top-level keys added, removed and changed, never the values.
func diffVariables(oldText, newText string) audit.InventoryVariablesMetadata {
	oldVars := writtenVariables(oldText)
	newVars := writtenVariables(newText)

	var diff audit.InventoryVariablesMetadata
	for _, key := range slices.Sorted(maps.Keys(newVars)) {
		oldValue, ok := oldVars[key]
		if !ok {
			diff.AddedVariables = append(diff.AddedVariables, key)
		} else if !reflect.DeepEqual(oldValue, newVars[key]) {
			diff.ChangedVariables = append(diff.ChangedVariables, key)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(oldVars)) {
		if _, ok := newVars[key]; !ok {
			diff.RemovedVariables = append(diff.RemovedVariables, key)
		}
	}
	return diff
}

// recordInventoryHost writes the event log entry and the audit record of a host change.
func recordInventoryHost(r *http.Request, action helpers.EventLogType, verb string, kind audit.Kind, inventory db.Inventory, host db.InventoryHost, metadata audit.InventoryHostMetadata) {
	helpers.EventLog(r, action, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   inventory.ProjectID,
		ObjectType:  db.EventInventoryHost,
		ObjectID:    host.ID,
		Description: fmt.Sprintf("Host %s %s in inventory %s", host.Name, verb, inventory.Name),
	})

	metadata.InventoryID = inventory.ID
	metadata.InventoryName = inventory.Name
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      kind,
		Target:    audit.ResourceTarget(audit.TargetInventoryHost, host.ID, host.Name),
		ProjectID: inventory.ProjectID,
		Metadata:  metadata,
	})
}

// managedInventory takes the inventory from the context and answers 400 when it is not managed.
func managedInventory(w http.ResponseWriter, r *http.Request) (db.Inventory, bool) {
	inventory := helpers.GetFromContext(r, "inventory").(db.Inventory)
	if inventory.Type != db.InventoryManaged {
		helpers.WriteErrorStatus(w, "inventory is not managed", http.StatusBadRequest)
		return inventory, false
	}
	return inventory, true
}

// GetInventoryHosts returns a page of the hosts of a managed inventory
func GetInventoryHosts(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	params := db.InventoryHostListParams{Search: r.URL.Query().Get("search")}

	var err error
	if s := r.URL.Query().Get("offset"); s != "" {
		if params.Offset, err = strconv.Atoi(s); err != nil || params.Offset < 0 {
			helpers.WriteErrorStatus(w, "invalid offset", http.StatusBadRequest)
			return
		}
	}
	if s := r.URL.Query().Get("count"); s != "" {
		if params.Count, err = strconv.Atoi(s); err != nil || params.Count < 0 {
			helpers.WriteErrorStatus(w, "invalid count", http.StatusBadRequest)
			return
		}
	}

	if params.Count == 0 {
		params.Count = 50
	}
	params.Count = min(params.Count, 1000)

	hosts, total, err := helpers.Store(r).GetInventoryHosts(inventory.ProjectID, inventory.ID, params)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	helpers.WriteJSON(w, http.StatusOK, hosts)
}

// AddInventoryHost creates a host in a managed inventory
func AddInventoryHost(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	// A host is enabled unless the request says otherwise
	host := db.InventoryHost{Enabled: true}
	if !helpers.Bind(w, r, &host) {
		return
	}

	host.ID = 0
	host.InventoryID = inventory.ID
	host.ProjectID = inventory.ProjectID

	newHost, err := helpers.Store(r).CreateInventoryHost(host)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventoryHost(r, helpers.EventLogCreate, "created", audit.ResourceInventoryHostCreate, inventory, newHost,
		audit.InventoryHostMetadata{AddedVariables: diffVariables("", newHost.Variables).AddedVariables})

	helpers.WriteJSON(w, http.StatusCreated, newHost)
}

// GetInventoryHost returns a host of a managed inventory
func GetInventoryHost(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	hostID, ok := helpers.GetIntParamOrAbort("inventory_host_id", w, r)
	if !ok {
		return
	}

	host, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, host)
}

// UpdateInventoryHost writes updated values to a host of a managed inventory
func UpdateInventoryHost(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	hostID, ok := helpers.GetIntParamOrAbort("inventory_host_id", w, r)
	if !ok {
		return
	}

	// The old host is the base of the audit record
	oldHost, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var host db.InventoryHost
	if !helpers.Bind(w, r, &host) {
		return
	}

	host.ID = hostID
	host.InventoryID = inventory.ID
	host.ProjectID = inventory.ProjectID

	if err := helpers.Store(r).UpdateInventoryHost(host); err != nil {
		helpers.WriteError(w, err)
		return
	}

	var metadata audit.InventoryHostMetadata
	if host.Name != oldHost.Name {
		metadata.ChangedFields = append(metadata.ChangedFields, "name")
		metadata.OldName = oldHost.Name
	}
	if host.Description != oldHost.Description {
		metadata.ChangedFields = append(metadata.ChangedFields, "description")
	}
	diff := diffVariables(oldHost.Variables, host.Variables)
	if len(diff.AddedVariables)+len(diff.RemovedVariables)+len(diff.ChangedVariables) > 0 {
		metadata.ChangedFields = append(metadata.ChangedFields, "variables")
	}
	metadata.AddedVariables = diff.AddedVariables
	metadata.RemovedVariables = diff.RemovedVariables
	metadata.ChangedVariables = diff.ChangedVariables

	if len(metadata.ChangedFields) > 0 {
		recordInventoryHost(r, helpers.EventLogUpdate, "updated", audit.ResourceInventoryHostUpdate, inventory, host, metadata)
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetInventoryHostEnabled enables or disables a host and leaves its other fields alone.
func SetInventoryHostEnabled(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	hostID, ok := helpers.GetIntParamOrAbort("inventory_host_id", w, r)
	if !ok {
		return
	}

	host, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if !helpers.Bind(w, r, &body) {
		return
	}

	// An empty body would otherwise disable the host
	if body.Enabled == nil {
		helpers.WriteErrorStatus(w, "enabled is required", http.StatusBadRequest)
		return
	}

	changed, err := helpers.Store(r).SetInventoryHostEnabled(inventory.ProjectID, inventory.ID, hostID, *body.Enabled)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The update itself tells whether the flag changed, so two requests never record one change twice
	if changed {
		verb, kind := "disabled", audit.ResourceInventoryHostDisable
		if *body.Enabled {
			verb, kind = "enabled", audit.ResourceInventoryHostEnable
		}
		recordInventoryHost(r, helpers.EventLogUpdate, verb, kind, inventory, host, audit.InventoryHostMetadata{})
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventoryHost deletes a host of a managed inventory
func RemoveInventoryHost(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	hostID, ok := helpers.GetIntParamOrAbort("inventory_host_id", w, r)
	if !ok {
		return
	}

	host, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).DeleteInventoryHost(inventory.ProjectID, inventory.ID, hostID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventoryHost(r, helpers.EventLogDelete, "deleted", audit.ResourceInventoryHostDelete, inventory, host, audit.InventoryHostMetadata{})

	w.WriteHeader(http.StatusNoContent)
}
