package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type EndpointAliasExporter struct {
	ValueMap[db.EndpointAlias]
}

func (e *EndpointAliasExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		endpoints, err := exporter.getLoadedKeysInt(Endpoint, strconv.Itoa(proj))
		if err != nil {
			return err
		}

		allValues := make([]db.EndpointAlias, 0)
		for _, endpoint := range endpoints {
			vals, err := store.GetEndpointAliases(proj, endpoint)
			if err != nil {
				return err
			}
			allValues = append(allValues, vals...)
		}

		err = e.appendValues(allValues, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EndpointAliasExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *EndpointAliasExporter) restoreValue(val EntityObject[db.EndpointAlias], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.EndpointID, err = exporter.getNewKeyInt(Endpoint, val.scope, old.EndpointID)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	alias, err := store.CreateEndpointAlias(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), alias.GetDbKey())
}

func (e *EndpointAliasExporter) getName() string {
	return EndpointAlias
}

func (e *EndpointAliasExporter) exportDependsOn() []string {
	return []string{Project, Endpoint}
}

func (e *EndpointAliasExporter) importDependsOn() []string {
	return []string{Project, Endpoint}
}
