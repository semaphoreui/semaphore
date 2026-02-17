package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ViewExporter struct {
	ValueMap[db.View]
}

func (a *ViewExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {

		envs, err := store.GetViews(proj)
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

func (a *ViewExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		newView, err := store.CreateView(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, newView.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ViewExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *ViewExporter) importDependsOn() []string {
	return []string{Project}
}

func (a *ViewExporter) getName() string {
	return View
}
