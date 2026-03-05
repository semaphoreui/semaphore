package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateVaultExporter struct {
	ValueMap[db.TemplateVault]
}

func (e *TemplateVaultExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		templates, err := exporter.getLoadedKeysInt(Template, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		vaultsArr := make([]db.TemplateVault, 0)

		for key := range templates {

			vaults, err := store.GetTemplateVaults(projId, key)
			if err != nil {
				return err
			}
			vaultsArr = append(vaultsArr, vaults...)
		}

		err = e.appendValues(vaultsArr, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *TemplateVaultExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *TemplateVaultExporter) restoreValue(val EntityObject[db.TemplateVault], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.VaultKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.VaultKeyID, e)
	if err != nil {
		return err
	}

	old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID)
	if err != nil {
		return err
	}

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	newObj, err := store.CreateTemplateVault(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *TemplateVaultExporter) getName() string {
	return TemplateVault
}

func (e *TemplateVaultExporter) importDependsOn() []string {
	return []string{Project, Template, AccessKey}
}

func (e *TemplateVaultExporter) exportDependsOn() []string {
	return []string{Template}
}
