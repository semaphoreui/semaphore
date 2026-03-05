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
	return e.restoreValues(store, exporter, progress, e)
}

func (e *TemplateRoleExporter) restoreValue(val EntityObject[db.TemplateRolePerm], store db.Store, exporter DataExporter) (err error) {
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

	newObj, err := store.CreateTemplateRole(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *TemplateRoleExporter) getName() string {
	return TemplateRole
}

func (e *TemplateRoleExporter) importDependsOn() []string {
	return []string{Role, Template, Project}
}

func (e *TemplateRoleExporter) exportDependsOn() []string {
	return []string{Template, Project}
}
