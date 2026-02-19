package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RoleExporter struct {
	ValueMap[db.Role]
}

func (e *RoleExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		roles, err := store.GetProjectRoles(proj)
		if err != nil {
			return err
		}
		err = e.appendValues(roles, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	roles, err := store.GetGlobalRoles()
	if err != nil {
		return err
	}

	return e.appendValues(roles, GlobalScope)
}

func (e *RoleExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		newRole, err := store.CreateRole(old)
		if err != nil {
			return err
		}

		err = exporter.mapKeys(e.getName(), val.scope, old.Slug, newRole.Slug)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *RoleExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *RoleExporter) importDependsOn() []string {
	return []string{Project}
}

func (e *RoleExporter) getName() string {
	return Role
}
