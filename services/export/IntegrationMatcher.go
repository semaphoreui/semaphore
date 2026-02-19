package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationMatcherExporter struct {
	ValueMap[db.IntegrationMatcher]
}

func (e *IntegrationMatcherExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

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
		err = e.appendValues(allValues, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *IntegrationMatcherExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.IntegrationID, err = exporter.getNewKeyInt(Integration, val.scope, old.IntegrationID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateIntegrationMatcher(0, old)
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

func (e *IntegrationMatcherExporter) getName() string {
	return IntegrationMatcher
}

func (e *IntegrationMatcherExporter) exportDependsOn() []string {
	return []string{Project, Integration}
}

func (e *IntegrationMatcherExporter) importDependsOn() []string {
	return []string{Project, Integration}
}
