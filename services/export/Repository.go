package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RepositoryExporter struct {
	ValueMap[db.Repository]
}

func (e *RepositoryExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		envs, err := store.GetRepositories(projId, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(envs, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *RepositoryExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		old.SSHKeyID, err = exporter.getNewKeyInt(AccessKey, val.scope, old.SSHKeyID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateRepository(old)
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

func (e *RepositoryExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *RepositoryExporter) importDependsOn() []string {
	return []string{AccessKey}
}

func (e *RepositoryExporter) getName() string {
	return Repository
}
