package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type AlertExporter struct {
	ValueMap[db.Alert]
}

func (e *AlertExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		alerts, err := store.GetAlerts(proj, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(alerts, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *AlertExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *AlertExporter) restoreValue(val EntityObject[db.Alert], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	old.KeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.KeyID, e)
	if err != nil {
		return err
	}

	newObj, err := store.CreateAlert(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *AlertExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *AlertExporter) importDependsOn() []string {
	return []string{Project, AccessKey}
}

func (e *AlertExporter) getName() string {
	return Alert
}

func mapAlertIDs(mapper KeyMapper, scope string, ids []int) ([]int, error) {
	return mapAlertIDsOpt(mapper, scope, ids, false)
}

func mapAlertIDsOpt(mapper KeyMapper, scope string, ids []int, skipMissing bool) ([]int, error) {
	if len(ids) == 0 {
		return ids, nil
	}

	mapped := make([]int, 0, len(ids))
	for _, id := range ids {
		newID, err := mapper.getNewKeyInt(Alert, scope, id)
		if err != nil {
			if skipMissing {
				continue
			}
			return nil, err
		}
		mapped = append(mapped, newID)
	}
	return mapped, nil
}
