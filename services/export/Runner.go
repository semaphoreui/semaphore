package export

import (
	"github.com/semaphoreui/semaphore/db"
)

type RunnerExporter struct {
	ValueMap[db.Runner]
}

func (e *RunnerExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	envs, err := store.GetAllRunners(false, false)
	if err != nil {
		return err
	}

	err = e.appendValues(envs, GlobalScope)
	if err != nil {
		return err
	}
	return nil
}

func (e *RunnerExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *RunnerExporter) restoreValue(val EntityObject[db.Runner], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
	if err != nil {
		return err
	}

	newObj, err := store.CreateRunner(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
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
