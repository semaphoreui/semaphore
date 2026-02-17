package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateExporter struct {
	ValueMap[db.Template]
}

func (t *TemplateExporter) load(store db.Store, exporter DataExporter) (err error) {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		templates, err := store.GetTemplates(projId, db.TemplateFilter{}, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = t.appendValues(templates, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TemplateExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range t.values {
		old := val.value

		old.Vaults = nil

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.InventoryID)
		if err != nil {
			return err
		}

		old.EnvironmentID, err = exporter.getNewKeyIntRef(Environment, val.scope, old.EnvironmentID)
		if err != nil {
			return err
		}

		old.RepositoryID, err = exporter.getNewKeyInt(Repository, val.scope, old.RepositoryID)
		if err != nil {
			return err
		}

		newTmpl, err := store.CreateTemplate(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(t.getName(), val.scope, old.ID, newTmpl.ID)
		if err != nil {
			return err
		}
	}
	return
}

func (t *TemplateExporter) getName() string {
	return Template
}

func (t *TemplateExporter) exportDependsOn() []string {
	return []string{Project}
}

func (t *TemplateExporter) importDependsOn() []string {
	return []string{Project, Inventory, Environment, Repository, View}
}
