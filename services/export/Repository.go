package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RepositoryExporter struct {
	ValueMap[db.Repository]
}

func (a *RepositoryExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		envs, err := store.GetRepositories(projId, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = a.appendValues(envs, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *RepositoryExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.SSHKeyID, err = exporter.getNewKeyInt(AccessKey, val.scope, old.SSHKeyID)
		if err != nil {
			return err
		}

		newVault, err := store.CreateRepository(old)
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

func (a *RepositoryExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *RepositoryExporter) importDependsOn() []string {
	return []string{AccessKey}
}

func (a *RepositoryExporter) getName() string {
	return Repository
}
