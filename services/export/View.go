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

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		newView, err := store.CreateView(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(e.getName(), val.scope, old.ID, newView.ID)
		if err != nil {
			return err
		}
	}

	return nil
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
