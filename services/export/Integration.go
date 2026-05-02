package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationExporter struct {
	ValueMap[db.Integration]
}

func (e *IntegrationExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		keys, err := store.GetIntegrations(proj, db.RetrieveQueryParams{}, true)
		if err != nil {
			return err
		}
		err = e.appendValues(keys, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *IntegrationExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *IntegrationExporter) restoreValue(val EntityObject[db.Integration], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	if old.TaskParams != nil {
		old.TaskParams.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.TaskParams.InventoryID, e)
		if err != nil {
			return err
		}

		old.TaskParams.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}
	}

	old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID)
	if err != nil {
		return err
	}

	old.AuthSecretID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.AuthSecretID, e)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	integration, err := store.CreateIntegration(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), integration.GetDbKey())

}

func (e *IntegrationExporter) getName() string {
	return Integration
}

func (e *IntegrationExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *IntegrationExporter) importDependsOn() []string {
	return []string{Project, SecretStorage, Template, Inventory, AccessKey}
}
