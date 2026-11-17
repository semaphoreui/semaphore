package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryHostExporter struct {
	ValueMap[db.InventoryHost]
}

func (e *InventoryHostExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		inventories, err := exporter.getLoadedKeysInt(Inventory, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		hostsArr := make([]db.InventoryHost, 0)

		for _, inventoryId := range inventories {
			hosts, _, err := store.GetInventoryHosts(projId, inventoryId, db.InventoryHostListParams{Count: 0})
			if err != nil {
				return err
			}
			hostsArr = append(hostsArr, hosts...)
		}

		err = e.appendValues(hostsArr, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *InventoryHostExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *InventoryHostExporter) restoreValue(val EntityObject[db.InventoryHost], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.InventoryID, err = exporter.getNewKeyInt(Inventory, val.scope, old.InventoryID)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	newObj, err := store.CreateInventoryHost(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *InventoryHostExporter) getName() string {
	return InventoryHost
}

func (e *InventoryHostExporter) importDependsOn() []string {
	return []string{Project, Inventory}
}

func (e *InventoryHostExporter) exportDependsOn() []string {
	return []string{Inventory}
}
