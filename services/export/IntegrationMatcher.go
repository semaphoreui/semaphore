package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationMatcherExporter struct {
	ValueMap[db.IntegrationMatcher]
}

func (a *IntegrationMatcherExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {

		integrations, err := exporter.getLoadedKeysInt(Integration, strconv.Itoa(proj))
		if err != nil {
			return err
		}

		allValues := make([]db.IntegrationMatcher, 0)

		for _, integration := range integrations {
			vals, err := store.GetIntegrationMatchers(proj, db.RetrieveQueryParams{}, integration)
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

func (a *IntegrationMatcherExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.IntegrationID, err = exporter.getNewKeyInt(Integration, val.scope, old.IntegrationID)
		if err != nil {
			return err
		}

		newVault, err := store.CreateIntegrationMatcher(0, old)
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

func (a *IntegrationMatcherExporter) getName() string {
	return IntegrationMatcher
}

func (a *IntegrationMatcherExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (a *IntegrationMatcherExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
