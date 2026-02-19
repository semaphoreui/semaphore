package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RunnerExporter struct {
	ValueMap[db.Runner]
}

func (e *RunnerExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {

		envs, err := store.GetRunners(proj, false, nil)
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

func (e *RunnerExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		newView, err := store.CreateRunner(old)
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

func (e *RunnerExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *RunnerExporter) importDependsOn() []string {
	return []string{Project}
}

func (e *RunnerExporter) getName() string {
	return Runner
}
