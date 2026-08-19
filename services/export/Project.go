package export

import "github.com/semaphoreui/semaphore/db"

type ProjectExporter struct {
	ValueMap[db.Project]
}

func (e *ProjectExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	allKeys := make([]db.Project, 0)
	projects, err := store.GetAllProjects()
	if err != nil {
		return err
	}

	for _, proj := range projects {
		allKeys = append(allKeys, proj)
	}

	return e.appendValues(allKeys, GlobalScope)
}

func (e *ProjectExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *ProjectExporter) restoreValue(val EntityObject[db.Project], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	newObj, err := store.CreateProject(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *ProjectExporter) exportDependsOn() []string {
	return []string{}
}

func (e *ProjectExporter) getName() string {
	return Project
}
