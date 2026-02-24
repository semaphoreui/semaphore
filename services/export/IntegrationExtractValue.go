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

	for _, val := range e.values {
		old := val.value

		old.IntegrationID, err = exporter.getNewKeyInt(Integration, val.scope, old.IntegrationID, e)
		if err != nil {
			return err
		}
		// TODO projectId?
		newVault, err := store.CreateIntegrationExtractValue(0, old)
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

func (e *IntegrationExtractValueExporter) getName() string {
	return IntegrationExtractValue
}

func (e *IntegrationExtractValueExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (e *IntegrationExtractValueExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
