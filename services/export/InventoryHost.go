package export

import (
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryHostExporter struct {
	ValueMap[db.InventoryHost]
	contents *inventoryContents
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
			content, err := e.contents.get(store, projId, inventoryId)
			if err != nil {
				return err
			}
			hostsArr = append(hostsArr, content.Hosts...)
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

	sourceID := old.SourceID
	old.SourceID, err = exporter.getNewKeyIntRef(InventorySource, val.scope, old.SourceID, e)
	if err != nil {
		return err
	}
	// Without its source the row would come back as a manual one
	if sourceID != nil && old.SourceID == nil {
		return fmt.Errorf("the source of host %s is not restored", old.Name)
	}

	newObj, err := store.CreateInventoryHost(old, nil)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *InventoryHostExporter) getName() string {
	return InventoryHost
}

func (e *InventoryHostExporter) importDependsOn() []string {
	return []string{Project, Inventory, InventorySource}
}

func (e *InventoryHostExporter) exportDependsOn() []string {
	return []string{Inventory}
}
