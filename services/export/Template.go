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
	for _, val := range e.values {
		old := val.value

		old.Vaults = nil

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
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

		old.RepositoryID, err = exporter.getNewKeyInt(Repository, val.scope, old.RepositoryID, e)
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

		newTmpl, err := store.CreateTemplate(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(e.getName(), val.scope, old.ID, newTmpl.ID)
		if err != nil {
			return err
		}
	}
	return
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
