package export

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryGroupExporter struct {
	ValueMap[db.InventoryGroup]
	// hosts holds the old host ids of each old group id
	hosts map[int][]int
}

func (e *InventoryGroupExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	e.hosts = make(map[int][]int)

	for _, projId := range projs {
		inventories, err := exporter.getLoadedKeysInt(Inventory, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		groupsArr := make([]db.InventoryGroup, 0)

		for _, inventoryId := range inventories {
			groups, err := store.GetInventoryGroups(projId, inventoryId)
			if err != nil {
				return err
			}
			groupsArr = append(groupsArr, groups...)

			links, err := store.GetInventoryGroupHosts(projId, inventoryId)
			if err != nil {
				return err
			}
			for _, link := range links {
				e.hosts[link.GroupID] = append(e.hosts[link.GroupID], link.HostID)
			}
		}

		err = e.appendValues(groupsArr, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *InventoryGroupExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	err = e.restoreValues(store, exporter, progress, e)
	if err != nil {
		return err
	}

	// Links are set once every group exists, a child may come after its parent
	for _, val := range e.values {
		group := val.value

		// A group that was not created has its error reported already
		group.ID, err = exporter.getNewKeyInt(InventoryGroup, val.scope, val.value.ID)
		if err != nil {
			continue
		}
		group.InventoryID, err = exporter.getNewKeyInt(Inventory, val.scope, val.value.InventoryID)
		if err != nil {
			continue
		}
		group.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, val.value.ProjectID)
		if err != nil {
			continue
		}

		var children []int
		for _, oldChild := range val.value.Children {
			child, err := exporter.getNewKeyInt(InventoryGroup, val.scope, oldChild)
			if err != nil {
				e.onError(fmt.Sprintf("Unable to restore %s links: %s", e.getName(), err.Error()))
				continue
			}
			children = append(children, child)
		}
		if len(children) > 0 {
			_, _, err = store.UpdateInventoryGroup(group, children, nil)
			if err != nil {
				e.onError(fmt.Sprintf("Unable to restore %s links: %s", e.getName(), err.Error()))
			}
		}

		var hosts []int
		for _, oldHost := range e.hosts[val.value.ID] {
			host, err := exporter.getNewKeyInt(InventoryHost, val.scope, oldHost)
			if err != nil {
				e.onError(fmt.Sprintf("Unable to restore %s links: %s", e.getName(), err.Error()))
				continue
			}
			hosts = append(hosts, host)
		}
		// Batches keep the host id list of one call within database parameter limits
		for batch := range slices.Chunk(hosts, 1000) {
			_, _, err = store.ChangeInventoryHostGroups(group.ProjectID, group.InventoryID, batch, []int{group.ID}, nil)
			if err != nil {
				e.onError(fmt.Sprintf("Unable to restore %s links: %s", e.getName(), err.Error()))
			}
		}
	}

	return nil
}

func (e *InventoryGroupExporter) restoreValue(val EntityObject[db.InventoryGroup], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.InventoryID, err = exporter.getNewKeyInt(Inventory, val.scope, old.InventoryID)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	newObj, err := store.CreateInventoryGroup(old, nil)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *InventoryGroupExporter) getName() string {
	return InventoryGroup
}

func (e *InventoryGroupExporter) importDependsOn() []string {
	return []string{Project, Inventory, InventoryHost}
}

func (e *InventoryGroupExporter) exportDependsOn() []string {
	return []string{Inventory}
}

func (e *InventoryGroupExporter) clear() {
	e.ValueMap.clear()
	e.hosts = nil
}
