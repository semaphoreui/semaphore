package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationExtractValueExporter struct {
	ValueMap[db.IntegrationExtractValue]
}

func (e *IntegrationExtractValueExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {

		integrations, err := exporter.getLoadedKeysInt(Integration, strconv.Itoa(proj))
		if err != nil {
			return err
		}
		allValues := make([]db.IntegrationExtractValue, 0)

		for _, integration := range integrations {
			vals, err := store.GetIntegrationExtractValues(proj, db.RetrieveQueryParams{}, integration)
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

func (e *IntegrationExtractValueExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *IntegrationExtractValueExporter) restoreValue(val EntityObject[db.IntegrationExtractValue], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.IntegrationID, err = exporter.getNewKeyInt(Integration, val.scope, old.IntegrationID)
	if err != nil {
		return err
	}

	newVault, err := store.CreateIntegrationExtractValue(0, old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newVault.GetDbKey())
}

func (e *IntegrationExtractValueExporter) getName() string {
	return IntegrationExtractValue
}

func (e *IntegrationExtractValueExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (e *IntegrationExtractValueExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
