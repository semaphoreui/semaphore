package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type EnvironmentExporter struct {
	ValueMap[db.Environment]
}

func (e *EnvironmentExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		envs, err := store.GetEnvironments(proj, db.RetrieveQueryParams{})
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

func (e *EnvironmentExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		old.SecretStorageID, err = exporter.getNewKeyIntRef(SecretStorage, val.scope, old.SecretStorageID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateEnvironment(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(e.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EnvironmentExporter) getName() string {
	return Environment
}

func (e *EnvironmentExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *EnvironmentExporter) importDependsOn() []string {
	return []string{SecretStorage}
}
