package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type EnvironmentExporter struct {
	ValueMap[db.Environment]
}

func (a *EnvironmentExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		envs, err := store.GetEnvironments(proj, db.RetrieveQueryParams{})
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

func (a *EnvironmentExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.SecretStorageID, err = exporter.getNewKeyIntRef(SecretStorage, val.scope, old.SecretStorageID)
		if err != nil {
			return err
		}

		newVault, err := store.CreateEnvironment(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *EnvironmentExporter) getName() string {
	return Environment
}

func (a *EnvironmentExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *EnvironmentExporter) importDependsOn() []string {
	return []string{SecretStorage}
}
