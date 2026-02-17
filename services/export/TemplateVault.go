package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateVaultExporter struct {
	ValueMap[db.TemplateVault]
}

func (t *TemplateVaultExporter) load(store db.Store, exporter DataExporter) (err error) {

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

		err = t.appendValues(vaultsArr, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TemplateVaultExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range t.values {
		old := val.value

		old.VaultKeyID, err = exporter.getNewKeyIntRef(AccessKey, val.scope, old.VaultKeyID)
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

		newVault, err := store.CreateTemplateVault(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(t.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TemplateVaultExporter) getName() string {
	return TemplateVault
}

func (t *TemplateVaultExporter) importDependsOn() []string {
	return []string{Template, AccessKey}
}

func (t *TemplateVaultExporter) exportDependsOn() []string {
	return []string{Template}
}
