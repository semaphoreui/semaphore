package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type EndpointExporter struct {
	ValueMap[db.Endpoint]
}

func (e *EndpointExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		endpoints, err := store.GetEndpoints(proj, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(endpoints, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EndpointExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *EndpointExporter) restoreValue(val EntityObject[db.Endpoint], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.AuthSecretID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.AuthSecretID, e)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	endpoint, err := store.CreateEndpoint(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), endpoint.GetDbKey())
}

func (e *EndpointExporter) getName() string {
	return Endpoint
}

func (e *EndpointExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *EndpointExporter) importDependsOn() []string {
	return []string{Project, AccessKey}
}
