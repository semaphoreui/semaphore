package projects

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

// recordInventoryGroup writes one event log entry and an audit record for each metadata of a group change.
func recordInventoryGroup(r *http.Request, action helpers.EventLogType, verb string, kind audit.Kind, inventory db.Inventory, group db.InventoryGroup, metadata ...audit.InventoryGroupMetadata) {
	helpers.EventLog(r, action, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   inventory.ProjectID,
		ObjectType:  db.EventInventoryGroup,
		ObjectID:    group.ID,
		Description: fmt.Sprintf("Group %s %s in inventory %s", group.Name, verb, inventory.Name),
	})

	for _, part := range metadata {
		part.InventoryID = inventory.ID
		part.InventoryName = inventory.Name
		helpers.Audit(r).Record(r.Context(), audit.Event{
			Kind:      kind,
			Target:    audit.ResourceTarget(audit.TargetInventoryGroup, group.ID, group.Name),
			ProjectID: inventory.ProjectID,
			Metadata:  part,
		})
	}
}

// auditNameChunks cuts names into lists that each fit one audit record of the inventory.
// The audit drops metadata over the limit whole, so a long list goes to several records.
func auditNameChunks(inventory db.Inventory, names []string) [][]string {
	base, _ := json.Marshal(audit.InventoryGroupMetadata{InventoryID: inventory.ID, InventoryName: inventory.Name})
	chunks := [][]string{nil}
	size := len(base)
	for _, name := range names {
		raw, _ := json.Marshal(name)
		if len(chunks[len(chunks)-1]) > 0 && size+len(raw)+1 > audit.MaxMetadataBytes*3/4 {
			chunks = append(chunks, nil)
			size = len(base)
		}
		chunks[len(chunks)-1] = append(chunks[len(chunks)-1], name)
		size += len(raw) + 1
	}
	return chunks
}

// recordInventoryGroupLinks records each group whose links changed, names holds the names of the linked hosts or child groups.
// A name that could not be read is written as the id, so the record keeps what changed.
func recordInventoryGroupLinks(r *http.Request, inventory db.Inventory, kind audit.Kind, verb string, groupNames map[int]string, links map[int][]int, names map[int]string) {
	hosts := kind == audit.ResourceInventoryGroupHostAdd || kind == audit.ResourceInventoryGroupHostRemove

	for _, groupID := range slices.Sorted(maps.Keys(links)) {
		if len(links[groupID]) == 0 {
			continue
		}

		linked := make([]string, len(links[groupID]))
		for i, id := range links[groupID] {
			name, ok := names[id]
			if !ok {
				name = fmt.Sprintf("#%d", id)
			}
			linked[i] = name
		}

		chunks := auditNameChunks(inventory, linked)
		parts := make([]audit.InventoryGroupMetadata, len(chunks))
		for i, chunk := range chunks {
			if hosts {
				parts[i].Hosts = chunk
			} else {
				parts[i].Children = chunk
			}
			if len(chunks) > 1 {
				parts[i].Part, parts[i].Parts = i+1, len(chunks)
			}
		}
		groupName, ok := groupNames[groupID]
		if !ok {
			groupName = fmt.Sprintf("#%d", groupID)
		}
		recordInventoryGroup(r, helpers.EventLogUpdate, verb, kind, inventory, db.InventoryGroup{ID: groupID, Name: groupName}, parts...)
	}
}

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

	if len(body.AddChildren) > 1000 || len(body.RemoveChildren) > 1000 {
		helpers.WriteErrorStatus(w, "at most 1000 child groups can be added or removed at once", http.StatusBadRequest)
		return
	}

	group := body.InventoryGroup
	group.ID = 0
	group.HostCount = 0
	group.InventoryID = inventory.ID
	group.ProjectID = inventory.ProjectID

	newGroup, err := helpers.Store(r).CreateInventoryGroup(group, body.AddChildren)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventoryGroup(r, helpers.EventLogCreate, "created", audit.ResourceInventoryGroupCreate, inventory, newGroup,
		audit.InventoryGroupMetadata{AddedVariables: diffVariables("", newGroup.Variables).AddedVariables})
	if len(newGroup.Children) > 0 {
		recordInventoryGroupLinks(r, inventory, audit.ResourceInventoryGroupChildAdd, "had child groups added",
			map[int]string{newGroup.ID: newGroup.Name}, map[int][]int{newGroup.ID: newGroup.Children}, inventoryGroupNames(r, inventory))
	}

	// The stored group carries the counted fields, the request may hold anything in them
	stored, err := helpers.Store(r).GetInventoryGroup(inventory.ProjectID, inventory.ID, newGroup.ID)
	if err != nil {
		// The group exists, an error would only make a retry fail on the name
		log.WithError(err).Error("Failed to read a created group")
		stored = newGroup
	}

	helpers.WriteJSON(w, http.StatusCreated, stored)
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

	// The old group is the base of the audit record
	oldGroup, err := helpers.Store(r).GetInventoryGroup(inventory.ProjectID, inventory.ID, groupID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var body inventoryGroupRequest
	if !helpers.Bind(w, r, &body) {
		return
	}

	if len(body.AddChildren) > 1000 || len(body.RemoveChildren) > 1000 {
		helpers.WriteErrorStatus(w, "at most 1000 child groups can be added or removed at once", http.StatusBadRequest)
		return
	}

	group := body.InventoryGroup
	group.ID = groupID
	group.InventoryID = inventory.ID
	group.ProjectID = inventory.ProjectID

	added, removed, err := helpers.Store(r).UpdateInventoryGroup(group, body.AddChildren, body.RemoveChildren)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var metadata audit.InventoryGroupMetadata
	if group.Name != oldGroup.Name {
		metadata.ChangedFields = append(metadata.ChangedFields, "name")
		metadata.OldName = oldGroup.Name
	}
	if group.Description != oldGroup.Description {
		metadata.ChangedFields = append(metadata.ChangedFields, "description")
	}
	diff := diffVariables(oldGroup.Variables, group.Variables)
	if len(diff.AddedVariables)+len(diff.RemovedVariables)+len(diff.ChangedVariables) > 0 {
		metadata.ChangedFields = append(metadata.ChangedFields, "variables")
	}
	metadata.AddedVariables = diff.AddedVariables
	metadata.RemovedVariables = diff.RemovedVariables
	metadata.ChangedVariables = diff.ChangedVariables

	if len(metadata.ChangedFields) > 0 {
		recordInventoryGroup(r, helpers.EventLogUpdate, "updated", audit.ResourceInventoryGroupUpdate, inventory, group, metadata)
	}
	if len(added)+len(removed) > 0 {
		groupNames := inventoryGroupNames(r, inventory)
		target := map[int]string{groupID: group.Name}
		recordInventoryGroupLinks(r, inventory, audit.ResourceInventoryGroupChildAdd, "had child groups added", target, map[int][]int{groupID: added}, groupNames)
		recordInventoryGroupLinks(r, inventory, audit.ResourceInventoryGroupChildRemove, "had child groups removed", target, map[int][]int{groupID: removed}, groupNames)
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

	// A failed read only leaves names and the count out of the record, as for a host
	groups, err := helpers.Store(r).GetInventoryGroups(inventory.ProjectID, inventory.ID)
	if err != nil {
		log.WithError(err).Error("Failed to read the groups of an inventory")
	}
	links, err := helpers.Store(r).GetInventoryGroupHosts(inventory.ProjectID, inventory.ID)
	if err != nil {
		log.WithError(err).Error("Failed to read the group hosts of an inventory")
	}

	group := db.InventoryGroup{ID: groupID}
	names := map[int]string{}
	var parents, children []string
	for _, g := range groups {
		names[g.ID] = g.Name
		if g.ID == groupID {
			group = g
		}
		// The delete also removes, by cascade, the links of the parents to the group
		if slices.Contains(g.Children, groupID) {
			parents = append(parents, g.Name)
		}
	}
	for _, id := range group.Children {
		if name, ok := names[id]; ok {
			children = append(children, name)
		}
	}
	hostCount := 0
	for _, link := range links {
		if link.GroupID == groupID {
			hostCount++
		}
	}

	if err := helpers.Store(r).DeleteInventoryGroup(inventory.ProjectID, inventory.ID, groupID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	chunks := auditNameChunks(inventory, slices.Concat(parents, children))
	parts := make([]audit.InventoryGroupMetadata, len(chunks))
	rest := len(parents)
	for i, chunk := range chunks {
		n := min(rest, len(chunk))
		parts[i].Parents, parts[i].Children = chunk[:n], chunk[n:]
		rest -= n
		if len(chunks) > 1 {
			parts[i].Part, parts[i].Parts = i+1, len(chunks)
		}
	}
	parts[0].HostCount = hostCount
	recordInventoryGroup(r, helpers.EventLogDelete, "deleted", audit.ResourceInventoryGroupDelete, inventory, group, parts...)

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
	if len(body.Add)+len(body.Remove) > 100 {
		helpers.WriteErrorStatus(w, "at most 100 groups can be changed at once", http.StatusBadRequest)
		return
	}

	added, removed, err := helpers.Store(r).ChangeInventoryHostGroups(inventory.ProjectID, inventory.ID, body.Hosts, body.Add, body.Remove)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if len(added)+len(removed) > 0 {
		// One read names every host, a host deleted meanwhile is recorded by its id
		hosts, _, err := helpers.Store(r).GetInventoryHosts(inventory.ProjectID, inventory.ID, db.InventoryHostListParams{IDs: body.Hosts, Count: len(body.Hosts)})
		if err != nil {
			log.WithError(err).Error("Failed to read the host names of an inventory")
		}
		hostNames := map[int]string{}
		for _, host := range hosts {
			hostNames[host.ID] = host.Name
		}
		groupNames := inventoryGroupNames(r, inventory)
		recordInventoryGroupLinks(r, inventory, audit.ResourceInventoryGroupHostAdd, "had hosts added", groupNames, added, hostNames)
		recordInventoryGroupLinks(r, inventory, audit.ResourceInventoryGroupHostRemove, "had hosts removed", groupNames, removed, hostNames)
	}

	w.WriteHeader(http.StatusNoContent)
}
