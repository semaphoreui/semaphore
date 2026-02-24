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
	for _, val := range e.values {
		old := val.value

		old.VaultKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.VaultKeyID, e)
		if err != nil {
			return err
		}

		old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID, e)
		if err != nil {
			return err
		}

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateTemplateVault(old)
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

func (e *TemplateVaultExporter) getName() string {
	return TemplateVault
}

func (e *TemplateVaultExporter) importDependsOn() []string {
	return []string{Template, AccessKey}
}

func (e *TemplateVaultExporter) exportDependsOn() []string {
	return []string{Template}
}
