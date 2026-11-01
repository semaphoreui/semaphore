package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationGroupExporter struct {
	ValueMap[db.IntegrationGroup]
}

func (e *IntegrationGroupExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		groups, err := store.GetIntegrationGroups(proj, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(groups, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *IntegrationGroupExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *IntegrationGroupExporter) restoreValue(val EntityObject[db.IntegrationGroup], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.EndpointID, err = exporter.getNewKeyInt(Endpoint, val.scope, old.EndpointID)
	if err != nil {
		return err
	}

	old.AuthSecretID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.AuthSecretID, e)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	group, err := store.CreateIntegrationGroup(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), group.GetDbKey())
}

func (e *IntegrationGroupExporter) getName() string {
	return IntegrationGroup
}

func (e *IntegrationGroupExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *IntegrationGroupExporter) importDependsOn() []string {
	return []string{Project, Endpoint, AccessKey}
}
