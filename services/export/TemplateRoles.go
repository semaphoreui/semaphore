package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateRoleExporter struct {
	ValueMap[db.TemplateRolePerm]
}

func (t *TemplateRoleExporter) load(store db.Store, exporter DataExporter) (err error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		templates, err := exporter.getLoadedKeysInt(Template, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		roles := make([]db.TemplateRolePerm, 0)

		for key := range templates {
			templateRoles, err := store.GetTemplateRoles(projId, key)
			if err != nil {
				return err
			}
			roles = append(roles, templateRoles...)
		}

		err = t.appendValues(roles, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TemplateRoleExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range t.values {
		old := val.value

		old.RoleSlug, err = exporter.getNewKey(Role, val.scope, old.RoleSlug)
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

		newVault, err := store.CreateTemplateRole(old)
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

func (t *TemplateRoleExporter) getName() string {
	return TemplateRole
}

func (t *TemplateRoleExporter) importDependsOn() []string {
	return []string{Template, Project}
}

func (t *TemplateRoleExporter) exportDependsOn() []string {
	return []string{Template, Project}
}
