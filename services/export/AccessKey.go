package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type AccessKeyExporter struct {
	ValueMap[db.AccessKey]
}

func (e *AccessKeyExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		keys, err := store.GetAccessKeys(proj, db.GetAccessKeyOptions{IgnoreOwner: true}, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(keys, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *AccessKeyExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	for _, val := range e.values {
		old := val.value

		old.EnvironmentID, err = exporter.getNewKeyIntRef(Environment, val.scope, old.EnvironmentID, e)
		if err != nil {
			return err
		}

		old.StorageID, err = exporter.getNewKeyIntRef(SecretStorage, val.scope, old.StorageID, e)
		if err != nil {
			return err
		}

		old.UserID, err = exporter.getNewKeyIntRef(User, val.scope, old.UserID, e)
		if err != nil {
			return err
		}

		old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		old.SourceStorageID, err = exporter.getNewKeyIntRef(SecretStorage, val.scope, old.SourceStorageID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateAccessKey(old)
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

func (e *AccessKeyExporter) getName() string {
	return AccessKey
}

func (e *AccessKeyExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *AccessKeyExporter) importDependsOn() []string {
	return []string{Project, SecretStorage, Environment}
}
