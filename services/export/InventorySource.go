package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

// inventoryContents reads each inventory once, so its hosts, groups and sources come from one state
type inventoryContents struct {
	byInventory map[int]db.InventoryContent
}

func (c *inventoryContents) get(store db.Store, projectID, inventoryID int) (db.InventoryContent, error) {
	if content, ok := c.byInventory[inventoryID]; ok {
		return content, nil
	}

	content, err := store.GetInventoryContent(projectID, inventoryID)
	if err != nil {
		return db.InventoryContent{}, err
	}
	c.byInventory[inventoryID] = content
	return content, nil
}

type InventorySourceExporter struct {
	ValueMap[db.InventorySource]
	contents *inventoryContents
}

func (e *InventorySourceExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		inventories, err := exporter.getLoadedKeysInt(Inventory, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		sourcesArr := make([]db.InventorySource, 0)

		for _, inventoryId := range inventories {
			content, err := e.contents.get(store, projId, inventoryId)
			if err != nil {
				return err
			}
			// The sources come by id, the restore numbers them again in this order and a task prefers the lower number
			sourcesArr = append(sourcesArr, content.Sources...)
		}

		err = e.appendValues(sourcesArr, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *InventorySourceExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *InventorySourceExporter) restoreValue(val EntityObject[db.InventorySource], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	// The sync state stays behind, the next sync fills it
	source := db.InventorySource{
		Name:          old.Name,
		Type:          old.Type,
		Params:        old.Params,
		Variables:     old.Variables,
		VariableCount: old.VariableCount,
	}

	source.InventoryID, err = exporter.getNewKeyInt(Inventory, val.scope, old.InventoryID)
	if err != nil {
		return err
	}

	source.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	source.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID, e)
	if err != nil {
		return err
	}

	source.KeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.KeyID, e)
	if err != nil {
		return err
	}

	newObj, err := store.CreateInventorySource(source)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *InventorySourceExporter) getName() string {
	return InventorySource
}

func (e *InventorySourceExporter) importDependsOn() []string {
	return []string{Project, Inventory, Repository}
}

func (e *InventorySourceExporter) exportDependsOn() []string {
	return []string{Inventory}
}
