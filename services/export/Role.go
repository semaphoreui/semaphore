package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type RoleExporter struct {
	ValueMap[db.Role]
}

func getNewRoleID(exporter DataExporter, projectID int, roleID int) (int, error) {
	newRoleID, err := exporter.getNewKeyInt(Role, strconv.Itoa(projectID), roleID)
	if err == nil {
		return newRoleID, nil
	}
	return exporter.getNewKeyInt(Role, GlobalScope, roleID)
}

func (e *RoleExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		roles, err := store.GetRoles(db.ProjectRolesQuery{ProjectID: proj})
		if err != nil {
			return err
		}
		err = e.appendValues(roles, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	roles, err := store.GetRoles(db.GlobalRolesQuery{Kinds: db.RoleKindAll})
	if err != nil {
		return err
	}

	return e.appendValues(roles, GlobalScope)
}

func (e *RoleExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *RoleExporter) restoreValue(val EntityObject[db.Role], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	if old.IsBuiltin() {
		destRole, err := store.GetRole(db.BuiltinRoleQuery{Key: *old.BuiltinKey})
		if err != nil {
			return err
		}
		return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), destRole.GetDbKey())
	}

	old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
	if err != nil {
		return err
	}

	destRole, err := store.CreateRole(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), destRole.GetDbKey())
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
