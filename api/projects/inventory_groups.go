package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// inventoryGroupRequest is a group with the changes of its child groups
type inventoryGroupRequest struct {
	db.InventoryGroup
	AddChildren    []int `json:"add_children"`
	RemoveChildren []int `json:"remove_children"`
}

// GetInventoryGroups returns all groups of a managed inventory
func GetInventoryGroups(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	groups, err := helpers.Store(r).GetInventoryGroups(inventory.ProjectID, inventory.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The nil field hides the variables text, a list row carries only the count
	rows := make([]struct {
		db.InventoryGroup
		Variables *string `json:"variables,omitempty"`
	}, len(groups))
	for i, group := range groups {
		rows[i].InventoryGroup = group
	}

	helpers.WriteJSON(w, http.StatusOK, rows)
}

// AddInventoryGroup creates a group in a managed inventory
func AddInventoryGroup(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	var body inventoryGroupRequest
	if !helpers.Bind(w, r, &body) {
		return
	}

	group := body.InventoryGroup
	group.ID = 0
	group.InventoryID = inventory.ID
	group.ProjectID = inventory.ProjectID

	newGroup, err := helpers.Store(r).CreateInventoryGroup(group, body.AddChildren)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The stored group carries the counted fields, the request may hold anything in them
	newGroup, err = helpers.Store(r).GetInventoryGroup(inventory.ProjectID, inventory.ID, newGroup.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newGroup)
}

// GetInventoryGroup returns a group of a managed inventory
func GetInventoryGroup(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	groupID, ok := helpers.GetIntParamOrAbort("inventory_group_id", w, r)
	if !ok {
		return
	}

	group, err := helpers.Store(r).GetInventoryGroup(inventory.ProjectID, inventory.ID, groupID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, group)
}

// UpdateInventoryGroup writes updated values and child groups to a group of a managed inventory
func UpdateInventoryGroup(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	groupID, ok := helpers.GetIntParamOrAbort("inventory_group_id", w, r)
	if !ok {
		return
	}

	var body inventoryGroupRequest
	if !helpers.Bind(w, r, &body) {
		return
	}

	group := body.InventoryGroup
	group.ID = groupID
	group.InventoryID = inventory.ID
	group.ProjectID = inventory.ProjectID

	if _, _, err := helpers.Store(r).UpdateInventoryGroup(group, body.AddChildren, body.RemoveChildren); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventoryGroup deletes a group of a managed inventory and leaves its hosts and child groups alone
func RemoveInventoryGroup(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	groupID, ok := helpers.GetIntParamOrAbort("inventory_group_id", w, r)
	if !ok {
		return
	}

	if err := helpers.Store(r).DeleteInventoryGroup(inventory.ProjectID, inventory.ID, groupID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetInventoryHostGroups adds hosts of a managed inventory to groups and removes them from groups
func SetInventoryHostGroups(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	var body struct {
		Hosts  []int `json:"hosts"`
		Add    []int `json:"add"`
		Remove []int `json:"remove"`
	}
	if !helpers.Bind(w, r, &body) {
		return
	}

	if len(body.Hosts) == 0 {
		helpers.WriteErrorStatus(w, "hosts are required", http.StatusBadRequest)
		return
	}
	if len(body.Hosts) > 1000 {
		helpers.WriteErrorStatus(w, "at most 1000 hosts can be changed at once", http.StatusBadRequest)
		return
	}
	if len(body.Add) == 0 && len(body.Remove) == 0 {
		helpers.WriteErrorStatus(w, "add or remove is required", http.StatusBadRequest)
		return
	}

	_, _, err := helpers.Store(r).ChangeInventoryHostGroups(inventory.ProjectID, inventory.ID, body.Hosts, body.Add, body.Remove)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
