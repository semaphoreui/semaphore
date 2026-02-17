package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type SecretStorageExporter struct {
	ValueMap[db.SecretStorage]
}

func (a *SecretStorageExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		keys, err := store.GetSecretStorages(projId)
		if err != nil {
			return err
		}

		err = a.appendValues(keys, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *SecretStorageExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value
		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)

		newVault, err := store.CreateSecretStorage(old)
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

func (a *SecretStorageExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *SecretStorageExporter) importDependsOn() []string {
	return []string{Project}
}

func (a *SecretStorageExporter) getName() string {
	return SecretStorage
}
