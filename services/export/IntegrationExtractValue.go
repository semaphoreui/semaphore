package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationExtractValueExporter struct {
	ValueMap[db.IntegrationExtractValue]
}

func (a *IntegrationExtractValueExporter) load(store db.Store, exporter DataExporter) error {

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

		err = a.appendValues(allValues, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *IntegrationExtractValueExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.IntegrationID, err = exporter.getNewKeyInt(Integration, val.scope, old.IntegrationID)
		if err != nil {
			return err
		}
		// TODO projectId?
		newVault, err := store.CreateIntegrationExtractValue(0, old)
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

func (a *IntegrationExtractValueExporter) getName() string {
	return IntegrationExtractValue
}

func (a *IntegrationExtractValueExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (a *IntegrationExtractValueExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
