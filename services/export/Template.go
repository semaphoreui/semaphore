package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateExporter struct {
	ValueMap[db.Template]
}

func (e *TemplateExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		templates, err := store.GetTemplates(projId, db.TemplateFilter{}, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(templates, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *TemplateExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *TemplateExporter) restoreValue(val EntityObject[db.Template], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.Vaults = nil

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	old.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.InventoryID, e)
	if err != nil {
		return err
	}

	old.EnvironmentID, err = exporter.getNewKeyIntRef(Environment, val.scope, old.EnvironmentID, e)
	if err != nil {
		return err
	}

	old.RepositoryID, err = exporter.getNewKeyInt(Repository, val.scope, old.RepositoryID)
	if err != nil {
		return err
	}

	old.ViewID, err = exporter.getNewKeyIntRef(View, val.scope, old.ViewID, e)
	if err != nil {
		return err
	}

	old.BuildTemplateID, err = exporter.getNewKeyIntRef(Template, val.scope, old.BuildTemplateID, e)
	if err != nil {
		return err
	}

	newObj, err := store.CreateTemplate(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *TemplateExporter) getName() string {
	return Template
}

func (e *TemplateExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *TemplateExporter) importDependsOn() []string {
	return []string{Project, Inventory, Environment, Repository, View}
}
