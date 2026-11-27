package export

import (
	"github.com/semaphoreui/semaphore/db"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
)

type roleMapping struct {
	ID        int    `json:"id"`
	Provider  string `json:"provider"`
	Group     string `json:"group"`
	ProjectID int    `json:"project_id"`
	RoleID    int    `json:"role_id"`
}

func (m roleMapping) GetDbKey() EntityKey {
	return NewKeyFromInt(m.ID)
}

// RoleMappingExporter copies the mappings, not the mapped roles: every user signs in again after a copy.
type RoleMappingExporter struct {
	ValueMap[roleMapping]
}

func (e *RoleMappingExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	mappings, err := proFactory.NewRoleMappingStore(store).GetRoleMappings()
	if err != nil {
		return err
	}

	values := make([]roleMapping, 0, len(mappings))
	for _, m := range mappings {
		values = append(values, roleMapping{ID: m.ID, Provider: m.Provider, Group: m.Group, ProjectID: m.ProjectID, RoleID: m.RoleID})
	}

	return e.appendValues(values, GlobalScope)
}

func (e *RoleMappingExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *RoleMappingExporter) restoreValue(val EntityObject[roleMapping], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	projectID, err := exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	roleID, err := exporter.getNewKeyInt(Role, GlobalScope, old.RoleID)
	if err != nil {
		return err
	}

	newObj, err := proFactory.NewRoleMappingStore(store).CreateRoleMapping(db.RoleMapping{
		Provider:  old.Provider,
		Group:     old.Group,
		ProjectID: projectID,
		RoleID:    roleID,
	})
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), NewKeyFromInt(newObj.ID))
}

func (e *RoleMappingExporter) exportDependsOn() []string {
	return []string{Project, Role}
}

func (e *RoleMappingExporter) importDependsOn() []string {
	return []string{Project, Role}
}

func (e *RoleMappingExporter) getName() string {
	return RoleMapping
}
