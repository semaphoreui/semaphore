package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type HostConfigExporter struct {
	ValueMap[db.HostConfig]
}

func (e *HostConfigExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		hostConfigs, err := store.GetHostConfigs(projId, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValues(hostConfigs, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *HostConfigExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *HostConfigExporter) restoreValue(val EntityObject[db.HostConfig], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	old.SSHKeyID, err = exporter.getNewKeyInt(AccessKey, val.scope, old.SSHKeyID)
	if err != nil {
		return err
	}

	newObj, err := store.CreateHostConfig(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *HostConfigExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *HostConfigExporter) importDependsOn() []string {
	return []string{Project, AccessKey}
}

func (e *HostConfigExporter) getName() string {
	return HostConfig
}
