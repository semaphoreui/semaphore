package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryExporter struct {
	ValueMap[db.Inventory]
}

func (e *InventoryExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		envs, err := store.GetInventories(proj, db.RetrieveQueryParams{}, []db.InventoryType{})
		if err != nil {
			return err
		}
		err = e.appendValues(envs, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *InventoryExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *InventoryExporter) restoreValue(val EntityObject[db.Inventory], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	old.SSHKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.SSHKeyID, e)
	if err != nil {
		return err
	}

	old.BecomeKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.BecomeKeyID, e)
	if err != nil {
		return err
	}

	old.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID, e)
	if err != nil {
		return err
	}

	//templateId, err := exporter.getKeyMapForType(Template, *old.BecomeKeyID)
	//if err != nil {
	//	return err
	//}
	//old.TemplateID = &templateId

	newObj, err := store.CreateInventory(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *InventoryExporter) getName() string {
	return Inventory
}

func (e *InventoryExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *InventoryExporter) importDependsOn() []string {
	return []string{Project, AccessKey, Repository}
}
