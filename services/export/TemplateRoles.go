package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TemplateRoleExporter struct {
	ValueMap[db.TemplateRolePerm]
}

func (e *TemplateRoleExporter) load(store db.Store, exporter DataExporter, progress Progress) (err error) {

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

		err = e.appendValues(roles, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *TemplateRoleExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	for _, val := range e.values {
		old := val.value

		old.RoleSlug, err = exporter.getNewKey(Role, val.scope, old.RoleSlug, e)
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

		newVault, err := store.CreateTemplateRole(old)
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

func (e *TemplateRoleExporter) getName() string {
	return TemplateRole
}

func (e *TemplateRoleExporter) importDependsOn() []string {
	return []string{Template, Project}
}

func (e *TemplateRoleExporter) exportDependsOn() []string {
	return []string{Template, Project}
}
