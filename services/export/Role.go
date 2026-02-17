package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RoleExporter struct {
	ValueMap[db.Role]
}

func (a *RoleExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		roles, err := store.GetProjectRoles(proj)
		if err != nil {
			return err
		}
		err = a.appendValues(roles, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	roles, err := store.GetGlobalRoles()
	if err != nil {
		return err
	}

	return a.appendValues(roles, GlobalScope)
}

func (a *RoleExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		newRole, err := store.CreateRole(old)
		if err != nil {
			return err
		}

		err = exporter.mapKeys(a.getName(), val.scope, old.Slug, newRole.Slug)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *RoleExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *RoleExporter) importDependsOn() []string {
	return []string{Project}
}

func (a *RoleExporter) getName() string {
	return Role
}
