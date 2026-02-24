package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type SecretStorageExporter struct {
	ValueMap[db.SecretStorage]
}

func (e *SecretStorageExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		keys, err := store.GetSecretStorages(projId)
		if err != nil {
			return err
		}

		err = e.appendValues(keys, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *SecretStorageExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value
		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)

		newVault, err := store.CreateSecretStorage(old)
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

func (e *SecretStorageExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *SecretStorageExporter) importDependsOn() []string {
	return []string{Project}
}

func (e *SecretStorageExporter) getName() string {
	return SecretStorage
}
