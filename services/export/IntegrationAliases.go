package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationAliasExporter struct {
	ValueMap[db.IntegrationAlias]
}

func (e *IntegrationAliasExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		vals, err := store.GetIntegrationAliases(proj, nil)
		if err != nil {
			return err
		}

		allValues := make([]db.IntegrationAlias, 0)
		allValues = append(allValues, vals...)

		integrations, err := exporter.getLoadedKeysInt(Integration, strconv.Itoa(proj))
		if err != nil {
			return err
		}

		for _, integration := range integrations {
			vals, err = store.GetIntegrationAliases(proj, &integration)
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

func (e *IntegrationAliasExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *IntegrationAliasExporter) restoreValue(val EntityObject[db.IntegrationAlias], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.IntegrationID, err = exporter.getNewKeyIntRef(Integration, val.scope, old.IntegrationID, e)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	newVault, err := store.CreateIntegrationAlias(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newVault.GetDbKey())

}

func (e *IntegrationAliasExporter) getName() string {
	return IntegrationAlias
}

func (e *IntegrationAliasExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (e *IntegrationAliasExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
