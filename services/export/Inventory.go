package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryExporter struct {
	ValueMap[db.Inventory]
}

func (a *InventoryExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		envs, err := store.GetInventories(proj, db.RetrieveQueryParams{}, []db.InventoryType{})
		if err != nil {
			return err
		}
		err = a.appendValues(envs, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *InventoryExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.SSHKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.SSHKeyID)
		if err != nil {
			return err
		}

		old.BecomeKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.BecomeKeyID)
		if err != nil {
			return err
		}

		old.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID)
		if err != nil {
			return err
		}

		//templateId, err := exporter.getKeyMapForType(Template, *old.BecomeKeyID)
		//if err != nil {
		//	return err
		//}
		//old.TemplateID = &templateId

		newVault, err := store.CreateInventory(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *InventoryExporter) getName() string {
	return Inventory
}

func (a *InventoryExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *InventoryExporter) importDependsOn() []string {
	return []string{AccessKey, Repository}
}
