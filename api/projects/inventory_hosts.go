package projects

import (
	"net/http"
	"strconv"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

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

	if err := host.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	newHost, err := helpers.Store(r).CreateInventoryHost(host)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

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

	// The store ignores a missing host, so it is loaded first.
	if _, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID); err != nil {
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

	if err := host.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateInventoryHost(host); err != nil {
		helpers.WriteError(w, err)
		return
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

	if _, err := helpers.Store(r).GetInventoryHost(inventory.ProjectID, inventory.ID, hostID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).DeleteInventoryHost(inventory.ProjectID, inventory.ID, hostID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
