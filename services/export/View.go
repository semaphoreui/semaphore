package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ViewExporter struct {
	ValueMap[db.View]
}

func (e *ViewExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {

		envs, err := store.GetViews(proj)
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

func (e *ViewExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *ViewExporter) restoreValue(val EntityObject[db.View], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	newObj, err := store.CreateView(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *ViewExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *ViewExporter) importDependsOn() []string {
	return []string{Project}
}

func (e *ViewExporter) getName() string {
	return View
}
