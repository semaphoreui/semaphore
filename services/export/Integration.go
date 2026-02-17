package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationExporter struct {
	ValueMap[db.Integration]
}

func (a *IntegrationExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		keys, err := store.GetIntegrations(proj, db.RetrieveQueryParams{}, true)
		if err != nil {
			return err
		}
		err = a.appendValues(keys, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *IntegrationExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		if old.TaskParams != nil {
			old.TaskParams.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.TaskParams.InventoryID)
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

		old.AuthSecretID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.AuthSecretID)
		if err != nil {
			return err
		}

		//old.TaskParamsID, err = exporter.getNewKeyIntRef(TaskParams, val.scope, old.TaskParamsID)
		//if err != nil {
		//	return err
		//}

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		integration, err := store.CreateIntegration(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, integration.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *IntegrationExporter) getName() string {
	return Integration
}

func (a *IntegrationExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *IntegrationExporter) importDependsOn() []string {
	return []string{Project, SecretStorage, Environment, Template}
}
